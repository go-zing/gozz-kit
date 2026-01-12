package zswagger_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/go-zing/gozz-kit/internal/ztest"
	"github.com/go-zing/gozz-kit/zapi/zswagger"
	"github.com/go-zing/gozz-kit/zdoc"
)

func TestParse(t *testing.T) {
	swagger := zswagger.Parse(ztest.Apis{},
		zswagger.WithDocFunc(zdoc.TypesDoc(ztest.ZZ_types_doc).TypeFieldDoc),
		zswagger.WithBindings(map[string]zswagger.Binding{
			"GET": {
				Path:   "uri",
				Query:  "form",
				Header: "",
				Body:   false,
			},
			"POST": {
				Path:   "uri",
				Header: "",
				Body:   true,
			},
		}),
	)

	// Test JSON marshaling
	b, err := json.MarshalIndent(swagger, "", "    ")
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	// Verify the JSON can be unmarshaled back
	var unmarshaled spec.Swagger
	if err := json.Unmarshal(b, &unmarshaled); err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	_ = ioutil.WriteFile("example.json", b, 0o664)
}
