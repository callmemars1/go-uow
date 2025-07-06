package uow

import (
	"context"
	"errors"
)

type TxActionWithResult[TRepoRegistry any, TReturn any] func(UOW[TRepoRegistry]) (*TReturn, error)

func RunTxWithResult[TRepoRegistry any, TReturn any](
	ctx context.Context,
	factory Factory[TRepoRegistry],
	action TxActionWithResult[TRepoRegistry, TReturn],
	options TxOptions,
) (*TReturn, error) {
	uow, err := factory.NewUOW(ctx)
	if err != nil {
		return nil, err
	}

	if err := uow.Begin(ctx, options); err != nil {
		return nil, FailedToStartTransaction(err)
	}

	result, err := action(uow)

	if err != nil {
		if rollbackErr := uow.Rollback(ctx); rollbackErr != nil {
			joinedError := errors.Join(err, rollbackErr)
			return nil, FailedToRollback(joinedError)
		}
		return nil, err
	}

	if commitErr := uow.Commit(ctx); commitErr != nil {
		return nil, FailedToCommit(commitErr)
	}

	return result, nil
}

type TxAction[TRepoRegistry any] func(UOW[TRepoRegistry]) error

func RunTx[TRepoRegistry any](
	ctx context.Context,
	factory Factory[TRepoRegistry],
	action TxAction[TRepoRegistry],
	options TxOptions,
) error {
	_, err := RunTxWithResult(ctx, factory, func(uow UOW[TRepoRegistry]) (*any, error) {
		return nil, action(uow)
	}, options)
	return err
}
