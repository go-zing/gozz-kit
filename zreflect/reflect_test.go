package zreflect

import (
	"testing"
)

func TestParseStructTag(t *testing.T) {
	v := ParseTag(`json:"test" bson:"test,omitempty"`)
	if len(v) != 2 || v[1].Value != "test,omitempty" {
		t.Fatal()
	}
}
