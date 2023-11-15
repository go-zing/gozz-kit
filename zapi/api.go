package zapi

import (
	"context"
	"reflect"
	"strings"

	"github.com/go-zing/gozz-kit/internal/helpers"
	"github.com/go-zing/gozz-kit/zreflect"
)

type (
	InvokeFunc = func(ctx context.Context, dec func(interface{}) error) (ret interface{}, err error)

	Api struct {
		Name     string
		Doc      string
		Resource string
		Options  map[string]string
		Invoke   InvokeFunc
		Request  reflect.Type
		Response reflect.Type

		function func() interface{}
	}

	ApiGroup struct {
		Handler interface{}
		Package string
		Name    string
		Doc     string
		Apis    []Api
	}

	HttpApi struct {
		Api
		Method string
		Path   string
	}

	Iterator interface {
		Iterate(fn func(interface{}, []map[string]interface{}))
	}

	PayloadType struct {
		Kind     reflect.Kind
		Type     reflect.Type
		Name     string
		Package  string
		Doc      string
		Entity   interface{}
		Elements []PayloadElement
	}

	PayloadElement struct {
		Type reflect.Type
		Flag int
		Name string
		Doc  string
		Tags zreflect.StructTags
	}
)

const (
	flagUnexported = 1 << iota
	flagPointer
	flagAnonymous
)

func (group ApiGroup) Fullname() string {
	if len(group.Package) > 0 {
		return group.Package + "." + group.Name
	}
	if len(group.Name) > 0 {
		return group.Name
	}
	return ""
}

func (api *Api) Func() interface{} {
	return api.function()
}

func (typ PayloadType) Fullname() string {
	if len(typ.Package) > 0 {
		return typ.Package + "." + typ.Name
	}
	if len(typ.Name) > 0 {
		return typ.Name
	}
	return ""
}

func SplitFn(sep string) func(resource string) (method, path string) {
	return func(resource string) (method, path string) {
		sp := strings.SplitN(resource, sep, 2)[:2]
		return sp[0], sp[1]
	}
}

func (p *Parser) parseApi(rv reflect.Value, rt reflect.Type, spec map[string]interface{}) (api Api) {
	api.Name, _ = spec["name"].(string)
	fm, ok := rt.MethodByName(api.Name)
	if !ok {
		return
	}
	api.function = func() interface{} { return rv.MethodByName(api.Name) }
	api.Doc = p.getFieldDoc(rt, api.Name)
	api.Request, api.Response = p.parseFuncPayload(fm.Type)
	api.Resource, _ = spec["resource"].(string)
	api.Options, _ = spec["options"].(map[string]string)
	api.Invoke, _ = spec["invoke"].(InvokeFunc)
	return
}

func (p *Parser) parseApiGroup(handler interface{}, specs []map[string]interface{}) ApiGroup {
	rt := helpers.IndirectType(reflect.TypeOf(handler))
	rv := reflect.Indirect(reflect.ValueOf(handler))
	group := ApiGroup{
		Handler: rv.Interface(),
		Package: rt.PkgPath(),
		Name:    rt.Name(),
		Doc:     p.getFieldDoc(rt, ""),
		Apis:    make([]Api, 0, len(specs)),
	}
	for _, spec := range specs {
		if api := p.parseApi(rv, rt, spec); len(api.Name) > 0 {
			group.Apis = append(group.Apis, api)
		}
	}
	return group
}

func (p *Parser) Parse(iterator Iterator) (groups []ApiGroup, payloads map[reflect.Type]*PayloadType) {
	iterator.Iterate(func(handler interface{}, specs []map[string]interface{}) {
		groups = append(groups, p.parseApiGroup(handler, specs))
	})

	payloads = make(map[reflect.Type]*PayloadType, len(p.types))
	for _, typ := range p.types {
		payloads[typ.Type] = typ
	}
	return
}
