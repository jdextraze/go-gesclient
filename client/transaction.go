package client

import (
	"errors"
	"github.com/jdextraze/go-gesclient/tasks"
)

type Transaction struct {
	transactionId   int64
	userCredentials *UserCredentials
	connection      TransactionConnection
	isRolledBack    bool
	isCommitted     bool
}

type TransactionConnection interface {
	TransactionalWriteAsync(*Transaction, []*EventData, *UserCredentials) (*tasks.Task, error)
	CommitTransactionAsync(*Transaction, *UserCredentials) (*tasks.Task, error)
}

var (
	CannotCommitRolledBackTransaction = errors.New("cannot commit a rolled back transaction")
	TransactionIsAlreadyCommitted     = errors.New("transaction is already committed")
)

func NewTransaction(
	transactionId int64,
	userCredentials *UserCredentials,
	connection TransactionConnection,
) *Transaction {
	if transactionId < 0 {
		panic("transactionId must be positive")
	}

	return &Transaction{
		transactionId:   transactionId,
		userCredentials: userCredentials,
		connection:      connection,
	}
}

func (t *Transaction) TransactionId() int64 { return t.transactionId }

// Task.Result() returns *client.WriteResult
func (t *Transaction) CommitAsync() (*tasks.Task, error) {
	if t.isRolledBack {
		return nil, CannotCommitRolledBackTransaction
	}
	if t.isCommitted {
		return nil, TransactionIsAlreadyCommitted
	}
	t.isCommitted = true
	return t.connection.CommitTransactionAsync(t, t.userCredentials)
}

func (t *Transaction) WriteAsync(events []*EventData) (*tasks.Task, error) {
	if t.isRolledBack {
		return nil, CannotCommitRolledBackTransaction
	}
	if t.isCommitted {
		return nil, TransactionIsAlreadyCommitted
	}
	return t.connection.TransactionalWriteAsync(t, events, t.userCredentials)
}

func (t *Transaction) Rollback() error {
	if t.isCommitted {
		return TransactionIsAlreadyCommitted
	}
	t.isRolledBack = true
	return nil
}
