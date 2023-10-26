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

	"github.com/go-zing/gozz-kit/internal/helpers"
	"github.com/go-zing/gozz-kit/zapi"
)

var escapeReplacer = strings.NewReplacer("/", "_")

func escape(str string) string { return escapeReplacer.Replace(str) }

type schemaParser struct {
	option      Option
	payloads    map[reflect.Type]zapi.PayloadType
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

//go:generate gozz run -p "option" .
// +zz:option
type Option struct {
	HttpCast func(api zapi.Api) zapi.HttpApi
	Bindings map[string]Binding
}

type Binding struct {
	Path   string
	Query  string
	Header string
	Body   bool
}

func parseBinding(api *zapi.HttpApi, rules map[string]Binding) Binding {
	rule, ok := rules[strings.ToUpper(api.Method)]
	if !ok {
		return rules["*"]
	}
	rv := reflect.ValueOf(&rule).Elem()
	for i := 0; i < rv.NumField(); i++ {
		v, valid := rv.Field(i).Addr().Interface().(*string)
		if !valid {
			continue
		}
		if ref, exist := rules[strings.ToUpper(*v)]; exist {
			*v = reflect.ValueOf(ref).Field(i).String()
		}
	}
	return rule
}

func parseElements(payloads map[reflect.Type]zapi.PayloadType, root zapi.PayloadType, tag string, fn func(zapi.PayloadElement, zapi.TagValues)) {
	if len(tag) == 0 {
		return
	}
	delete(payloads, root.Type)
	for _, ele := range root.Elements {
		typ, ok := payloads[ele.Type]
		if !ok {
			continue
		}
		values := ele.Tags.Get(tag).Split(",")
		value := values[0]
		if value == "-" {
			continue
		} else if ele.IsAnonymous() && len(value) == 0 && typ.Kind == reflect.Struct {
			parseElements(payloads, typ, tag, fn)
		} else {
			fn(ele, values)
		}
	}
}

func setOperation(paths map[string]spec.PathItem, method, path string, operate spec.Operation) {
	item := paths[path]
	fieldName := strings.Title(strings.ToLower(method))
	if v := reflect.ValueOf(&item).Elem().FieldByName(fieldName); v.IsValid() {
		v.Set(reflect.ValueOf(&operate))
		paths[path] = item
	}
}

func Parse(groups []zapi.ApiGroup, payloads map[reflect.Type]zapi.PayloadType, option ...func(*Option)) (swagger *spec.Swagger) {
	swagger = &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Info:        &spec.Info{},
			Schemes:     []string{"http"},
			Swagger:     "2.0",
			Paths:       &spec.Paths{Paths: make(map[string]spec.PathItem)},
			Definitions: make(map[string]spec.Schema),
		},
	}

	parser := &schemaParser{payloads: payloads, definitions: swagger.Definitions}
	parser.option.applyOptions(option...)

	for _, group := range groups {
		swagger.Tags = append(swagger.Tags, spec.NewTag(group.Fullname(), group.Doc, nil))
		for _, api := range group.Apis {
			h := parser.httpCast(api)
			h.Method = strings.ToUpper(h.Method)
			h.Path = "/" + strings.TrimPrefix(h.Path, "/")
			setOperation(swagger.Paths.Paths, h.Method, h.Path, parser.parseOperation(&group, &h))
		}
	}
	return
}

func parseParam(param *spec.Parameter, element zapi.PayloadElement, values zapi.TagValues) {
	typ, format, max, min := parseBasicKind(element.Type.Kind())
	if param.Typed(typ, format); max > 0 {
		param.WithMaximum(max, false)
		param.WithMinimum(min, false)
	}
}

func (p *schemaParser) httpCast(api zapi.Api) (h zapi.HttpApi) {
	if p.option.HttpCast != nil {
		return p.option.HttpCast(api)
	}
	return
}

func (p *schemaParser) parseBinding(h *zapi.HttpApi) Binding {
	return parseBinding(h, p.option.Bindings)
}

