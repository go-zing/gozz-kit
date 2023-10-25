package ztest

import (
	"context"

	"github.com/go-zing/gozz-kit/zdoc"
)

//go:generate gozz run -p "doc" -p "api" -p "tag" ./

var Docs = zdoc.Docs{}

func init() {
	Docs.LoadTypes(_types_doc)
}

// +zz:doc
// +zz:tag:json,form:{{ snake .FieldName }}
type (
	// +zz:api:./
	// api interface
	Api interface {
		// +zz:api:get|{id}
		// get method
		Get(ctx context.Context, query Payload) (Payload, error)
		// +zz:api:post
		// post method
		Post(ctx context.Context, form Payload) (Payload, error)
	}

	// query payload
	Payload struct {
		Embedded
		// string field
		String string `form:"string" json:"string"`
		// int field
		Int int `form:"int" json:"int"`
		// uint field
		Uint uint `form:"uint" json:"uint"`
		// map field
		Map map[string]Payload `form:"map" json:"map"`
		// map field with pointer element
		MapPointer map[string]*Payload `form:"map_pointer" json:"map_pointer"`
		// slice field
		Slice []Payload `form:"slice" json:"slice"`
		// slice field with pointer element
		SlicePointer []*Payload `form:"slice_pointer" json:"slice_pointer"`
		// anonymous struct field
		Struct struct {
			String string `form:"string" json:"string"`
		} `form:"struct" json:"struct"`
		// anonymous struct field with pointer element
		StructPointer *struct {
			String string `form:"string" json:"string"`
		} `form:"struct_pointer" json:"struct_pointer"`
		// bytes field
		Bytes []byte `form:"bytes" json:"bytes"`

		unexportedEmbedded

		pointerUnexportedEmbedded

		// refer struct
		Refer Embedded `form:"refer" json:"refer"`

		// exported recursive
		Exported *Payload `form:"exported" json:"exported"`

		// unexported recursive
		unexported *Payload
	}

	unexportedEmbedded struct {
		String3 string `form:"string3" json:"string3"`
	}

	pointerUnexportedEmbedded struct {
		String4 string `form:"string4" json:"string4"`
	}

	Embedded struct {
		Sting2 string `form:"sting2" json:"sting2"`
		// struct field with pointer element
		Struct2 *Payload `form:"struct2" json:"struct2"`
		// anonymous field with pointer element
		*Payload
	}
)
