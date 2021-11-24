package envparser

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
)

type Unmarshaler interface {
	UnmarshalEnv(e string) error
}

var unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

func ParseInto(target interface{}) error {
	return parseInto(target, os.LookupEnv)
}

func getEnvKey(f reflect.StructField) string {
	var key string
	rawTags := f.Tag.Get("env")
	tags := strings.Split(rawTags, ",")
	if len(tags) > 0 {
		key = tags[0]
	}
	if key == "" {
		key = f.Name
	}
	return key
}

func parseInto(target interface{}, lookupFn func(string) (string, bool)) error {
	var targetVal reflect.Value
	if v := reflect.ValueOf(target); v.Kind() == reflect.Ptr {
		targetVal = v.Elem()
	}
	if targetVal.Kind() != reflect.Struct {
		panic("envparser: target must be non-nil pointer to struct")
	}

	var parseError ParseError

	for t, i := targetVal.Type(), 0; i < t.NumField(); i++ {
		field := targetVal.Type().Field(i)
		if !field.IsExported() {
			continue
		}

		key := getEnvKey(field)
		val, ok := lookupFn(key)
		if !ok {
			continue
		}

		f := targetVal.Field(i)
		switch {
		case f.Type().Implements(unmarshalerType):
			if err := f.Interface().(Unmarshaler).UnmarshalEnv(val); err != nil {
				parseError.append(key, val, err)
			}
		case f.Addr().Type().Implements(unmarshalerType):
			if err := f.Addr().Interface().(Unmarshaler).UnmarshalEnv(val); err != nil {
				parseError.append(key, val, err)
			}
		case f.Kind() == reflect.String:
			f.Set(reflect.ValueOf(val))
		default:
			if err := json.Unmarshal([]byte(val), f.Addr().Interface()); err != nil {
				parseError.append(key, val, err)
			}
		}
	}

	if len(parseError.Items) > 0 {
		return &parseError
	}

	return nil
}