func (p *schemaParser) parseOperation(group *zapi.ApiGroup, api *zapi.HttpApi) (o spec.Operation) {
	// meta info
	o.ID = group.Fullname() + "." + api.Name
	o.Description = api.Doc
	o.Tags = append(o.Tags, group.Fullname())
	// parse params
	o.Parameters = p.parseParams(api, p.parseBinding(api))
	// parse response
	if api.Response != nil {
		schema := p.Parse(p.payloads[api.Response])
		o.RespondsWith(http.StatusOK, spec.NewResponse().WithSchema(&schema))
	}
	return
}

func (p *schemaParser) parseParams(api *zapi.HttpApi, binding Binding) (params []spec.Parameter) {
	added := make(map[string]int)

	addParam := func(param *spec.Parameter) {
		if _, exist := added[param.Name]; !exist && len(param.Type) > 0 && len(param.Name) > 0 {
			added[param.Name] = len(params)
			params = append(params, *param)
		}
	}

	for _, path := range strings.Split(api.Path, "/") {
		if strings.HasPrefix(path, "{") && strings.HasSuffix(path, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(path, "{"), "}")
			addParam(spec.PathParam(name).Typed("string", ""))
		}
	}

	if api.Request != nil && api.Request.Kind() == reflect.Struct {
		parsePayload := func(tag string, fn func(element zapi.PayloadElement, values zapi.TagValues)) {
			cp := make(map[reflect.Type]zapi.PayloadType, len(p.payloads))
			for k, v := range p.payloads {
				cp[k] = v
			}
			parseElements(cp, cp[api.Request], tag, func(element zapi.PayloadElement, values zapi.TagValues) {
				if _, exist := added[values[0]]; !exist && len(values[0]) > 0 {
					fn(element, values)
				}
			})
		}

		newWith := func(in string) func(element zapi.PayloadElement, values zapi.TagValues) {
			return func(element zapi.PayloadElement, values zapi.TagValues) {
				param := &spec.Parameter{ParamProps: spec.ParamProps{Name: values[0], In: in}}
				parseParam(param, element, values)
				addParam(param)
			}
		}

		parsePayload(binding.Path, func(element zapi.PayloadElement, values zapi.TagValues) {
			if index, ok := added[values[0]]; ok {
				parseParam(&params[index], element, values)
			}
		})

		parsePayload(binding.Query, newWith("query"))
		parsePayload(binding.Header, newWith("header"))
	}

	if binding.Body {
		schema := p.Parse(p.payloads[api.Request])
		params = append(params, *spec.BodyParam("", &schema).Named("body"))
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
	embed := p.Parse(p.payloads[ele.Type])
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
	key := values[0]
	typ := p.payloads[ele.Type]

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
	addElementProperty(schema, &property, key, !ele.IsPointer() && !values.Exist("omitempty"))
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
		schema.SetProperty(".*", p.Parse(p.payloads[typ.Elements[0].Type]))
	case reflect.Slice, reflect.Array:
		schema.Typed("array", "")
		itemSchema := p.Parse(p.payloads[typ.Elements[0].Type])
		schema.Items = &spec.SchemaOrArray{Schema: &itemSchema}
	}

	if t, format, max, min := parseBasicKind(typ.Kind); len(t) > 0 {
		if schema.Typed(t, format); max > 0 {
			schema.WithMaximum(max, false)
			schema.WithMinimum(min, false)
		}
	}
	return
}

func parseBasicKind(k reflect.Kind) (typ, format string, max, min float64) {
	kind := k.String()

	switch k {
	case reflect.Bool:
		return "boolean", "", 0, 0
	case reflect.String:
		return "string", "", 0, 0
	case reflect.Float32:
		return "number", kind, 0, 0
	case reflect.Float64:
		return "number", kind, 0, 0
	}

	if strings.Contains(kind, "int") {
		typ = "integer"
		format = kind
		unsigned := helpers.Btoi[strings.HasPrefix(kind, "u")]
		size := strconv.IntSize
		if ss := strings.TrimPrefix(kind[unsigned:], "int"); len(ss) > 0 {
			size, _ = strconv.Atoi(ss)
		} else {
			format += strconv.Itoa(size)
		}
		max = float64(uintptr(1<<unsigned)<<(size-1) - 1)
		min = max * float64(unsigned-1)
	}
	return
}
