package uow

import (
	"errors"
)

var (
	ErrTransactionNotStarted = errors.New("transaction not started")
)
