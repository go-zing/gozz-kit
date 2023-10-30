package ztree

import (
	"fmt"
	"os/exec"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"github.com/go-zing/gozz-kit/internal/helpers"
)

//go:generate gozz run -p "option" -p "tag" .

// +zz:option
type Option struct {
	ExpandPackages []interface{}
	ExpandTypes    []interface{}
	Unexported     bool
	DocFunc        func(p reflect.Type, field string) string
}

// +zz:tag:json:{{ snake .FieldName }}
type (
	Value struct {
		Id       string            `json:"id"`
		Type     string            `json:"type"`
		Referred int               `json:"referred"`
		Elements map[string]string `json:"elements"`
		Flags    map[string]int    `json:"flags"`
	}

	Type struct {
		Id       string            `json:"id"`
		Kind     string            `json:"kind"`
		Package  string            `json:"package"`
		Name     string            `json:"name"`
		String   string            `json:"string"`
		Elements map[string]string `json:"elements"`
		Docs     map[string]string `json:"docs"`
	}

	Tree struct {
		Values []Value `json:"values"`
		Types  []Type  `json:"types"`
	}
)

const (
	flagAnonymous = 1 << iota
	flagPointer
	flagUnexported
)

func (typ Type) Fullname() string {
	if len(typ.Package) > 0 && len(typ.Name) > 0 {
		return typ.Package + "." + typ.Name
	} else {
		return typ.String
	}
}

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

	expandPkg  map[string]bool
	expandType map[reflect.Type]bool
	types      map[reflect.Type]*Type
	values     map[reflect.Value]*Value
}

func setupParser(v interface{}, opts ...func(*Option)) *parser {
	p := &parser{}

	p.Option.applyOptions(opts...)

	for _, iv := range append(p.Option.ExpandPackages, v) {
		rt := reflect.TypeOf(iv)
		for rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		if p.expandPkg == nil {
			p.expandPkg = make(map[string]bool)
		}
		stdout, _ := exec.Command("go", "list", "-m").Output()
		if pkg := strings.TrimSpace(string(stdout)); len(pkg) > 0 {
			p.expandPkg[pkg] = true
		} else {
			p.expandPkg[rt.PkgPath()] = true
		}
	}

	for _, iv := range append(p.Option.ExpandTypes, v) {
		rt := reflect.TypeOf(iv)
		for rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		if p.expandType == nil {
			p.expandType = make(map[reflect.Type]bool)
		}
		p.expandType[rt] = true
	}
	return p
}

func Parse(v interface{}, opts ...func(*Option)) (tree Tree) {
	p := setupParser(v, opts...)
	p.ParseValues(reflect.ValueOf(v), true)
	return p.tree()
}

func (p *parser) tree() Tree {
	// types
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
	// values
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

func (p *parser) fieldDoc(rt reflect.Type, field string) string {
	if p.Option.DocFunc == nil {
		return ""
	}
	return p.Option.DocFunc(rt, field)
}

func (p *parser) ParseTypes(rt reflect.Type) (id string) {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	typ, ok := p.types[rt]
	if ok {
		return typ.Id
	}

	id = strconv.Itoa(len(p.types))
	typ = &Type{
		Id:       id,
		Kind:     rt.Kind().String(),
		Package:  rt.PkgPath(),
		Name:     rt.Name(),
		String:   rt.String(),
		Elements: make(map[string]string),
		Docs:     map[string]string{"": p.fieldDoc(rt, "")},
	}
	if p.types == nil {
		p.types = make(map[reflect.Type]*Type)
	}
	p.types[rt] = typ

	switch rt.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		typ.Elements[""] = p.ParseTypes(rt.Elem())
	case reflect.Interface:
		for i := 0; i < rt.NumMethod(); i++ {
			vi := rt.Method(i)
			typ.Docs[vi.Name] = p.fieldDoc(rt, vi.Name)
			typ.Elements[vi.Name] = p.ParseTypes(vi.Type)
		}
	case reflect.Struct:
		for i := 0; i < rt.NumField(); i++ {
			ti := rt.Field(i)
			typ.Docs[ti.Name] = p.fieldDoc(rt, ti.Name)
			typ.Elements[ti.Name] = p.ParseTypes(ti.Type)
		}
	}
	return
}

func getUnexportedField(field reflect.Value) reflect.Value {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}

func (p *parser) isExpand(rt reflect.Type) bool {
	rt = helpers.IndirectType(rt)
	for path := range p.expandPkg {
		if strings.HasPrefix(rt.PkgPath(), path) {
			return true
		}
	}
	return p.expandType[rt]
}

func (p *parser) ParseValues(rv reflect.Value, exported bool) (id string) {
	for rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	if !exported && rv.CanAddr() {
		rv = getUnexportedField(rv)
	}

	value := p.values[rv]
	if value != nil {
		value.Referred += 1
		return value.Id
	}

	rt := rv.Type()
	value = &Value{
		Id:       strconv.Itoa(len(p.values)),
		Elements: make(map[string]string),
		Flags:    make(map[string]int),
		Type:     p.ParseTypes(rt),
	}
	if p.values == nil {
		p.values = make(map[reflect.Value]*Value)
	}
	p.values[rv] = value

	if id = value.Id; !(exported || p.Option.Unexported || p.isExpand(rt) || rt.Kind() == reflect.Interface) {
		return
	}

	switch rt.Kind() {
	case reflect.Interface:
		if !rv.IsNil() {
			p.parseValue(value, "", rv.Elem(), exported, false)
		}

	case reflect.Struct:
		if p.isExpand(rt) {
			n := rv.NumField()
			value.Referred += n
			for i := 0; i < n; i++ {
				ti := rt.Field(i)
				p.parseValue(value, ti.Name, rv.Field(i), len(ti.PkgPath) == 0, ti.Anonymous)
			}
		}

	case reflect.Map:
		keys := rv.MapKeys()
		value.Referred += len(keys)
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprintf("%v", keys[i].Interface()) < fmt.Sprintf("%v", keys[j].Interface())
		})
		for i, key := range keys {
			p.parseValue(value, strconv.Itoa(i), rv.MapIndex(key), true, false)
		}

	case reflect.Slice, reflect.Array:
		l := rv.Len()
		value.Referred += l
		for i := 0; i < l; i++ {
			p.parseValue(value, strconv.Itoa(i), rv.Index(i), true, false)
		}
	}
	return
}

func (p *parser) parseValue(v *Value, key string, rv reflect.Value, exported, anonymous bool) {
	v.Flags[key] |= flagAnonymous * helpers.Btoi[anonymous]
	v.Flags[key] |= flagUnexported * helpers.Btoi[!exported]
	v.Flags[key] |= flagPointer * helpers.Btoi[rv.Kind() == reflect.Ptr]
	v.Elements[key] = p.ParseValues(rv, exported)
}
