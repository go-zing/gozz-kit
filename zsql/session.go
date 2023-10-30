package zsql

import (
	"context"
	"database/sql"
	"sync"
)

type (
	sessionKey struct{ DB }

	sessionConn struct {
		db   DB
		conn Conn
	}

	sessionTx struct {
		Conn
		sync.Mutex
		commits []func(ctx context.Context)
	}
)

func (conn sessionConn) get(ctx context.Context) Conn {
	if sc, ok := ctx.Value(sessionKey{DB: conn.db}).(Conn); ok {
		return sc
	}
	return conn.conn
}

func (conn sessionConn) QueryContext(ctx context.Context, statement string, args ...interface{}) (rows *sql.Rows, err error) {
	return conn.get(ctx).QueryContext(ctx, statement, args...)
}

func (conn sessionConn) QueryRowContext(ctx context.Context, statement string, args ...interface{}) (row *sql.Row) {
	return conn.get(ctx).QueryRowContext(ctx, statement, args...)
}

func (conn sessionConn) ExecContext(ctx context.Context, statement string, args ...interface{}) (res sql.Result, err error) {
	return conn.get(ctx).ExecContext(ctx, statement, args...)
}

func (conn sessionConn) PrepareContext(ctx context.Context, statement string) (stmt *sql.Stmt, err error) {
	return conn.get(ctx).PrepareContext(ctx, statement)
}

func (stx *sessionTx) commit(ctx context.Context) {
	wg := &sync.WaitGroup{}
	wg.Add(len(stx.commits))
	for i := range stx.commits {
		go func(i int) { defer wg.Done(); stx.commits[i](ctx) }(i)
	}
	wg.Wait()
}

func SessionConn(db DB, opts ...func(Conn) Conn) Conn {
	conn := Conn(db)
	for _, opt := range opts {
		conn = opt(conn)
	}
	return sessionConn{db: db, conn: conn}
}

func WithSessionTx(ctx context.Context, db DB, fn func(context.Context) error, onCommits ...func(ctx context.Context)) (err error) {
	key := sessionKey{DB: db}
	if stx, in := ctx.Value(key).(*sessionTx); in {
		stx.Lock()
		stx.commits = append(stx.commits, onCommits...)
		stx.Unlock()
		return fn(ctx)
	} else if err = WithTx(ctx, db, func(ctx context.Context, conn Conn) error {
		stx = &sessionTx{commits: onCommits, Conn: conn}
		return fn(context.WithValue(ctx, key, stx))
	}); err == nil && len(stx.commits) > 0 {
		stx.commit(ctx)
	}
	return
}
