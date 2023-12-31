// Code generated by gozz:api github.com/go-zing/gozz. DO NOT EDIT.

package ztest

import (
	"context"
)

var _ = context.Context(nil)

type Apis struct {
	Api Api
}

func (s Apis) Iterate(fn func(interface{}, []map[string]interface{})) {
	for _, f := range []func() (interface{}, []map[string]interface{}){
		s._Api,
	} {
		fn(f())
	}
}

func (s Apis) _Api() (interface{}, []map[string]interface{}) {
	t := s.Api
	return &t, []map[string]interface{}{
		{
			"name":     "Get",
			"resource": "get|{id}",
			"options":  map[string]string{},
			"invoke": func(ctx context.Context, dec func(interface{}) error) (interface{}, error) {
				var in Payload
				if err := dec(&in); err != nil {
					return nil, err
				}
				return t.Get(ctx, in)
			},
		},
		{
			"name":     "Post",
			"resource": "post",
			"options":  map[string]string{},
			"invoke": func(ctx context.Context, dec func(interface{}) error) (interface{}, error) {
				var in Payload
				if err := dec(&in); err != nil {
					return nil, err
				}
				return t.Post(ctx, in)
			},
		},
	}
}
