package helpers

import (
	"reflect"
	"strings"
)

func IndirectType(rt reflect.Type) reflect.Type {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	return rt
}

var Btoi = map[bool]int{true: 1}

func IsGoStandardPackage(pkg string) bool {
	return !strings.Contains(strings.SplitN(pkg, "/", 2)[0], ".")
}
