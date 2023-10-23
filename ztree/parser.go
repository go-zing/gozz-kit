package ztree

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

//go:generate gozz run -p "option" -p "tag" .

// +zz:tag:json:{{ snake .FieldName }}
type (
	Value struct {
		Id       string            `json:"id"`
		Type     string            `json:"type"`
		Referred int               `json:"referred"`
		Elements map[string]string `json:"elements"`
	}

	Type struct {
		Id        string            `json:"id"`
		Kind      string            `json:"kind"`
		Package   string            `json:"package"`
		Name      string            `json:"name"`
		String    string            `json:"string"`
		Elements  map[string]string `json:"elements"`
		Anonymous map[string]bool   `json:"anonymous"`
		Docs      map[string]string `json:"docs"`
	}

	Tree struct {
		Values []Value `json:"values"`
		Types  []Type  `json:"types"`
	}
)

func (tree Tree) typesMap() map[string]*Type {
	types := make(map[string]*Type)
	for i, ti := range tree.Types {
		types[ti.Id] = &tree.Types[i]
	}
	return types
}
func (tree Tree) maxReferred() int {
	max := 0
	for _, v := range tree.Values {
		if v.Referred > max {
			max = v.Referred
		}
	}
	return max
}

type parser struct {
	Option Option

	types  map[reflect.Type]*Type
	values map[reflect.Value]*Value
}

// +zz:option
type Option struct {
	Unexported bool
	DocFunc    func(p reflect.Type, field string) string
}

func Parse(v interface{}, opts ...func(*Option)) (tree Tree) {
	p := &parser{Option: Option{
		Unexported: false,
		DocFunc:    func(p reflect.Type, field string) string { return "" },
	}}
	p.Option.applyOptions(opts...)
	p.ParseValues(reflect.ValueOf(v))

	tks := make([]reflect.Type, 0)
	for k := range p.types {
		tks = append(tks, k)
	}
	sort.Slice(tks, func(i, j int) bool {
		i, _ = strconv.Atoi(p.types[tks[i]].Id)
		j, _ = strconv.Atoi(p.types[tks[j]].Id)
		return i < j
	})

	types := make([]Type, 0, len(tks))
	for _, key := range tks {
		types = append(types, *p.types[key])
	}

	vks := make([]reflect.Value, 0)
	for k := range p.values {
		vks = append(vks, k)
	}
	sort.Slice(vks, func(i, j int) bool {
		i, _ = strconv.Atoi(p.values[vks[i]].Id)
		j, _ = strconv.Atoi(p.values[vks[j]].Id)
		return i < j
	})
	values := make([]Value, 0, len(vks))
	for _, key := range vks {
		values = append(values, *p.values[key])
	}

	return Tree{Values: values, Types: types}
}

func (p *parser) ParseTypes(rt reflect.Type) (typ *Type) {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	typ, ok := p.types[rt]
	if ok {
		return typ
	}
	typ = &Type{
		Id:        strconv.Itoa(len(p.types)),
		Kind:      rt.Kind().String(),
		Package:   rt.PkgPath(),
		Name:      rt.Name(),
		String:    strings.Replace(rt.String(), "interface {", "interface{", -1),
		Elements:  make(map[string]string),
		Anonymous: make(map[string]bool),
		Docs:      make(map[string]string),
	}

	if p.types == nil {
		p.types = make(map[reflect.Type]*Type)
	}
	p.types[rt] = typ

	typ.Docs[""] = p.Option.DocFunc(rt, "")

	switch rt.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		typ.Elements[""] = p.ParseTypes(rt.Elem()).Id

	case reflect.Interface:
		n := rt.NumMethod()
		for i := 0; i < n; i++ {
			vi := rt.Method(i)
			typ.Docs[vi.Name] = p.Option.DocFunc(rt, vi.Name)
			typ.Elements[vi.Name] = p.ParseTypes(vi.Type).Id
		}

	case reflect.Struct:
		n := rt.NumField()
		for i := 0; i < n; i++ {
			if ti := rt.Field(i); ti.Anonymous || p.Option.Unexported || len(ti.PkgPath) == 0 {
				typ.Docs[ti.Name] = p.Option.DocFunc(rt, ti.Name)
				if ti.Anonymous {
					typ.Anonymous[ti.Name] = ti.Anonymous
				}
				typ.Elements[ti.Name] = p.ParseTypes(ti.Type).Id
			}
		}
	}
	return
}

func getUnexportedField(field reflect.Value) reflect.Value {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}

func (p *parser) ParseValues(rv reflect.Value) (object *Value) {
	for rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	object, ok := p.values[rv]
	if ok {
		object.Referred += 1
		return object
	}

	object = &Value{
		Id:       strconv.Itoa(len(p.values)),
		Elements: make(map[string]string),
	}
	if p.values == nil {
		p.values = make(map[reflect.Value]*Value)
	}
	p.values[rv] = object

	rt := rv.Type()
	object.Type = p.ParseTypes(rt).Id

	if !(p.Option.Unexported || rv.CanInterface()) {
		return
	}

	switch rt.Kind() {
	case reflect.Map:
		keys := rv.MapKeys()
		object.Referred += len(keys)
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprintf("%v", keys[i].Interface()) < fmt.Sprintf("%v", keys[j].Interface())
		})
		for i, key := range keys {
			object.Elements[strconv.Itoa(i)] = p.ParseValues(rv.MapIndex(key)).Id
		}

	case reflect.Slice, reflect.Array:
		l := rv.Len()
		object.Referred += l
		for i := 0; i < l; i++ {
			object.Elements[strconv.Itoa(i)] = p.ParseValues(rv.Index(i)).Id
		}

	case reflect.Interface:
		if !rv.IsNil() {
			object.Elements[""] = p.ParseValues(rv.Elem()).Id
		}

	case reflect.Struct:
		n := rv.NumField()
		object.Referred += n
		for i := 0; i < n; i++ {
			ti := rt.Field(i)
			vi := rv.Field(i)

			if len(ti.PkgPath) > 0 && vi.CanAddr() {
				vi = getUnexportedField(vi)
			}

			if ti.Anonymous || p.Option.Unexported || len(ti.PkgPath) == 0 {
				object.Elements[ti.Name] = p.ParseValues(vi).Id
			}
		}
	}
	return
}
