package zapi

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type (
	InvokeFunc = func(ctx context.Context, dec func(interface{}) error) (ret interface{}, err error)

	Api struct {
		Name     string
		Doc      string
		Resource string
		Options  map[string]string
		Invoke   InvokeFunc
		Param    int
		Result   int
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
		Method     string
		Path       string
		PathParams []string
	}

	Iterator interface {
		Iterate(fn func(interface{}, []map[string]interface{}))
	}

	PayloadType struct {
		Id       int
		Kind     string
		Name     string
		Package  string
		Doc      string
		Entity   interface{}
		Elements []PayloadElement
	}

	PayloadElement struct {
		Type int
		Flag int
		Name string
		Doc  string
		Tags StructTags
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

func (typ PayloadType) Fullname() string {
	if len(typ.Package) > 0 {
		return typ.Package + "." + typ.Name
	}
	if len(typ.Name) > 0 {
		return typ.Name
	}
	return fmt.Sprintf("anonymous_%s_%d", typ.Kind, typ.Id)
}

func SplitFn(sep string) func(resource string) (method, path string) {
	return func(resource string) (method, path string) {
		sp := strings.SplitN(resource, sep, 2)
		if len(sp) > 1 {
			return sp[0], sp[1]
		}
		return sp[0], ""
	}
}

func (p *Parser) parseApi(rt reflect.Type, spec map[string]interface{}) (api Api) {
	api.Name, _ = spec["name"].(string)
	fm, ok := rt.MethodByName(api.Name)
	if !ok {
		return
	}
	api.Doc = p.getFieldDoc(rt, api.Name)
	api.Param, api.Result = p.parseFuncPayload(fm.Type)
	api.Resource, _ = spec["resource"].(string)
	api.Options, _ = spec["options"].(map[string]string)
	api.Invoke, _ = spec["invoke"].(InvokeFunc)
	return
}

func (p *Parser) parseApiGroup(handler interface{}, specs []map[string]interface{}) ApiGroup {
	rt := indirectType(reflect.TypeOf(handler))
	group := ApiGroup{
		Handler: handler,
		Package: rt.PkgPath(),
		Name:    rt.Name(),
		Doc:     p.getFieldDoc(rt, ""),
		Apis:    make([]Api, 0, len(specs)),
	}
	for _, spec := range specs {
		if api := p.parseApi(rt, spec); len(api.Name) > 0 {
			group.Apis = append(group.Apis, api)
		}
	}
	return group
}

func (p *Parser) Parse(iterator Iterator) (groups []ApiGroup, payloads map[int]PayloadType) {
	iterator.Iterate(func(handler interface{}, specs []map[string]interface{}) {
		groups = append(groups, p.parseApiGroup(handler, specs))
	})

	payloads = make(map[int]PayloadType, len(p.types))
	for _, typ := range p.types {
		payloads[typ.Id] = *typ
	}
	return
}
