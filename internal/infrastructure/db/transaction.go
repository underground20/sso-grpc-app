package db

import (
	"context"
	"log/slog"
)

type TransactionExecutor struct {
	db     *Database
	logger *slog.Logger
}

func NewTransactionExecutor(db *Database, logger *slog.Logger) TransactionExecutor {
	return TransactionExecutor{db: db, logger: logger}
}

func (t TransactionExecutor) Execute(ctx context.Context, handler func() error) error {
	transaction, err := t.db.Conn.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rollbackErr := transaction.Rollback(ctx); rollbackErr != nil {
				t.logger.Error("failed to rollback transaction", slog.String("error", rollbackErr.Error()))
			}
		}
	}()

	if err = handler(); err != nil {
		return err
	}

	if err = transaction.Commit(ctx); err != nil {
		return err
	}

	return nil
}
