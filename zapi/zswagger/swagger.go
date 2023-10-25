package zswagger

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"

	"github.com/go-zing/gozz-kit/zapi"
)

var escapeReplacer = strings.NewReplacer("/", "_")

func escape(str string) string { return escapeReplacer.Replace(str) }

type schemaParser struct {
	types       map[int]zapi.PayloadType
	definitions spec.Definitions
}

func Parse(groups []zapi.ApiGroup, types map[int]zapi.PayloadType, cast func(api zapi.Api) zapi.HttpApi) (swagger *spec.Swagger) {
	swagger = &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Info:        &spec.Info{},
			Schemes:     []string{"http"},
			Swagger:     "2.0",
			Paths:       &spec.Paths{Paths: make(map[string]spec.PathItem)},
			Definitions: make(map[string]spec.Schema),
		},
	}

	parser := schemaParser{types: types, definitions: swagger.Definitions}

	setOperation := func(method, path string, operate spec.Operation) {
		item := swagger.Paths.Paths[path]
		fieldName := strings.Title(strings.ToLower(method))
		if v := reflect.ValueOf(&item.PathItemProps).Elem().FieldByName(fieldName); v.IsValid() {
			v.Set(reflect.ValueOf(&operate))
			swagger.Paths.Paths[path] = item
		}
	}

	parseParams := func(api *zapi.HttpApi) (params []spec.Parameter) {
		for _, param := range api.PathParams {
			p := spec.PathParam(param)
			params = append(params, *p)
		}
		return
	}

	parseResponse := func(api *zapi.HttpApi) (response *spec.Schema) {
		if api.Result >= 0 {
			response = parser.Parse(types[api.Result])
		}
		return
	}

	parseOperation := func(group *zapi.ApiGroup, api *zapi.HttpApi) (o spec.Operation) {
		o.ID = group.Fullname() + "." + api.Name
		o.Tags = append(o.Tags, group.Fullname())
		o.Parameters = parseParams(api)
		o.RespondsWith(http.StatusOK, spec.NewResponse().WithSchema(parseResponse(api)))
		return
	}

	for _, group := range groups {
		swagger.Tags = append(swagger.Tags,
			spec.NewTag(group.Fullname(), group.Doc, nil),
		)

		for _, api := range group.Apis {
			h := cast(api)
			h.Method = strings.ToUpper(h.Method)
			h.Path = "/" + strings.TrimPrefix(h.Path, "/")
			setOperation(h.Method, h.Path, parseOperation(&group, &h))
		}
	}
	return
}

func refSchema(name string) *spec.Schema {
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Ref: spec.MustCreateRef("#/definitions/" + name),
		},
	}
}

func (p *schemaParser) Parse(typ zapi.PayloadType) (schema *spec.Schema) {
	name := escape(typ.Fullname())
	if _, ok := p.definitions[name]; ok {
		return refSchema(name)
	}
	return p.parseTypeSchema(typ)
}

func (p *schemaParser) parseEmbedProperties(ele zapi.PayloadElement, schema *spec.Schema) {
	embed := p.Parse(p.types[ele.Type])

	if ref := embed.Ref.String(); len(ref) > 0 {
		v, _ := p.definitions[strings.TrimPrefix(ref, "#/definitions/")]
		embed = &v
	}

	required := make(map[string]bool)
	for _, req := range embed.Required {
		required[req] = true
	}

	for k, v := range embed.Properties {
		if _, exist := schema.Properties[k]; !exist {
			if schema.SetProperty(k, v); !ele.IsPointer() && required[k] {
				schema.AddRequired(k)
			}
		}
	}
}

func (p *schemaParser) parseElementProperty(ele zapi.PayloadElement, schema *spec.Schema) {
	tag := ele.Tags.Get("json")
	key := strings.Split(tag, ",")[0]
	typ := p.types[ele.Type]
	embedded := ele.IsAnonymous() && typ.Kind == "struct"

	if key == "-" {
		return
	} else if embedded && len(key) == 0 {
		p.parseEmbedProperties(ele, schema)
		return
	} else if !embedded && ele.IsUnexported() {
		return
	} else if len(key) == 0 {
		key = ele.Name
	}

	property := p.Parse(typ)
	property.WithDescription(ele.Doc)
	property.AddExtension("x-order", strconv.Itoa(len(schema.Properties)))

	if ref := property.Ref; len(ref.String()) > 0 {
		property.Ref = spec.Ref{}
		property.AddToAllOf(spec.Schema{SchemaProps: spec.SchemaProps{Ref: ref}})
	}

	if schema.SetProperty(key, *property); !ele.IsPointer() {
		schema.AddRequired(key)
	}
}

func isStandardPackage(pkg string) bool {
	return !strings.Contains(strings.SplitN(pkg, "/", 2)[0], ".")
}

func (p *schemaParser) parseTypeSchema(typ zapi.PayloadType) (schema *spec.Schema) {
	if schema = new(spec.Schema); len(typ.Package) > 0 && !isStandardPackage(typ.Package) {
		name := escape(typ.Fullname())
		p.definitions[name] = *schema
		defer func() {
			p.definitions[name] = *schema
			*schema = *refSchema(name)
		}()
	}

	schema.WithExample(typ.Entity)
	schema.WithDescription(typ.Doc)

	switch typ.Kind {
	case "interface":
		schema.Nullable = true
	case "struct":
		schema.Typed("object", "")
		for _, ele := range typ.Elements {
			p.parseElementProperty(ele, schema)
		}
	case "map":
		schema.Typed("object", "")
		schema.SetProperty(".*", *p.Parse(p.types[typ.Elements[0].Type]))
	case "array", "slice":
		schema.Typed("array", "")
		schema.Items = &spec.SchemaOrArray{Schema: p.Parse(p.types[typ.Elements[0].Type])}
	case "bool":
		schema.Typed("boolean", "")
	case "string":
		schema.Typed("string", "")
	case "float32", "float64":
		schema.Typed("number", typ.Kind)
	}

	if kind := typ.Kind; strings.Contains(kind, "int") {
		schema.Typed("integer", kind)
		max := uintptr(1)
		min := float64(-1)
		if strings.HasPrefix(kind, "u") {
			kind = kind[1:]
			max <<= 1
			min = 0
		}
		size := strconv.IntSize
		if ss := strings.TrimPrefix(kind, "int"); len(ss) > 0 {
			size, _ = strconv.Atoi(ss)
		} else {
			schema.Format += strconv.Itoa(size)
		}
		max = max<<(size-1) - 1
		schema.WithMaximum(float64(max), false)
		schema.WithMinimum(min*float64(max), false)
	}
	return
}
