package envparser

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
)

type Env string

type Parser interface {
	Parse(e Env) error
}

var parserType = reflect.TypeOf((*Parser)(nil)).Elem()

func Parse(target interface{}) error {
	return parse(target, os.LookupEnv)
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

func parse(target interface{}, lookupFn func(string) (string, bool)) error {
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
		case f.Type().Implements(parserType):
			if err := f.Interface().(Parser).Parse(Env(val)); err != nil {
				parseError.append(key, val, err)
			}
		case f.Addr().Type().Implements(parserType):
			if err := f.Addr().Interface().(Parser).Parse(Env(val)); err != nil {
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
