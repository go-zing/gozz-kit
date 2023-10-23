package ztree

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

var want = []byte(`{"values":[{"id":"0","type":"0","referred":2,"elements":{"0":"1","1":"6"}},{"id":"1","type":"1","referred":0,"elements":{"":"2"}},{"id":"2","type":"2","referred":4,"elements":{"Buff":"4","Field":"3","_field":"5"}},{"id":"3","type":"3","referred":2,"elements":{}},{"id":"4","type":"8","referred":0,"elements":{}},{"id":"5","type":"10","referred":0,"elements":{"":"3"}},{"id":"6","type":"1","referred":0,"elements":{"":"7"}},{"id":"7","type":"2","referred":4,"elements":{"Buff":"8","Field":"3","_field":"9"}},{"id":"8","type":"8","referred":0,"elements":{}},{"id":"9","type":"10","referred":0,"elements":{"":"10"}},{"id":"10","type":"3","referred":0,"elements":{}}],"types":[{"id":"0","kind":"slice","package":"","name":"","string":"[]interface{}","elements":{"":"1"},"anonymous":{},"docs":{"":""}},{"id":"1","kind":"interface","package":"","name":"","string":"interface{}","elements":{},"anonymous":{},"docs":{"":""}},{"id":"2","kind":"struct","package":"github.com/go-zing/gozz-kit/ztree","name":"_T0","string":"ztree._T0","elements":{"Buff":"8","Field":"3","X":"3","_field":"10"},"anonymous":{"Buff":false,"Field":false,"X":false,"_field":true},"docs":{"":"type for testing","Buff":"export field with interface reader type","Field":"export field with pointer struct type","X":"","_field":"unexported field interface"}},{"id":"3","kind":"struct","package":"bytes","name":"Buffer","string":"bytes.Buffer","elements":{"buf":"4","lastRead":"7","off":"6"},"anonymous":{"buf":false,"lastRead":false,"off":false},"docs":{"":"","buf":"","lastRead":"","off":""}},{"id":"4","kind":"slice","package":"","name":"","string":"[]uint8","elements":{"":"5"},"anonymous":{},"docs":{"":""}},{"id":"5","kind":"uint8","package":"","name":"uint8","string":"uint8","elements":{},"anonymous":{},"docs":{"":""}},{"id":"6","kind":"int","package":"","name":"int","string":"int","elements":{},"anonymous":{},"docs":{"":""}},{"id":"7","kind":"int8","package":"bytes","name":"readOp","string":"bytes.readOp","elements":{},"anonymous":{},"docs":{"":""}},{"id":"8","kind":"interface","package":"io","name":"Reader","string":"io.Reader","elements":{"Read":"9"},"anonymous":{},"docs":{"":"","Read":""}},{"id":"9","kind":"func","package":"","name":"","string":"func([]uint8) (int, error)","elements":{},"anonymous":{},"docs":{"":""}},{"id":"10","kind":"interface","package":"github.com/go-zing/gozz-kit/ztree","name":"_field","string":"ztree._field","elements":{"Read":"9"},"anonymous":{},"docs":{"":"field interface","Read":"read function"}}]}`)

var (
	ptr = new(bytes.Buffer)
	doc = make(map[reflect.Type]map[string]string)
)

func init() {
	for k, v := range _types_doc {
		doc[reflect.TypeOf(k).Elem()] = v
	}
}

func TestParse(t *testing.T) {
	b, err := json.Marshal(Parse(
		[]interface{}{&_T0{
			Field:  ptr,
			_field: ptr,
		}, _T0{
			Field:  ptr,
			_field: ptr,
		}},
		WithExpandTypes([]interface{}{_T0{}}),
		WithDocFunc(func(p reflect.Type, field string) string { return doc[p][field] })),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, want) {
		t.Fatalf("%s", b)
	}
}

func TestDraw(t *testing.T) {
	t.Logf("%s", Draw("test",
		Parse([]interface{}{&_T0{
			Field:  ptr,
			_field: ptr,
			X:      &bytes.Buffer{},
		}, _T0{
			Field:  ptr,
			_field: ptr,
		}, ptr, 9},
			WithExpandPackages([]interface{}{bytes.Buffer{}}),
			WithExpandTypes([]interface{}{_T0{}}),
			WithUnexported(true),
			WithDocFunc(func(p reflect.Type, field string) string { return doc[p][field] }),
		),
	))
}
