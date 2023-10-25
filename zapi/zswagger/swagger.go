package zswagger

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/spec"

	"github.com/go-zing/gozz-kit/zapi"
)

var escapeReplacer = strings.NewReplacer("/", "_")

func escape(str string) string { return escapeReplacer.Replace(str) }

type schemaParser struct {
	types       map[reflect.Type]zapi.PayloadType
	definitions spec.Definitions
}

var defined = map[reflect.Type]func(*spec.Schema){
	reflect.TypeOf(net.IP{}):          func(schema *spec.Schema) { schema.Typed("string", "ipv4") },
	reflect.TypeOf(time.Time{}):       func(schema *spec.Schema) { schema.Typed("string", "date-time") },
	reflect.TypeOf(url.URL{}):         func(schema *spec.Schema) { schema.Typed("string", "uri") },
	reflect.TypeOf([]byte(nil)):       func(schema *spec.Schema) { schema.Typed("string", "base64") },
	reflect.TypeOf(json.RawMessage{}): func(schema *spec.Schema) { schema.Typed("object", "") },
	reflect.TypeOf(struct{}{}):        func(schema *spec.Schema) { schema.Typed("null", "") },
}

func RegisterSchemaType(typ reflect.Type, fn func(*spec.Schema)) { defined[typ] = fn }

func Parse(groups []zapi.ApiGroup, types map[reflect.Type]zapi.PayloadType, cast func(api zapi.Api) zapi.HttpApi) (swagger *spec.Swagger) {
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
		if api.Response != nil {
			schema := parser.Parse(types[api.Response])
			response = &schema
		}
		return
	}

	parseOperation := func(group *zapi.ApiGroup, api *zapi.HttpApi) (o spec.Operation) {
		o.ID = group.Fullname() + "." + api.Name
		o.Description = api.Doc
		o.Tags = append(o.Tags, group.Fullname())
		o.Parameters = parseParams(api)
		o.RespondsWith(http.StatusOK, spec.NewResponse().WithSchema(parseResponse(api)))
		return
	}

	for _, group := range groups {
		swagger.Tags = append(swagger.Tags, spec.NewTag(group.Fullname(), group.Doc, nil))

		for _, api := range group.Apis {
			h := cast(api)
			h.Method = strings.ToUpper(h.Method)
			h.Path = "/" + strings.TrimPrefix(h.Path, "/")
			setOperation(h.Method, h.Path, parseOperation(&group, &h))
		}
	}
	return
}

const (
	definitionsPrefix = "#/definitions/"
	extensionKeyOrder = "x-order"
)

func refSchema(name string) spec.Schema {
	return spec.Schema{
		SchemaProps: spec.SchemaProps{
			Ref: spec.MustCreateRef(definitionsPrefix + name),
		},
	}
}

func (p *schemaParser) Parse(typ zapi.PayloadType) (schema spec.Schema) {
	name := escape(typ.Fullname())
	if _, ok := p.definitions[name]; ok {
		return refSchema(name)
	}
	return p.parseTypeSchema(typ)
}

func (p *schemaParser) parseEmbedProperties(ele zapi.PayloadElement, schema *spec.Schema) {
	embed := p.Parse(p.types[ele.Type])
	if ref := embed.Ref.String(); len(ref) > 0 {
		embed = p.definitions[strings.TrimPrefix(ref, definitionsPrefix)]
	}
	required := make(map[string]bool, len(embed.Required))
	for _, req := range embed.Required {
		required[req] = !ele.IsPointer()
	}
	keys := make([]string, 0, len(embed.Properties))
	for key := range embed.Properties {
		if _, exist := schema.Properties[key]; !exist {
			keys = append(keys, key)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		orderI, _ := embed.Properties[keys[i]].Extensions.GetInt(extensionKeyOrder)
		orderJ, _ := embed.Properties[keys[j]].Extensions.GetInt(extensionKeyOrder)
		return orderI < orderJ
	})
	for _, key := range keys {
		property := embed.Properties[key]
		addElementProperty(schema, &property, key, required[key])
	}
}

func addElementProperty(dst, property *spec.Schema, key string, required bool) {
	property.AddExtension(extensionKeyOrder, strconv.Itoa(len(dst.Properties)))
	if dst.SetProperty(key, *property); required {
		dst.AddRequired(key)
	}
}

func (p *schemaParser) parseElementProperty(ele zapi.PayloadElement, schema *spec.Schema) {
	values := ele.Tags.Get("json").Split(",")
	omitempty := values.Exist("omitempty")
	key := values[0]
	typ := p.types[ele.Type]

	if key == "-" {
		return
	} else if ele.IsAnonymous() && typ.Kind == reflect.Struct && len(key) == 0 {
		p.parseEmbedProperties(ele, schema)
		return
	} else if len(key) == 0 {
		key = ele.Name
	}

	property := p.Parse(typ)
	setSchemaDoc(&property, ele.Doc)
	if ref := property.Ref; len(ref.String()) > 0 {
		property.Ref = spec.Ref{}
		property.AddToAllOf(spec.Schema{SchemaProps: spec.SchemaProps{Ref: ref}})
	}
	addElementProperty(schema, &property, key, !ele.IsPointer() && !omitempty)
}

func isStandardPackage(pkg string) bool {
	return !strings.Contains(strings.SplitN(pkg, "/", 2)[0], ".")
}

func setSchemaDoc(schema *spec.Schema, doc string) {
	sp := strings.SplitN(doc, "\n", 2)[:2]
	schema.WithTitle(strings.TrimSpace(sp[0]))
	schema.WithDescription(strings.TrimSpace(sp[1]))
}

func (p *schemaParser) parseTypeSchema(typ zapi.PayloadType) (schema spec.Schema) {
	if len(typ.Package) > 0 && !isStandardPackage(typ.Package) {
		name := escape(typ.Fullname())
		p.definitions[name] = schema
		defer func() {
			p.definitions[name] = schema
			schema = refSchema(name)
		}()
	}

	setSchemaDoc(&schema, typ.Doc)

	if define, ok := defined[typ.Type]; ok {
		if define(&schema); len(schema.Type) > 0 {
			return
		}
	}

	schema.Example = typ.Entity
	kind := typ.Kind.String()

	switch typ.Kind {
	case reflect.Interface:
		schema.Nullable = true
	case reflect.Struct:
		schema.Typed("object", "")
		for _, ele := range typ.Elements {
			p.parseElementProperty(ele, &schema)
		}
	case reflect.Map:
		schema.Typed("object", "")
		schema.SetProperty(".*", p.Parse(p.types[typ.Elements[0].Type]))
	case reflect.Slice, reflect.Array:
		schema.Typed("array", "")
		itemSchema := p.Parse(p.types[typ.Elements[0].Type])
		schema.Items = &spec.SchemaOrArray{Schema: &itemSchema}
	case reflect.Bool:
		schema.Typed("boolean", "")
	case reflect.String:
		schema.Typed("string", "")
	case reflect.Float32, reflect.Float64:
		schema.Typed("number", kind)
	}

	if strings.Contains(kind, "int") {
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
