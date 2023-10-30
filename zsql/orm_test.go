package zsql_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/go-zing/gozz-kit/zsql"
)

type assertSql struct {
	Statement string
	Args      []interface{}
}

var successError = errors.New("assert success")

func (m assertSql) assert(statement string, args []interface{}) (err error) {
	err = fmt.Errorf("want %v %v ,got %v %v", m.Statement, m.Args, statement, args)
	if statement != m.Statement || len(args) != len(m.Args) {
		return
	}
	for i, arg := range args {
		if arg != m.Args[i] {
			return
		}
	}
	return successError
}

func (m assertSql) ExecContext(ctx context.Context, statement string, args ...interface{}) (sql.Result, error) {
	return nil, m.assert(statement, args)
}

func (m assertSql) QueryContext(ctx context.Context, statement string, args ...interface{}) (*sql.Rows, error) {
	return nil, m.assert(statement, args)
}

func (m assertSql) PrepareContext(ctx context.Context, statement string) (*sql.Stmt, error) {
	return nil, nil
}

func (m assertSql) QueryRowContext(ctx context.Context, statement string, args ...interface{}) *sql.Row {
	return nil
}

func newAssert(statement string, args ...interface{}) zsql.Litorm {
	return zsql.Litorm{Conn: assertSql{Statement: statement, Args: args}}
}

var ctx = context.Background()

type T struct {
	FieldA string
	FieldB string
}

func (t T) TableName() string { return "test" }

func (t *T) FieldMapping(dst map[string]interface{}) {
	dst["field_a"] = &t.FieldA
	dst["field_b"] = &t.FieldB
}

func check(t *testing.T, err error) {
	if t.Helper(); !errors.Is(err, successError) {
		t.Fatal(err)
	}
}

func TestSelect(t *testing.T) {
	check(t, newAssert("SELECT `field_a`,`field_b` FROM `test`").
		Select(ctx, &T{}, nil))
	check(t, newAssert("SELECT `field_a` FROM `test`").
		Select(ctx, &T{}, []string{"field_a"}))
	check(t, newAssert("SELECT `field_a` FROM `test` WHERE `field_b` = ?", 1).
		Select(ctx, &T{}, []string{"field_a"}, "WHERE `field_b` = ?", 1))
	check(t, newAssert("SELECT `field_a`,sum(*) FROM `test` GROUP BY $1").
		Select(ctx, &T{}, []string{"field_a", "sum(*)"}, "GROUP BY $1"))
}

func TestUpdate(t *testing.T) {
	v := &T{}
	{
		_, err := newAssert(
			"UPDATE `test` SET `field_a` = ?,`field_b` = ? WHERE `field_b` = ?", &v.FieldA, &v.FieldB, 1).
			Update(ctx, v, nil, "WHERE `field_b` = ?", 1)
		check(t, err)
	}
	{
		_, err := newAssert(
			"UPDATE `test` SET `field_b` = ? WHERE `field_b` = ?", &v.FieldB, 1).
			Update(ctx, v, []string{`field_b`}, "WHERE `field_b` = ?", 1)
		check(t, err)
	}
}

type sliceT []T

func (s *sliceT) Iterate(f func(v interface{}, alloc bool) (next bool)) {
	for i := 0; ; i++ {
		if c := i >= len(*s); !c {
			if !f(&(*s)[i], c) {
				return
			}
		} else if n := append(*s, T{}); f(&n[i], c) {
			*s = n
		} else {
			*s = n[:i]
			return
		}
	}
}

func TestInsert(t *testing.T) {
	v := &T{}
	{
		_, err := newAssert(
			"INSERT IGNORE INTO `test` (`field_a`,`field_b`) VALUES (?,?)", &v.FieldA, &v.FieldB).
			Insert(ctx, true, v, nil)
		check(t, err)
	}
	{
		_, err := newAssert(
			"INSERT IGNORE INTO `test` (`field_a`,`field_b`) VALUES (?,?) ON DUPLICATE KEY UPDATE `field_a` = ?",
			&v.FieldA, &v.FieldB, 1).
			Insert(ctx, true, v, nil, "ON DUPLICATE KEY UPDATE `field_a` = ?", 1)
		check(t, err)
	}
	{
		_, err := newAssert(
			"INSERT IGNORE INTO `test` (`field_a`) VALUES (?)", &v.FieldA).
			Insert(ctx, true, v, []string{"field_a"})
		check(t, err)
	}
	{
		_, err := newAssert(
			"INSERT INTO `test` (`field_a`) VALUES (?)", &v.FieldA).
			Insert(ctx, false, v, []string{"field_a"})
		check(t, err)
	}
	{
		st := &sliceT{{}, {}}
		_, err := newAssert(
			"INSERT INTO `test` (`field_a`) VALUES (?),(?)", &(*st)[0].FieldA, &(*st)[1].FieldA).
			Inserts(ctx, false, st, []string{"field_a"})
		check(t, err)
	}
	{
		st := &sliceT{{}, {}}
		_, err := newAssert(
			"INSERT INTO `test` (`field_a`,`field_b`) VALUES (?,?),(?,?)",
			&(*st)[0].FieldA, &(*st)[0].FieldB, &(*st)[1].FieldA, &(*st)[1].FieldB,
		).
			Inserts(ctx, false, st, nil)
		check(t, err)
	}
}
