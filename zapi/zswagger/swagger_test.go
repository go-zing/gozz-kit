package zswagger

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/go-zing/gozz-kit/internal/ztest"
	"github.com/go-zing/gozz-kit/zapi"
)

func TestParse(t *testing.T) {
	groups, payloads := zapi.NewParser(zapi.WithDocFunc(ztest.Docs.TypeFieldDoc)).Parse(ztest.Apis{})
	fn := zapi.SplitFn("|")
	swagger := Parse(groups, payloads, func(api zapi.Api) zapi.HttpApi {
		method, path := fn(api.Resource)
		return zapi.HttpApi{
			Api:    api,
			Method: method,
			Path:   path,
		}
	})
	b, err := json.MarshalIndent(swagger, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	_ = ioutil.WriteFile("swagger.json", b, 0o664)
}
