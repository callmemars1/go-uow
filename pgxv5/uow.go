package pgxv5

import (
	"context"
	"database/sql"

	"github.com/callmemars1/go-uow"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ uow.UOW[any] = &pgUOW[any]{}

type pgUOW[TRepoRegistry any] struct {
	tx           pgx.Tx
	repoRegistry TRepoRegistry

	conn                *pgxpool.Conn
	repoRegistryFactory RepoRegistryFactory[TRepoRegistry]
}

func newPgUOW[TRepoRegistry any](conn *pgxpool.Conn, repoRegistryFactory RepoRegistryFactory[TRepoRegistry]) *pgUOW[TRepoRegistry] {
	return &pgUOW[TRepoRegistry]{
		conn:                conn,
		repoRegistryFactory: repoRegistryFactory,
	}
}

func (u *pgUOW[TRepoRegistry]) RepoRegistry() (*TRepoRegistry, error) {
	if u.tx == nil {
		return nil, uow.ErrTransactionNotStarted
	}

	return &u.repoRegistry, nil
}

func (u *pgUOW[TRepoRegistry]) MustRepoRegistry() TRepoRegistry {
	if u.tx == nil {
		panic("transaction not started")
	}
	return u.repoRegistry
}

func (u *pgUOW[TRepoRegistry]) Begin(ctx context.Context, options *sql.TxOptions) error {
	pgxOptions := mapSQLTxOptionsToPgx(options)

	tx, err := u.conn.BeginTx(ctx, pgxOptions)
	if err != nil {
		return err
	}
	u.tx = tx

	repoRegistry := u.repoRegistryFactory(tx)
	u.repoRegistry = repoRegistry

	return nil
}

func (u *pgUOW[TRepoRegistry]) Commit(ctx context.Context) error {
	defer u.conn.Release()
	if u.tx == nil {
		return uow.ErrTransactionNotStarted
	}

	return u.tx.Commit(ctx)
}

func (u *pgUOW[TRepoRegistry]) Rollback(ctx context.Context) error {
	defer u.conn.Release()
	if u.tx == nil {
		return uow.ErrTransactionNotStarted
	}

	return u.tx.Rollback(ctx)
}

func mapSQLTxOptionsToPgx(options *sql.TxOptions) pgx.TxOptions {
	txOptions := pgx.TxOptions{}

	if options == nil {
		return txOptions
	}

	// Map isolation levels from sql to pgx
	switch options.Isolation {
	case sql.LevelReadUncommitted:
		txOptions.IsoLevel = pgx.ReadUncommitted
	case sql.LevelReadCommitted:
		txOptions.IsoLevel = pgx.ReadCommitted
	case sql.LevelRepeatableRead:
		txOptions.IsoLevel = pgx.RepeatableRead
	case sql.LevelSerializable:
		txOptions.IsoLevel = pgx.Serializable
	default:
		// Use default isolation level
		txOptions.IsoLevel = pgx.ReadCommitted
	}

	if options.ReadOnly {
		txOptions.AccessMode = pgx.ReadOnly
	} else {
		txOptions.AccessMode = pgx.ReadWrite
	}

	return txOptions
}
