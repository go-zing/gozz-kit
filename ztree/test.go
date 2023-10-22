package ztree

import (
	"bytes"
	"io"
)

//go:generate gozz run -p "doc" ./${GOFILE}
// +zz:doc
type (
	// type for testing
	_T0 struct {
		// export field with pointer struct type
		Field *bytes.Buffer
		// export field with interface reader type
		Buff io.Reader
		// unexported field interface
		_field
	}

	// field interface
	_field interface {
		// read function
		Read([]byte) (int, error)
	}
)
