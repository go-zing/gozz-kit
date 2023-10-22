package ztree

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

var want = []byte(`{"values":[{"id":"0","type":"0","referred":2,"elements":{"0":"1","1":"6"}},{"id":"1","type":"1","referred":0,"elements":{"":"2"}},{"id":"2","type":"2","referred":3,"elements":{"Buff":"4","Field":"3","_field":"5"}},{"id":"3","type":"3","referred":5,"elements":{}},{"id":"4","type":"4","referred":0,"elements":{}},{"id":"5","type":"6","referred":0,"elements":{"":"3"}},{"id":"6","type":"1","referred":0,"elements":{"":"7"}},{"id":"7","type":"2","referred":3,"elements":{"Buff":"8","Field":"3","_field":"9"}},{"id":"8","type":"4","referred":0,"elements":{}},{"id":"9","type":"6","referred":0,"elements":{"":"10"}},{"id":"10","type":"3","referred":3,"elements":{}}],"types":[{"id":"0","kind":"slice","package":"","name":"","string":"[]interface{}","elements":{"":"1"},"anonymous":{},"docs":{"":""}},{"id":"1","kind":"interface","package":"","name":"","string":"interface{}","elements":{},"anonymous":{},"docs":{"":""}},{"id":"2","kind":"struct","package":"github.com/go-zing/gozz-kit/ztree","name":"_T0","string":"ztree._T0","elements":{"Buff":"4","Field":"3","_field":"6"},"anonymous":{"_field":true},"docs":{"":"type for testing","Buff":"export field with interface reader type","Field":"export field with pointer struct type","_field":"unexported field interface"}},{"id":"3","kind":"struct","package":"bytes","name":"Buffer","string":"bytes.Buffer","elements":{},"anonymous":{},"docs":{"":""}},{"id":"4","kind":"interface","package":"io","name":"Reader","string":"io.Reader","elements":{"Read":"5"},"anonymous":{},"docs":{"":"","Read":""}},{"id":"5","kind":"func","package":"","name":"","string":"func([]uint8) (int, error)","elements":{},"anonymous":{},"docs":{"":""}},{"id":"6","kind":"interface","package":"github.com/go-zing/gozz-kit/ztree","name":"_field","string":"ztree._field","elements":{"Read":"5"},"anonymous":{},"docs":{"":"field interface","Read":"read function"}}]}`)

func TestParse(t *testing.T) {
	ptr := new(bytes.Buffer)
	doc := make(map[reflect.Type]map[string]string)
	for k, v := range _types_doc {
		doc[reflect.TypeOf(k).Elem()] = v
	}

	b, err := json.Marshal(Parse([]interface{}{&_T0{
		Field:  ptr,
		_field: ptr,
	}, _T0{
		Field:  ptr,
		_field: ptr,
	}}, WithDocFunc(func(p reflect.Type, field string) string { return doc[p][field] })))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(b, want) {
		t.Fatalf("%s", b)
	}
}

func TestDraw(t *testing.T) {
	tree := Tree{}
	err := json.Unmarshal(want, &tree)
	if err != nil {
		t.Fatal()
	}
	t.Logf("%s", Draw(tree))
}
