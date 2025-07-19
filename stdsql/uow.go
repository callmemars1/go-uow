package stdsql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/callmemars1/go-uow"
)

var _ uow.UOW[any] = &sqlUOW[any]{}

type sqlUOW[TRepoRegistry any] struct {
	tx           *sql.Tx
	repoRegistry TRepoRegistry

	db                  *sql.DB
	repoRegistryFactory RepoRegistryFactory[TRepoRegistry]
}

func newSQLUOW[TRepoRegistry any](db *sql.DB, repoRegistryFactory RepoRegistryFactory[TRepoRegistry]) *sqlUOW[TRepoRegistry] {
	return &sqlUOW[TRepoRegistry]{
		db:                  db,
		repoRegistryFactory: repoRegistryFactory,
	}
}

func (u *sqlUOW[TRepoRegistry]) MustRepoRegistry() TRepoRegistry {
	if u.tx == nil {
		panic("transaction not started")
	}
	return u.repoRegistry
}

func (u *sqlUOW[TRepoRegistry]) Begin(ctx context.Context, options *sql.TxOptions) error {
	tx, err := u.db.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	u.tx = tx

	repoRegistry := u.repoRegistryFactory(tx)
	u.repoRegistry = repoRegistry

	return nil
}

func (u *sqlUOW[TRepoRegistry]) Commit(ctx context.Context) error {
	if u.tx == nil {
		return uow.ErrTransactionNotStarted
	}

	return u.tx.Commit()
}

func (u *sqlUOW[TRepoRegistry]) Rollback(ctx context.Context) error {
	if u.tx == nil {
		return uow.ErrTransactionNotStarted
	}

	return u.tx.Rollback()
}
