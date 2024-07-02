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

	Orm interface {
		Conn
		Select(ctx context.Context, model Model, fields []string, ext ...interface{}) (err error)
		Selects(ctx context.Context, models ModelIterator, fields []string, ext ...interface{}) (err error)
		Insert(ctx context.Context, ignore bool, model Model, fields []string, ext ...interface{}) (result sql.Result, err error)
		Inserts(ctx context.Context, ignore bool, models ModelIterator, fields []string, ext ...interface{}) (result sql.Result, err error)
		Update(ctx context.Context, model Model, fields []string, condition string, args ...interface{}) (result sql.Result, err error)
	}
)
