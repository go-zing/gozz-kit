package zsql

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
)

var ErrInvalidModelsIterator = errors.New("invalid models iterator")

type (
	ModelIterator interface {
		Iterate(fn func(v interface{}, alloc bool) (next bool))
	}

	Model interface {
		TableName() string
		FieldMapping(dst map[string]interface{})
	}

	modelItem struct{ Model }

	FieldMapping map[string]interface{}

	Litorm struct{ Conn }

	SqlBuilder struct{ strings.Builder }
)

func (item modelItem) Iterate(f func(interface{}, bool) bool) { f(item, false) }

func (mapping FieldMapping) MapFields(model Model, fp *[]string) {
	model.FieldMapping(mapping)
	if fields := *fp; len(fields) == 0 {
		fields = make([]string, 0, len(mapping))
		for k := range mapping {
			if len(k) > 0 {
				fields = append(fields, k)
			}
		}
		sort.Strings(fields)
		*fp = fields
	}
}

func (mapping FieldMapping) MapValues(fields []string, ptr *[]interface{}) {
	for _, field := range fields {
		*ptr = append(*ptr, mapping[field])
	}
}

func (orm Litorm) Insert(ctx context.Context, ignore bool, model Model, fields []string, ext ...interface{}) (result sql.Result, err error) {
	return orm.Inserts(ctx, ignore, modelItem{Model: model}, fields, ext...)
}

func (orm Litorm) Inserts(ctx context.Context, ignore bool, models ModelIterator, fields []string, ext ...interface{}) (result sql.Result, err error) {
	statement := new(SqlBuilder)
	if ext, err = statement.BuildInsert(models, ignore, fields, ext); err != nil {
		return
	}
	return orm.ExecContext(ctx, statement.String(), ext...)
}

func (orm Litorm) Update(ctx context.Context, model Model, fields []string, condition string, args ...interface{}) (result sql.Result, err error) {
	statement := new(SqlBuilder)
	args = statement.BuildUpdate(model, fields, condition, args)
	return orm.ExecContext(ctx, statement.String(), args...)
}

func (orm Litorm) Selects(ctx context.Context, models ModelIterator, fields []string, ext ...interface{}) (err error) {
	if _, err = orm.selects(ctx, models, fields, ext...); err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (orm Litorm) Select(ctx context.Context, model Model, fields []string, ext ...interface{}) (err error) {
	_, err = orm.selects(ctx, modelItem{Model: model}, fields, ext...)
	return
}

func (orm Litorm) selects(ctx context.Context, models ModelIterator, fields []string, ext ...interface{}) (rows *sql.Rows, err error) {
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	mapping := make(FieldMapping, len(fields))
	models.Iterate(func(v interface{}, alloc bool) (next bool) {
		if model, ok := v.(Model); !ok {
			err = ErrInvalidModelsIterator
		} else if mapping.MapFields(model, &fields); rows == nil {
			statement := new(SqlBuilder)
			ext = statement.BuildSelect(model, fields, ext)
			rows, err = orm.QueryContext(ctx, statement.String(), ext...)
			if ext = ext[:0]; err == nil && rows.Next() {
				mapping.MapValues(fields, &ext)
				err = rows.Scan(ext...)
			} else if err == nil {
				err = sql.ErrNoRows
			}
			return err == nil
		} else if ext = ext[:0]; rows.Next() {
			mapping.MapValues(fields, &ext)
			err = rows.Scan(ext...)
			return err == nil
		}
		return false
	})
	return
}

func (bd *SqlBuilder) BuildUpdate(model Model, fields []string, ext string, xargs []interface{}) (args []interface{}) {
	mapping := make(FieldMapping, len(fields))
	mapping.MapFields(model, &fields)
	bd.WriteString("UPDATE ")
	bd.WriteTable(model.TableName())
	bd.WriteString(" SET ")
	bd.WriteFields(fields, true, " = ?", ",")
	if mapping.MapValues(fields, &args); len(ext) > 0 {
		bd.WriteRune(' ')
		bd.WriteString(ext)
		args = append(args, xargs...)
	}
	return
}

func (bd *SqlBuilder) BuildInsert(models ModelIterator, ignore bool, fields []string, ext []interface{}) (args []interface{}, err error) {
	mapping := make(FieldMapping, len(fields))
	if models.Iterate(func(v interface{}, alloc bool) (next bool) {
		if model, ok := v.(Model); alloc || !ok {
			return
		} else if mapping.MapFields(model, &fields); bd.Len() == 0 {
			if ignore {
				bd.WriteString("INSERT IGNORE INTO ")
			} else {
				bd.WriteString("INSERT INTO ")
			}
			bd.WriteTable(model.TableName())
			bd.WriteString(" (")
			bd.WriteFields(fields, true, "", ",")
			bd.WriteString(") VALUES (")
		} else {
			bd.WriteString(",(")
		}
		bd.WriteFields(fields, false, "?", ",")
		bd.WriteString(")")
		mapping.MapValues(fields, &args)
		return true
	}); bd.Len() == 0 {
		return nil, ErrInvalidModelsIterator
	}
	bd.WriteExtArgs(ext, &args)
	return
}

func (bd *SqlBuilder) quote(v string) { bd.WriteRune('`'); bd.WriteString(v); bd.WriteRune('`') }

func (bd *SqlBuilder) BuildSelect(model Model, fields []string, ext []interface{}) (args []interface{}) {
	bd.WriteString("SELECT ")
	bd.WriteFields(fields, true, "", ",")
	bd.WriteString(" FROM ")
	bd.WriteTable(model.TableName())
	bd.WriteExtArgs(ext, &args)
	return
}

func (bd *SqlBuilder) WriteTable(table string) { bd.quote(table) }

func (bd *SqlBuilder) WriteFields(fields []string, name bool, suffix, sep string) {
	for i, field := range fields {
		if len(field) == 0 {
			continue
		} else if name {
			if !strings.ContainsAny(field, "`(,") {
				bd.quote(field)
			} else {
				bd.WriteString(field)
			}
		}
		if bd.WriteString(suffix); len(fields)-1 != i {
			bd.WriteString(sep)
		}
	}
}

func (bd *SqlBuilder) WriteExtArgs(ext []interface{}, args *[]interface{}) {
	if len(ext) > 0 {
		if expr, ok := (ext)[0].(string); ok {
			bd.WriteRune(' ')
			bd.WriteString(expr)
			*args = append(*args, ext[1:]...)
		}
	}
}
