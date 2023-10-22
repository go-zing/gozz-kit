package ztree

import (
	"bytes"
	"fmt"
	"go/format"
	"sort"
	"strconv"
	"strings"
)

const (
	keyLabel        = "label"
	keyTooltip      = "tooltip"
	keyLabelTooltip = keyLabel + keyTooltip

	typeInterface = "interface"
	typeStruct    = "struct"
	typeMap       = "map"
	typeArray     = "array"
	typeSlice     = "slice"
	typeFunc      = "func"

	colorScheme = "pastel19"
)

var (
	kindColor = map[string]string{
		typeInterface: "1",
		typeStruct:    "2",
		typeMap:       "3",
		typeSlice:     "4",
		typeArray:     "5",
		typeFunc:      "6",
		"":            "7",       // default
		"bg":          "#f6fff6", // struct background
	}

	graphProperties = map[string]map[string]string{
		"node": {
			"shape":       "rect",
			"style":       "filled",
			"margin":      "0.05,0",
			"height":      "0.25",
			"colorscheme": colorScheme,
		},
		"edge": {
			"arrowsize": "0.6",
			"fontsize":  "10",
		},
		"graph": {
			"style":  "dotted",
			"margin": "0,0",
		},
	}
)

type drawer struct {
	builder *bytes.Buffer
	*Tree
}

func (d *drawer) color(kind string) string {
	color, set := kindColor[kind]
	if !set {
		color = kindColor[""]
	}
	return color
}

func (d *drawer) writeValueNode(max int, v *Value, typ *Type) {
	tooltip := typ.Name
	if len(typ.Name) == 0 {
		tooltip = typ.String
	} else if len(typ.Package) > 0 {
		tooltip = typ.Package + "." + typ.Name
	}

	if doc := typ.Docs[""]; len(doc) > 0 {
		tooltip += "\n" + doc
	}

	d.writeNode(v.Id, map[string]string{
		keyLabel:    typ.String,
		keyTooltip:  tooltip,
		"fillcolor": d.color(typ.Kind),
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

func Draw(name string, tree Tree) []byte {
	return (&drawer{builder: &bytes.Buffer{}, Tree: &tree}).Draw(name)
}

type mergeElements struct {
	Len int
	Id  string
}

func (d *drawer) maxReferred() int {
	return d.Tree.maxReferred()
}

func (d *drawer) writeValue(value *Value, types map[string]*Type, patch map[string]mergeElements) {
	valueType := types[value.Type]

	elements, merged := patch[value.Id]
	if !merged || elements.Id == value.Id {
		value.Referred += elements.Len
		d.writeValueNode(d.maxReferred(), value, valueType)
	} else {
		value.Id = elements.Id
	}

	switch valueType.Kind {
	case typeMap, typeSlice, typeArray:
		le := len(value.Elements)
		if elementType := types[valueType.Elements[""]]; elementType != nil && elementType.Kind == typeInterface && le > 1 {
			for key, element := range value.Elements {
				if key == "0" {
					d.writeEdge(value.Id, element, map[string]string{
						keyLabel:        "elements",
						keyLabelTooltip: "total: " + strconv.Itoa(le),
					})
				}
				patch[element] = mergeElements{Len: le, Id: value.Elements["0"]}
			}
		} else {
			d.writeElementsEdge(value, func(name string) map[string]string {
				return map[string]string{
					keyLabel:        "element",
					keyLabelTooltip: "index: " + name,
				}
			})
		}
	case typeStruct:
		_, _ = fmt.Fprintf(d.builder, "subgraph cluster_%s {\n", value.Id)
		d.writeProperty(map[string]string{
			keyTooltip: structDefine(valueType, types),
			"bgcolor":  kindColor["bg"],
		})
		d.writeElementsEdge(value, func(name string) map[string]string {
			tooltip := name
			if doc := valueType.Docs[name]; len(doc) > 0 {
				tooltip += ": " + doc
			}
			return map[string]string{
				keyLabel:        name,
				keyLabelTooltip: tooltip,
				"arrowhead":     "open",
			}
		})
		d.builder.WriteString("}\n")
	case typeInterface:
		attrs := map[string]string{
			keyLabel:        "implement",
			keyLabelTooltip: interfaceDefine(valueType, types),
			"dir":           "back",
			"arrowtail":     "onormal",
		}
		if !merged {
			attrs["style"] = "dashed"
		}
		d.writeElementsEdge(value, func(string) map[string]string { return attrs })
	}
}

func (d *drawer) Draw(name string) []byte {
	patch := make(map[string]mergeElements)
	types := d.Tree.typesMap()
	d.builder.WriteString("digraph ")
	d.builder.WriteString(name)
	d.builder.WriteString(" {\n")
	for _, str := range []string{"node", "edge", "graph"} {
		d.writeNode(str, graphProperties[str])
	}
	for index := range d.Tree.Values {
		d.writeValue(&d.Tree.Values[index], types, patch)
	}
	d.builder.WriteRune('}')
	return d.builder.Bytes()
}

func interfaceDefine(valueType *Type, types map[string]*Type) string {
	str := &strings.Builder{}
	rangeMap(valueType.Elements, func(key, value string, index int) {
		str.WriteString(strings.Replace(types[value].String, "func", "func "+key, 1))
		if doc := valueType.Docs[key]; len(doc) > 0 {
			str.WriteString("\n")
			str.WriteString(doc)
		}
		if index != len(valueType.Elements)-1 {
			str.WriteString("\n")
		}
	})
	if str.Len() == 0 {
		str.WriteString("any")
	}
	return str.String()
}

func structDefine(typ *Type, types map[string]*Type) string {
	str := &bytes.Buffer{}
	_, _ = fmt.Fprintf(str, "type %s struct {\n", typ.Name)
	rangeMap(typ.Elements, func(key, value string, index int) {
		if elementType, ok := types[value]; ok {
			if !typ.Anonymous[key] {
				_, _ = fmt.Fprintf(str, key)
			}
			_, _ = fmt.Fprintf(str, " %s\n", elementType.String)
		}
	})
	_, _ = fmt.Fprintf(str, "}")
	b, _ := format.Source(str.Bytes())
	return strings.Replace(string(b), "\t", "    ", -1)
}

func (d *drawer) writeProperty(properties map[string]string) {
	rangeMap(properties, func(key, value string, i int) {
		_, _ = fmt.Fprint(d.builder, key, "=", strconv.Quote(value))
		if i != len(properties)-1 {
			_, _ = fmt.Fprintf(d.builder, " ")
		}
	})
}

func (d *drawer) writeNode(name string, properties map[string]string) {
	d.builder.WriteString(name)
	d.builder.WriteString(" [")
	d.writeProperty(properties)
	d.builder.WriteString("];\n")
}

func (d *drawer) writeEdge(src, dst string, properties map[string]string) {
	_, _ = fmt.Fprint(d.builder, src, " -> ", dst, " [")
	d.writeProperty(properties)
	d.builder.WriteString("];\n")
}
