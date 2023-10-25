package helpers

import (
	"reflect"
)

func IndirectType(rt reflect.Type) reflect.Type {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	return rt
}
