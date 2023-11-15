package zsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
)

type contextKey int

const (
	contextKeyTxOption contextKey = iota + 1
)

//go:generate gozz run -p "option" ./
// +zz:option
type txOption struct {
	SqlTxOptions *sql.TxOptions
	Rollback     func(rollback func() error, cause error) error
	Recovery     func(exception interface{}) error
}

type txOptions = []func(option *txOption)

func WithTxOptions(ctx context.Context, opts ...func(*txOption)) context.Context {
	if options, ok := ctx.Value(contextKeyTxOption).(*txOptions); ok {
		*options = append(*options, opts...)
	}
	return context.WithValue(ctx, contextKeyTxOption, &opts)
}

func WithTx(ctx context.Context, db DB, fn func(context.Context, Conn) error) (err error) {
	opt := &txOption{}
	if options, ok := ctx.Value(contextKeyTxOption).(*txOptions); ok {
		opt.applyOptions(*options...)
	}

	tx, err := db.BeginTx(ctx, opt.SqlTxOptions)
	if err != nil {
		return
	}

	defer func() {
		if e := recover(); e != nil {
			if opt.Recovery != nil {
				err = opt.Recovery(e)
			} else {
				err = fmt.Errorf("%v", e)
			}
		}

		if err == nil {
			err = tx.Commit()
			return
		}

		if opt.Rollback != nil {
			err = opt.Rollback(tx.Rollback, err)
			return
		}

		switch err {
		case driver.ErrBadConn, context.Canceled:
		default:
			if rerr := tx.Rollback(); rerr != nil {
				err = fmt.Errorf("rollback error %v from error: %w", rerr, err)
			}
		}
	}()
	return fn(ctx, tx)
}
