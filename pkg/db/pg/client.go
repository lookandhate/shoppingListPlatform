package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lookandhate/course_platform_lib/pkg/db/db"
	"github.com/pkg/errors"
)

type pgClient struct {
	masterDbc db.DB
}

func New(ctx context.Context, dsn string) (db.Client, error) {
	dbc, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, errors.Errorf("failed to connect to db: %v", err)
	}

	return &pgClient{
		masterDbc: &pg{dbc: dbc},
	}, nil
}

func (c *pgClient) DB() db.DB {
	return c.masterDbc
}

func (c *pgClient) Close() error {
	if c.masterDbc != nil {
		c.masterDbc.Close()
	}

	return nil
}
