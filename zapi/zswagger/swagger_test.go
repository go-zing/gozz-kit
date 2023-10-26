package zswagger

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/go-zing/gozz-kit/internal/ztest"
	"github.com/go-zing/gozz-kit/zapi"
)

func TestParse(t *testing.T) {
	groups, payloads := zapi.NewParser(zapi.WithDocFunc(ztest.Docs.TypeFieldDoc)).Parse(ztest.Apis{})
	swagger := Parse(groups, payloads, WithHttpCast(func(api zapi.Api) zapi.HttpApi {
		sp := strings.SplitN(api.Resource, "|", 2)[:2]
		return zapi.HttpApi{
			Api:    api,
			Method: sp[0],
			Path:   sp[1],
		}
	}), WithBindings(map[string]Binding{
		"GET": {
			Path:   "uri",
			Query:  "form",
			Header: "",
			Body:   false,
		},
	}))
	b, err := json.MarshalIndent(swagger, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	_ = ioutil.WriteFile("swagger.json", b, 0o664)
}
