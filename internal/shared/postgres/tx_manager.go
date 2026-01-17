package postgres

import (
	"context"
	"fmt"
	"log/slog"
)

const txContextKey = "pgx.Tx"

type TxManager struct {
	db *Database
}

func NewTxManager(db *Database) *TxManager {
	return &TxManager{db: db}
}

// WithTx выполняет функцию внутри транзакции
func (tm *TxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	// Если транзакция уже существует в контексте, просто выполняем функцию
	if tx := extractTx(ctx); tx != nil {
		return fn(ctx)
	}

	// Начинаем новую транзакцию
	tx, err := tm.db.BeginTx(ctx)
	if err != nil {
		slog.Error("failed to begin transaction", "error", err.Error())
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Передаём транзакцию в контекст
	txCtx := context.WithValue(ctx, txContextKey, tx)

	err = fn(txCtx)
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			slog.Error("failed to rollback transaction", "error", rollbackErr.Error())
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("failed to commit transaction", "error", err.Error())
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTxFromContext извлекает транзакцию из контекста
func GetTxFromContext(ctx context.Context) *Tx {
	return extractTx(ctx)
}

func extractTx(ctx context.Context) *Tx {
	tx, ok := ctx.Value(txContextKey).(*Tx)
	if !ok {
		return nil
	}
	return tx
}
