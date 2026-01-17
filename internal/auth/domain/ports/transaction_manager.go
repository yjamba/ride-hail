package ports

import "context"

type TransactionManager interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}
