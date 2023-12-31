// Code generated by gozz:option github.com/go-zing/gozz. DO NOT EDIT.

package zswagger

import (
	zapi "github.com/go-zing/gozz-kit/zapi"
	"reflect"
)

// apply functional options for Option
func (o *Option) applyOptions(opts ...func(*Option)) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithHttpCast(v func(api zapi.Api) zapi.HttpApi) func(*Option) {
	return func(o *Option) { o.HttpCast = v }
}

func WithBindings(v map[string]Binding) func(*Option) { return func(o *Option) { o.Bindings = v } }

func WithDocFunc(v func(reflect.Type, string) string) func(*Option) {
	return func(o *Option) { o.DocFunc = v }
}
