package postgres

import (
	"context"
	"errors"
	"testing"
)

func TestIsConnectionError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "pool not initialized", err: errors.New("database pool is not initialized"), want: true},
		{name: "connection refused", err: errors.New("connection refused"), want: true},
		{name: "connection reset", err: errors.New("connection reset by peer"), want: true},
		{name: "timeout", err: errors.New("i/o timeout"), want: true},
		{name: "other", err: errors.New("something else"), want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := isConnectionError(tc.err); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestDatabase_QueryRowWithoutPool(t *testing.T) {
	db := NewDB(&DBConfig{Host: "localhost", Port: "5432", User: "u", Password: "p", DBName: "db", SSLMode: "disable"})
	row := db.QueryRow(context.Background(), "select 1")
	var n int
	err := row.Scan(&n)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDatabase_QueryWithoutPool(t *testing.T) {
	db := NewDB(&DBConfig{Host: "localhost", Port: "5432", User: "u", Password: "p", DBName: "db", SSLMode: "disable"})
	_, err := db.Query(context.Background(), "select 1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDatabase_ExecWithoutPool(t *testing.T) {
	db := NewDB(&DBConfig{Host: "localhost", Port: "5432", User: "u", Password: "p", DBName: "db", SSLMode: "disable"})
	_, err := db.Exec(context.Background(), "select 1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDatabase_WithRetry_NonConnectionError(t *testing.T) {
	db := NewDB(&DBConfig{Host: "localhost", Port: "5432", User: "u", Password: "p", DBName: "db", SSLMode: "disable"})
	calls := 0
	err := db.WithRetry(context.Background(), 3, func(ctx context.Context) error {
		calls++
		return errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}
