package zsql

import (
	"context"
	"database/sql/driver"
	"fmt"
)

func WithTx(ctx context.Context, db DB, fn func(context.Context, Conn) error) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
		switch err {
		case nil:
			err = tx.Commit()
		case driver.ErrBadConn, context.Canceled:
		default:
			if e := tx.Rollback(); e != nil {
				err = fmt.Errorf("rollback error %v from error: %w", e, err)
			}
		}
	}()
	return fn(ctx, tx)
}
