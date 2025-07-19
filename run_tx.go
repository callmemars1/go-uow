package uow

import (
	"context"
	"database/sql"
)

type TxActionWithResult[TRepoRegistry any, TReturn any] func(UOW[TRepoRegistry]) (*TReturn, error)

func RunTxWithResult[TRepoRegistry any, TReturn any](
	ctx context.Context,
	factory Factory[TRepoRegistry],
	action TxActionWithResult[TRepoRegistry, TReturn],
	options *sql.TxOptions,
) (res *TReturn, err error) {
	uow, err := factory.NewUOW(ctx)
	if err != nil {
		return nil, err
	}

	if err = uow.Begin(ctx, options); err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = uow.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = uow.Rollback(ctx)
		} else {
			err = uow.Commit(ctx)
		}
	}()

	res, err = action(uow)
	return
}

type TxAction[TRepoRegistry any] func(UOW[TRepoRegistry]) error

func RunTx[TRepoRegistry any](
	ctx context.Context,
	factory Factory[TRepoRegistry],
	action TxAction[TRepoRegistry],
	options *sql.TxOptions,
) error {
	_, err := RunTxWithResult(ctx, factory, func(uow UOW[TRepoRegistry]) (*any, error) {
		return nil, action(uow)
	}, options)
	return err
}
