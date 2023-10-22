package ztree

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"sort"
	"strconv"
	"strings"
)

var (
	kindColor = map[string]string{
		"interface": "#A6E7FF",
		"struct":    "#CCFF00",
		"map":       "#FFFF38",
		"slice":     "#FBAED2",
		"array":     "#FBAED2",
		"func":      "#FF9933",
		"":          "#FBE7B2",
	}
)

type drawer struct {
	builder *bytes.Buffer
}

func (d *drawer) color(kind string) string {
	color, set := kindColor[kind]
	if !set {
		color = kindColor[""]
	}
	return color
}

func (d *drawer) writeValueNode(max int, v *Value, typ *Type) {
	tip := typ.Name
	if len(typ.Name) == 0 {
		tip = typ.String
	} else if len(typ.Package) > 0 {
		tip = typ.Package + "." + typ.Name
	}

	if doc := typ.Docs[""]; len(doc) > 0 {
		tip += `\n` + doc
	}

	d.writeNode(v.Id, map[string]string{
		"label":     typ.String,
		"fillcolor": d.color(typ.Kind),
		"margin":    "0.05,0",
		"height":    "0.25",
		"tooltip":   tip,
		"fontsize":  fmt.Sprintf("%.1f", 8+8*float64(v.Referred)/float64(max)),
	})
}

func rangeMap(m map[string]string, fn func(key, value string, index int)) {
	keys := make([]string, 0, len(m))
	for name := range m {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for i, key := range keys {
		fn(key, m[key], i)
	}
}

func (d *drawer) writeElementsEdge(v *Value, fn func(name string) map[string]string) {
	rangeMap(v.Elements, func(key, value string, index int) { d.writeEdge(v.Id, value, fn(key)) })
}

func Draw(tree Tree) []byte {
	return new(drawer).Draw(tree)
}

type mergeElements struct {
	Len int
	Id  string
}

func (d *drawer) Draw(tree Tree) []byte {
	types := make(map[string]*Type)
	for i, ti := range tree.Types {
		types[ti.Id] = &tree.Types[i]
	}

	d.builder = &bytes.Buffer{}
	d.builder.WriteString("digraph name {\nnode [style=filled shape=rect]\n")

	patch := make(map[string]mergeElements)

	max := 0
	for _, v := range tree.Values {
		if v.Referred > max {
			max = v.Referred
		}
	}

	for index := range tree.Values {
		v := &tree.Values[index]
		typ := types[v.Type]

		elements, merged := patch[v.Id]
		if !merged || elements.Id == v.Id {
			v.Referred += elements.Len
			d.writeValueNode(max, v, typ)
		} else {
			v.Id = elements.Id
		}

		labelKey := "label"

		switch typ.Kind {
		case "map", "slice", "array":
			le := len(v.Elements)
			if et := types[typ.Elements[""]]; et != nil && et.Kind == "interface" && le > 1 {
				for key, element := range v.Elements {
					if key == "0" {
						d.writeEdge(v.Id, element, map[string]string{
							labelKey:       "elements",
							"labeltooltip": "total: " + strconv.Itoa(le),
						})
					}
					patch[element] = mergeElements{Len: le, Id: v.Elements["0"]}
				}
			} else {
				d.writeElementsEdge(v, func(name string) map[string]string {
					return map[string]string{
						labelKey:       "element",
						"labeltooltip": "index: " + name,
					}
				})
			}
		case "struct":
			_, _ = fmt.Fprintf(d.builder, "subgraph cluster_%s {\n", v.Id)
			_, _ = fmt.Fprintf(d.builder, "tooltip=%s;\n", strconv.Quote(structDefine(typ, types)))
			_, _ = fmt.Fprintln(d.builder, "style=dotted;")
			_, _ = fmt.Fprintln(d.builder, `bgcolor="#f2fff2";`)
			_, _ = fmt.Fprintln(d.builder, `margin="0,0";`)
			d.writeElementsEdge(v, func(name string) map[string]string {
				tip := name
				if doc := typ.Docs[name]; len(doc) > 0 {
					tip += ": " + doc
				}
				return map[string]string{
					labelKey:       name,
					"arrowhead":    "open",
					"labeltooltip": tip,
				}
			})
			d.builder.WriteString("}\n")
		case "interface":
			str := &strings.Builder{}
			rangeMap(typ.Elements, func(key, value string, index int) {
				str.WriteString(strings.Replace(types[value].String, "func", "func "+key, 1))
				if doc := typ.Docs[key]; len(doc) > 0 {
					str.WriteString(`\n`)
					str.WriteString(doc)
				}
				if index != len(typ.Elements)-1 {
					str.WriteString(`\n`)
				}
			})
			if str.Len() == 0 {
				str.WriteString("any")
			}

			m := map[string]string{
				labelKey:       "implement",
				"arrowhead":    "onormal",
				"labeltooltip": str.String(),
			}
			if !merged {
				m["style"] = "dashed"
			}
			d.writeElementsEdge(v, func(string) map[string]string { return m })
		}
	}

	d.builder.WriteRune('}')
	return d.builder.Bytes()
}

func structDefine(typ *Type, types map[string]*Type) string {
	str := &bytes.Buffer{}
	_, _ = fmt.Fprintf(str, "type %s struct {\n", typ.Name)
	rangeMap(typ.Elements, func(key, value string, index int) {
		if typ.Anonymous[key] {
			_, _ = fmt.Fprintf(str, "%s\n", types[value].String)
		} else {
			_, _ = fmt.Fprintf(str, "%s %s\n", key, types[value].String)
		}
	})
	_, _ = fmt.Fprintf(str, "}")
	b, _ := format.Source(str.Bytes())
	return strings.Replace(string(b), "\t", "    ", -1)
}

func (d *drawer) writeProperty(properties map[string]string) {
	d.builder.WriteRune('[')
	writeMap(properties, "=", `"`, ` `, d.builder)
	d.builder.WriteRune(']')
}

func (d *drawer) writeNode(name string, properties map[string]string) {
	d.builder.WriteString(name)
	d.builder.WriteString(" ")
	d.writeProperty(properties)
	d.builder.WriteString(";\n")
}

func (d *drawer) writeEdge(src, dst string, properties map[string]string) {
	_, _ = fmt.Fprint(d.builder, src, " -> ", dst, " ")
	properties["arrowsize"] = "0.7"
	properties["weight"] = "100"
	properties["fontsize"] = "10"
	d.writeProperty(properties)
	d.builder.WriteString(";\n")
}

func writeMap(kv map[string]string, sep, wrap, join string, writer io.Writer) {
	rangeMap(kv, func(key, value string, i int) {
		_, _ = fmt.Fprint(writer, key, sep, wrap, value, wrap)
		if i != len(kv)-1 {
			_, _ = fmt.Fprintf(writer, join)
		}
	})
}
