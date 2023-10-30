package zsql

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
)

type StmtCacheConn struct {
	Conn  Conn
	cache map[string]*sql.Stmt
	mu    sync.Mutex
}

func NewStmtCacher(conn Conn) *StmtCacheConn {
	if stmt, _ := conn.(*StmtCacheConn); stmt != nil {
		return stmt
	}
	return &StmtCacheConn{Conn: conn}
}

func (sc *StmtCacheConn) Close() error {
	var errs []string
	for _, cached := range sc.cache {
		if e := cached.Close(); e != nil {
			errs = append(errs, e.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, ". "))
}

func (sc *StmtCacheConn) PrepareContext(ctx context.Context, statement string) (*sql.Stmt, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	stmt, ok := sc.cache[statement]
	if ok {
		return stmt, nil
	}
	stmt, err := sc.Conn.PrepareContext(ctx, statement)
	if err == nil {
		if sc.cache == nil {
			sc.cache = make(map[string]*sql.Stmt)
		}
		sc.cache[statement] = stmt
	}
	return stmt, err
}

func (sc *StmtCacheConn) ExecContext(ctx context.Context, statement string, args ...interface{}) (res sql.Result, err error) {
	stmt, err := sc.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	return stmt.ExecContext(ctx, args...)
}

func (sc *StmtCacheConn) QueryContext(ctx context.Context, statement string, args ...interface{}) (rows *sql.Rows, err error) {
	stmt, err := sc.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	return stmt.QueryContext(ctx, args...)
}

func (sc *StmtCacheConn) QueryRowContext(ctx context.Context, statement string, args ...interface{}) *sql.Row {
	stmt, err := sc.PrepareContext(ctx, statement)
	if err != nil {
		return sc.Conn.QueryRowContext(ctx, statement, args...)
	}
	return stmt.QueryRowContext(ctx, args...)
}
