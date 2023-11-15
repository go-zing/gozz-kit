package zapi

import (
	"context"
	"reflect"

	"github.com/go-zing/gozz-kit/internal/helpers"
	"github.com/go-zing/gozz-kit/zreflect"
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
		Kind:    rt.Kind(),
		Type:    rt,
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
		Tags: zreflect.ParseTag(string(tag)),
		Type: p.parseType(rt),
		Flag: helpers.Btoi[unexported]*flagUnexported | helpers.Btoi[rt.Kind() == reflect.Ptr]*flagPointer | helpers.Btoi[anonymous]*flagAnonymous,
	}
}

func (p *Parser) parseFuncPayload(ft reflect.Type) (reflect.Type, reflect.Type) {
	param, result := funcPayload(ft)
	return p.parseType(param), p.parseType(result)
}

func (p *Parser) parseType(rt reflect.Type) reflect.Type {
	if rt == nil {
		return nil
	}
	rt = helpers.IndirectType(rt)
	typ, exist := p.getType(rt)
	if exist {
		return rt
	}
	switch rt.Kind() {
	case reflect.Struct:
		n := rt.NumField()
		for i := 0; i < n; i++ {
			if fi := rt.Field(i); len(fi.PkgPath) == 0 || (fi.Anonymous && fi.Type.Kind() == reflect.Struct) {
				typ.Elements = append(typ.Elements, p.parseFieldElement(rt, fi))
			}
		}
	case reflect.Map, reflect.Array, reflect.Slice:
		typ.Elements = append(typ.Elements, p.parseElement("", "", rt.Elem(), false, false))
	}
	return rt
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
