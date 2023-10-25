package zapi

import (
	"context"
	"reflect"
	"strconv"
)

//go:generate gozz run -p "option" ./

// +zz:option
type Parser struct {
	docFunc func(p reflect.Type, field string) string
	types   map[reflect.Type]*PayloadType
}

var (
	rTypeContext = reflect.TypeOf((*context.Context)(nil)).Elem()
	rTypeError   = reflect.TypeOf((*error)(nil)).Elem()

	btoi = map[bool]int{true: 1}
)

func (e PayloadElement) CheckFlag(flag int) bool { return e.Flag&flag > 0 }
func (e PayloadElement) IsUnexported() bool      { return e.CheckFlag(flagUnexported) }
func (e PayloadElement) IsPointer() bool         { return e.CheckFlag(flagPointer) }
func (e PayloadElement) IsAnonymous() bool       { return e.CheckFlag(flagAnonymous) }

func NewParser(options ...func(parser *Parser)) *Parser {
	p := &Parser{}
	p.applyOptions(options...)
	return p
}

func (p *Parser) getFieldDoc(rt reflect.Type, field string) string {
	if p.docFunc == nil {
		return ""
	}
	return p.docFunc(rt, field)
}

func (p *Parser) getType(rt reflect.Type) (typ *PayloadType, exist bool) {
	if typ, exist = p.types[rt]; exist {
		return typ, exist
	} else if p.types == nil {
		p.types = make(map[reflect.Type]*PayloadType)
	}

	entity := reflect.New(rt)

	if v, ok := entity.Interface().(interface{ Init() }); ok {
		v.Init()
	}

	p.types[rt] = &PayloadType{
		Id:      len(p.types),
		Kind:    rt.Kind().String(),
		Name:    rt.Name(),
		Package: rt.PkgPath(),
		Doc:     p.getFieldDoc(rt, ""),
		Entity:  entity.Elem().Interface(),
	}
	return p.types[rt], false
}

func (p *Parser) parseFieldElement(rt reflect.Type, field reflect.StructField) PayloadElement {
	pe := p.parseElement(field.Name, field.Tag, field.Type, len(field.PkgPath) > 0, field.Anonymous)
	pe.Doc = p.getFieldDoc(rt, field.Name)
	return pe
}

func (p *Parser) parseElement(name string, tag reflect.StructTag, rt reflect.Type, unexported, anonymous bool) PayloadElement {
	return PayloadElement{
		Name: name,
		Tags: parseReflectTag(string(tag)),
		Type: p.parseType(rt),
		Flag: btoi[unexported]*flagUnexported | btoi[rt.Kind() == reflect.Ptr]*flagPointer | btoi[anonymous]*flagAnonymous,
	}
}

func indirectType(rt reflect.Type) reflect.Type {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	return rt
}

func (p *Parser) parseFuncPayload(ft reflect.Type) (int, int) {
	param, result := funcPayload(ft)
	return p.parseType(param), p.parseType(result)
}

func (p *Parser) parseType(rt reflect.Type) (id int) {
	if rt == nil {
		return -1
	}
	rt = indirectType(rt)
	typ, exist := p.getType(rt)
	if id = typ.Id; exist {
		return
	}
	switch rt.Kind() {
	case reflect.Struct:
		n := rt.NumField()
		for i := 0; i < n; i++ {
			if fi := rt.Field(i); len(fi.PkgPath) == 0 || (fi.Anonymous && fi.Type.Kind() != reflect.Ptr) {
				typ.Elements = append(typ.Elements, p.parseFieldElement(rt, fi))
			}
		}
	case reflect.Map, reflect.Array, reflect.Slice:
		typ.Elements = append(typ.Elements, p.parseElement("", "", rt.Elem(), false, false))
	}
	return
}

func funcPayload(ft reflect.Type) (param, result reflect.Type) {
	if ft.Kind() != reflect.Func {
		return
	}
	in := ft.NumIn()
	for i := 0; i < in; i++ {
		if vi := ft.In(i); vi != rTypeContext {
			param = vi
			break
		}
	}
	out := ft.NumOut()
	for i := 0; i < out; i++ {
		if vi := ft.Out(out - 1 - i); vi != rTypeError {
			result = vi
			break
		}
	}
	return
}

type StructTag struct {
	Key   string
	Value string
}

type StructTags []StructTag

func (tags StructTags) Lookup(key string) (value string, found bool) {
	for i := range tags {
		if tags[i].Key == key {
			return tags[i].Value, true
		}
	}
	return
}

func (tags StructTags) Get(key string) (value string) {
	value, _ = tags.Lookup(key)
	return
}

func parseReflectTag(tag string) (tags StructTags) {
	for tag != "" {
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}

		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}

		key := tag[:i]
		tag = tag[i+1:]

		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}

		if i >= len(tag) {
			break
		}

		if value, err := strconv.Unquote(tag[:i+1]); err == nil {
			tags = append(tags, StructTag{
				Key:   key,
				Value: value,
			})
			tag = tag[i+1:]
		}
	}
	return
}
