package transaction

import (
	"context"

	"github.com/pkg/errors"

	"github.com/jackc/pgx/v5"
	"github.com/lookandhate/course_auth/internal/client/db"
	"github.com/lookandhate/course_auth/internal/client/db/pg"
)

type manager struct {
	db db.Transactor
}

// NewTransactionManager creates db.TxManager implementation.
func NewTransactionManager(db db.Transactor) db.TxManager {
	return &manager{
		db: db,
	}
}

// transaction function that calls passed handler in a transaction.
func (m *manager) transaction(ctx context.Context, opts pgx.TxOptions, fn db.Handler) (err error) {
	// If it is a wrapped transaction - just run the handler.
	tx, ok := ctx.Value(pg.TxKey).(pgx.Tx)
	if ok {
		return fn(ctx)
	}

	// Start new transaction.
	tx, err = m.db.BeginTx(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "can't begin transaction")
	}

	ctx = pg.MakeContextTx(ctx, tx)

	// defer cleaning up in case of failing.
	defer func() {
		// восстанавливаемся после паники
		if r := recover(); r != nil {
			err = errors.Errorf("panic recovered: %v", r)
		}

		// откатываем транзакцию, если произошла ошибка
		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = errors.Wrapf(err, "errRollback: %v", errRollback)
			}

			return
		}
		err = tx.Commit(ctx)
		if err != nil {
			err = errors.Wrap(err, "tx commit failed")
		}
	}()
	// Run handler.
	if err = fn(ctx); err != nil {
		err = errors.Wrap(err, "failed executing code inside transaction")
	}

	return err
}

func (m *manager) ReadCommitted(ctx context.Context, f db.Handler) error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	return m.transaction(ctx, txOpts, f)
}
