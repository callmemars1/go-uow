package stdsql

import (
	"context"
	"database/sql"

	"github.com/callmemars1/go-uow"
)

var _ uow.Factory[any] = &Factory[any]{}

type RepoRegistryFactory[TRepoRegistry any] func(tx *sql.Tx) TRepoRegistry

type Factory[TRepoRegistry any] struct {
	db *sql.DB

	repoRegistryFactory RepoRegistryFactory[TRepoRegistry]
}

func NewFactory[TRepoRegistry any](
	db *sql.DB,
	repoRegistryFactory RepoRegistryFactory[TRepoRegistry],
) (*Factory[TRepoRegistry], error) {
	return &Factory[TRepoRegistry]{
		db:                  db,
		repoRegistryFactory: repoRegistryFactory,
	}, nil
}

func (f *Factory[TRepoRegistry]) NewUOW(ctx context.Context) (uow.UOW[TRepoRegistry], error) {
	uow := newSQLUOW(f.db, f.repoRegistryFactory)

	return uow, nil
}

func (f *Factory[TRepoRegistry]) Release() error {
	return f.db.Close()
}
