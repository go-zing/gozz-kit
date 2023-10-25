package zapi

import (
	"encoding/json"
	"testing"

	ztest2 "github.com/go-zing/gozz-kit/internal/ztest"
)

func TestParseStructTag(t *testing.T) {
	v := parseReflectTag(`json:"test" bson:"test,omitempty"`)
	if len(v) != 2 || v[1].Value != "test,omitempty" {
		t.Fatal()
	}
}

func TestParse(t *testing.T) {
	p := Parser{docFunc: ztest2.Docs.TypeFieldDoc}
	groups, payloads := p.Parse(ztest2.Apis{})
	t.Log(groups, payloads)
}

type Str struct {
	Str2
	Test string `json:"test,omitempty"`
}

type Str2 struct {
	*Str
	D string `json:"d"`
}

func TestName(t *testing.T) {
	var d2 Str

	// _ = json.Unmarshal([]byte(`{"str":{"test":"xx"}}`), &d2)
	_ = json.Unmarshal([]byte(`{"test":"xx"}`), &d2)
	d2.Str = &Str{
		Test: "dddd",
	}
	d2.Test = ""
	d2.D = "ee"
	b, _ := json.Marshal(d2)
	t.Logf("%s", b)
}
