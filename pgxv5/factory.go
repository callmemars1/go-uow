package pgxv5

import (
	"context"

	"github.com/callmemars1/go-uow"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ uow.Factory[any] = &Factory[any]{}

type RepoRegistryFactory[TRepoRegistry any] func(tx pgx.Tx) TRepoRegistry

type Factory[TRepoRegistry any] struct {
	pool *pgxpool.Pool

	repoRegistryFactory RepoRegistryFactory[TRepoRegistry]
}

func NewFactory[TRepoRegistry any](
	pool *pgxpool.Pool,
	repoRegistryFactory RepoRegistryFactory[TRepoRegistry],
) (*Factory[TRepoRegistry], error) {
	return &Factory[TRepoRegistry]{
		pool:                pool,
		repoRegistryFactory: repoRegistryFactory,
	}, nil
}

func (f *Factory[TRepoRegistry]) NewUOW(ctx context.Context) (uow.UOW[TRepoRegistry], error) {
	conn, err := f.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	uow := newPgUOW(conn, f.repoRegistryFactory)

	return uow, nil
}

func (f *Factory[TRepoRegistry]) Release() error {
	f.pool.Close()
	return nil
}
