package ztree

import (
	"io/ioutil"
	"testing"
)

type D struct {
	V *string
	s
}

type s interface{}

func TestParse(t *testing.T) {
	str := ""
	ptr := &str
	d1 := &D{
		V: ptr,
		s: ptr,
	}
	s := D{
		V: ptr,
		s: ptr,
	}
	tree := Parse([]interface{}{d1, s}, WithUnexported(true))

	t.Log(d1.s == d1.V)
	bs := (&Drawer{}).Draw(tree)
	_ = ioutil.WriteFile("test.dot", bs, 0o644)
}
