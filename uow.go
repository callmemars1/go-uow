package uow

import (
	"context"
	"database/sql"
)

type UOW[TRepoRegistry any] interface {
	MustRepoRegistry() TRepoRegistry
	Begin(ctx context.Context, options *sql.TxOptions) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Factory[TRepoRegistry any] interface {
	NewUOW(ctx context.Context) (UOW[TRepoRegistry], error)
	Release() error
}
