# ZSwagger

Generate [Swagger 2.0](https://swagger.io/specification/v2/) OpenAPI Documentation

## Usage

```go
package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/go-zing/gozz-kit/zapi/zswagger"
	"github.com/go-zing/gozz-kit/zdoc"
)

func main() {
	swagger := zswagger.Parse(
		// Apis in zzgen.api.go 
		types.Apis{},
		// Docs in zzgen.doc.go
		zswagger.WithDocFunc(zdoc.TypesDoc(types.ZZ_types_doc).TypeFieldDoc),
		
		// binding rules
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
	b, err := json.MarshalIndent(swagger, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	_ = ioutil.WriteFile("example.json", b, 0o664)
}

```