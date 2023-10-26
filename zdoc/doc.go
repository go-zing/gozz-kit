package zdoc

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-zing/gozz-kit/internal/helpers"
)

type (
	Docs struct {
		mu     sync.Mutex
		types  map[reflect.Type]map[string]string
		values map[string]map[interface{}]string
	}

	TypesDoc map[interface{}]map[string]string
)

func Split(doc string) (title, description string) {
	sp := strings.SplitN(doc, "\n", 2)[:2]
	return strings.TrimSpace(sp[0]), strings.TrimSpace(sp[1])
}

func (d TypesDoc) TypeFieldDoc(rt reflect.Type, field string) string {
	for k, v := range d {
		ct, ok := k.(reflect.Type)
		if !ok {
			ct = helpers.IndirectType(reflect.TypeOf(k))
			d[ct] = v
		}
		if ct == rt {
			return v[field]
		}
	}
	return ""
}

func (d *Docs) TypeFieldDoc(rt reflect.Type, field string) string {
	rt = helpers.IndirectType(rt)
	d.mu.Lock()
	m := d.types[rt]
	d.mu.Unlock()
	return m[field]
}

func (d *Docs) LoadTypes(types TypesDoc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.types == nil {
		d.types = make(map[reflect.Type]map[string]string)
	}
	for k, m := range types {
		d.types[helpers.IndirectType(reflect.TypeOf(k))] = m
	}
}

func (d *Docs) LoadValues(values map[string]map[interface{}]string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.values == nil {
		d.values = make(map[string]map[interface{}]string)
	}
	for k, m := range values {
		d.values[k] = m
	}
}
