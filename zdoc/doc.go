package zdoc

import (
	"reflect"
	"sync"

	"github.com/go-zing/gozz-kit/internal/helpers"
)

type Docs struct {
	mu     sync.Mutex
	types  map[reflect.Type]map[string]string
	values map[string]map[interface{}]string
}

func (d *Docs) TypeFieldDoc(rt reflect.Type, field string) string {
	rt = helpers.IndirectType(rt)
	d.mu.Lock()
	m := d.types[rt]
	d.mu.Unlock()
	return m[field]
}

func (d *Docs) LoadTypes(types map[interface{}]map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.types == nil {
		d.types = make(map[reflect.Type]map[string]string)
	}
	for k, m := range types {
		d.types[helpers.IndirectType(reflect.TypeOf(k))] = m
	}
}
