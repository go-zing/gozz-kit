// Code generated by gozz:option github.com/go-zing/gozz. DO NOT EDIT.

package ztree

import (
	"reflect"
)

// apply functional options for Option
func (o *Option) applyOptions(opts ...func(*Option)) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithExpandPackages(v []interface{}) func(*Option) {
	return func(o *Option) { o.ExpandPackages = v }
}

func WithExpandTypes(v []interface{}) func(*Option) { return func(o *Option) { o.ExpandTypes = v } }

func WithUnexported(v bool) func(*Option) { return func(o *Option) { o.Unexported = v } }

func WithDocFunc(v func(p reflect.Type, field string) string) func(*Option) {
	return func(o *Option) { o.DocFunc = v }
}
