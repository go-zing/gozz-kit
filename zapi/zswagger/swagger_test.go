package zswagger

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/go-zing/gozz-kit/internal/ztest"
	"github.com/go-zing/gozz-kit/zapi"
	"github.com/go-zing/gozz-kit/zdoc"
)

func TestParse(t *testing.T) {
	swagger := Parse(ztest.Apis{},
		WithDocFunc(zdoc.TypesDoc(ztest.ZZ_types_doc).TypeFieldDoc),
		WithHttpCast(func(api zapi.Api) zapi.HttpApi {
			sp := strings.SplitN(api.Resource, "|", 2)[:2]
			return zapi.HttpApi{
				Api:    api,
				Method: sp[0],
				Path:   sp[1],
			}
		}),
		WithBindings(map[string]Binding{
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
	_ = ioutil.WriteFile("swagger.json", b, 0o664)
}
