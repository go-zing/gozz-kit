package zswagger

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/go-zing/gozz-kit/zapi"
	"github.com/go-zing/gozz-kit/zapi/internal/ztest"
)

func TestParse(t *testing.T) {
	groups, types := zapi.NewParser(zapi.WithDocFunc(ztest.Docs.TypeFieldDoc)).Parse(ztest.Apis{})
	fn := zapi.SplitFn("|")
	swagger := Parse(groups, types, func(api zapi.Api) zapi.HttpApi {
		method, path := fn(api.Resource)
		return zapi.HttpApi{
			Api:        api,
			Method:     method,
			Path:       path,
			PathParams: nil,
		}
	})
	b, err := json.MarshalIndent(swagger, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	_ = ioutil.WriteFile("swagger.json", b, 0o664)
}
