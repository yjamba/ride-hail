package postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	config    *DBConfig
	pool      *pgxpool.Pool
	mu        sync.RWMutex
	TxManager *TxManager
}
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type Tx struct {
	tx pgx.Tx
}

type failedRow struct {
	err error
}

func (fr *failedRow) Scan(dest ...interface{}) error {
	return fr.err
}

func NewDB(config *DBConfig) *Database {
	db := &Database{
		config: config,
	}
	db.TxManager = NewTxManager(db)
	return db
}

func (db *Database) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		db.config.User,
		db.config.Password,
		db.config.Host,
		db.config.Port,
		db.config.DBName,
		db.config.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Конфигурация пула соединений для надёжности
	poolConfig.MaxConns = 25                       // Максимум соединений
	poolConfig.MinConns = 5                        // Минимум соединений
	poolConfig.MaxConnLifetime = 15 * time.Minute  // Время жизни соединения
	poolConfig.MaxConnIdleTime = 5 * time.Minute   // Время простоя перед закрытием
	poolConfig.HealthCheckPeriod = 1 * time.Minute // Период проверки здоровья

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	db.mu.Lock()
	db.pool = pool
	db.mu.Unlock()

	return nil
}

func (db *Database) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.pool != nil {
		db.pool.Close()
		db.pool = nil
	}
	return nil
}

func (db *Database) GetPool() *pgxpool.Pool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.pool
}

// IsHealthy проверяет живо ли соединение с БД
func (db *Database) IsHealthy(ctx context.Context) bool {
	db.mu.RLock()
	pool := db.pool
	db.mu.RUnlock()

	if pool == nil {
		return false
	}

	return pool.Ping(ctx) == nil
}

// Reconnect переподключается к базе данных
func (db *Database) Reconnect(ctx context.Context) error {
	db.mu.Lock()
	if db.pool != nil {
		db.pool.Close()
		db.pool = nil
	}
	db.mu.Unlock()

	return db.Connect(ctx)
}

// Querier implementation
func (db *Database) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if db.pool == nil {
		// Вернуть фиктивный Row с ошибкой
		return &failedRow{err: fmt.Errorf("database pool is not initialized")}
	}
	return db.pool.QueryRow(ctx, sql, args...)
}

func (db *Database) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if db.pool == nil {
		return nil, fmt.Errorf("database pool is not initialized")
	}
	return db.pool.Query(ctx, sql, args...)
}

func (db *Database) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if db.pool == nil {
		return pgconn.CommandTag{}, fmt.Errorf("database pool is not initialized")
	}
	return db.pool.Exec(ctx, sql, arguments...)
}

// Transaction support
func (db *Database) BeginTx(ctx context.Context) (*Tx, error) {
	db.mu.RLock()
	pool := db.pool
	db.mu.RUnlock()

	if pool == nil {
		return nil, fmt.Errorf("database pool is not initialized")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &Tx{tx: tx}, nil
}

// Tx Querier implementation
func (t *Tx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return t.tx.QueryRow(ctx, sql, args...)
}

func (t *Tx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return t.tx.Query(ctx, sql, args...)
}

func (t *Tx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return t.tx.Exec(ctx, sql, arguments...)
}

func (t *Tx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *Tx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// WithRetry выполняет функцию с автоматическими повторами при ошибках подключения
func (db *Database) WithRetry(ctx context.Context, maxRetries int, fn func(context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Проверяем здоровье соединения
			if !db.IsHealthy(ctx) {
				// Пытаемся переподключиться
				if err := db.Reconnect(ctx); err != nil {
					lastErr = fmt.Errorf("reconnect failed (attempt %d): %w", attempt, err)
					continue
				}
			}
		}

		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}

		// Проверяем, является ли ошибка ошибкой соединения
		if !isConnectionError(lastErr) {
			return lastErr
		}

		if attempt < maxRetries-1 {
			// Экспоненциальный backoff: 100ms, 200ms, 400ms...
			backoff := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// isConnectionError определяет является ли ошибка ошибкой соединения
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	// Типичные ошибки соединения
	return errMsg == "database pool is not initialized" ||
		errMsg == "connection refused" ||
		errMsg == "connection reset by peer" ||
		errMsg == "i/o timeout"
}
