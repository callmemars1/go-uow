package uow

import (
	"errors"
	"fmt"
)

var (
	ErrTransactionNotStarted = errors.New("transaction not started")
)

func FailedToRollback(err error) error {
	return &FailedToRollbackError{Err: err}
}

type FailedToRollbackError struct {
	Err error
}

func (e *FailedToRollbackError) Error() string {
	return fmt.Errorf("failed to rollback transaction: %w", e.Err).Error()
}

func FailedToCommit(err error) error {
	return &FailedToCommitError{Err: err}
}

type FailedToCommitError struct {
	Err error
}

func (e *FailedToCommitError) Error() string {
	return fmt.Errorf("failed to commit transaction: %w", e.Err).Error()
}

func FailedToStartTransaction(err error) error {
	return &FailedToStartTransactionError{Err: err}
}

type FailedToStartTransactionError struct {
	Err error
}

func (e *FailedToStartTransactionError) Error() string {
	return fmt.Errorf("failed to start transaction: %w", e.Err).Error()
}
