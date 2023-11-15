package zreflect

import (
	"strconv"
	"strings"
)

type (
	StructTag struct {
		Key   string
		Value TagValue
	}

	StructTags []StructTag

	TagValue string

	TagValues []string
)

func (tags StructTags) Lookup(key string) (value TagValue, found bool) {
	for i := range tags {
		if tags[i].Key == key {
			return tags[i].Value, true
		}
	}
	return
}

func (tags StructTags) Get(key string) (value TagValue) {
	value, _ = tags.Lookup(key)
	return
}

func (value TagValue) Split(Sep string) TagValues {
	return strings.Split(string(value), Sep)
}

func (values TagValues) Exist(option string) bool {
	for _, v := range values {
		if option == v {
			return true
		}
	}
	return false
}

func ParseTag(tag string) (tags StructTags) {
	for tag != "" {
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}

		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}

		key := tag[:i]
		tag = tag[i+1:]

		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}

		if i >= len(tag) {
			break
		}

		if value, err := strconv.Unquote(tag[:i+1]); err == nil {
			tags = append(tags, StructTag{
				Key:   key,
				Value: TagValue(value),
			})
			tag = tag[i+1:]
		}
	}
	return
}
