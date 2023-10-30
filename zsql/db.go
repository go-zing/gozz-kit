package zsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
)

type (
	DB interface {
		Close() error
		Stats() sql.DBStats
		PingContext(ctx context.Context) error
		BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
		Conn
	}

	Conn interface {
		QueryContext(ctx context.Context, statement string, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, statement string, args ...interface{}) *sql.Row
		ExecContext(ctx context.Context, statement string, args ...interface{}) (sql.Result, error)
		PrepareContext(ctx context.Context, statement string) (*sql.Stmt, error)
	}

	Tx interface {
		Conn
		driver.Tx
	}
)
