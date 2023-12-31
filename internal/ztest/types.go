package ztest

import (
	"context"
)

//go:generate gozz run -p "doc" -p "api" -p "tag" ./

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
		// embedded field
		Embedded
		// string field
		// multi line doc
		String string `form:"string" json:"string"`
		// int field
		Int int `form:"int" json:"int"`
		// int64 field
		Int64 int `form:"int64" json:"int64"`
		// uint field
		Uint uint `form:"uint" json:"uint"`
		// uint16 field
		Uint16 uint16 `form:"uint16" json:"uint16"`
		// int16 field
		Int16 int16 `form:"int16" json:"int16"`
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
