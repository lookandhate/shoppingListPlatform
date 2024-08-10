package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Handler - function called in transaction.
type Handler func(ctx context.Context) error

// Client Database client.
type Client interface {
	DB() DB
	Close() error
}

// TxManager transaction manager that calls passed functions as handler.
type TxManager interface {
	ReadCommitted(ctx context.Context, f Handler) error
}

// Query wrapper on query that saves query itself and it's name.
type Query struct {
	Name     string
	QueryRaw string
}

// Transactor interface for transactions.
type Transactor interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// SQLExecer combines NamedExecer and QueryExecer.
type SQLExecer interface {
	NamedExecer
	QueryExecer
}

// NamedExecer used for work with structs with named tags.
type NamedExecer interface {
	ScanOneContext(ctx context.Context, dest interface{}, q Query, args ...interface{}) error
	ScanAllContext(ctx context.Context, dest interface{}, q Query, args ...interface{}) error
}

// QueryExecer used for plain queries.
type QueryExecer interface {
	ExecContext(ctx context.Context, q Query, args ...interface{}) (pgconn.CommandTag, error)
	QueryContext(ctx context.Context, q Query, args ...interface{}) (pgx.Rows, error)
	QueryRowContext(ctx context.Context, q Query, args ...interface{}) pgx.Row
}

// Pinger used to check db connection.
type Pinger interface {
	Ping(ctx context.Context) error
}

// DB interface to work with db.
type DB interface {
	SQLExecer
	Transactor
	Pinger
	Close()
}
