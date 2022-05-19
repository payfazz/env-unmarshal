package envparser

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"time"
)

type Unmarshaler interface {
	UnmarshalEnv(val string) error
}

var (
	unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
	timeType        = reflect.TypeOf((*time.Time)(nil)).Elem()
	durationType    = reflect.TypeOf((*time.Duration)(nil)).Elem()
)

func Unmarshal(target interface{}) error {
	return unmarshal(target, os.LookupEnv)
}

func lookupEnvName(f reflect.StructField) (string, bool) {
	if !f.IsExported() {
		return "", false
	}

	key := f.Tag.Get("env")
	if key == "" {
		key = f.Name
	}

	return key, true
}

func unmarshal(target interface{}, lookupEnvFn func(string) (string, bool)) error {
	var targetVal reflect.Value
	if v := reflect.ValueOf(target); v.Kind() == reflect.Ptr {
		targetVal = v.Elem()
	}
	if targetVal.Kind() != reflect.Struct {
		panic("envparser: target must be non-nil pointer to struct")
	}

	var parseError ParseError

	for i, t := 0, targetVal.Type(); i < t.NumField(); i++ {
		key, ok := lookupEnvName(targetVal.Type().Field(i))
		if !ok {
			continue
		}

		val, ok := lookupEnvFn(key)
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
			f.SetString(val)
		case f.Type() == timeType:
			t, err := time.Parse(time.RFC3339Nano, val)
			if err != nil {
				parseError.append(key, val, err)
			} else {
				f.Set(reflect.ValueOf(t))
			}
		case f.Type() == durationType:
			d, err := time.ParseDuration(val)
			if err != nil {
				parseError.append(key, val, err)
			} else {
				f.Set(reflect.ValueOf(d))
			}
		default:
			if err := json.Unmarshal([]byte(val), f.Addr().Interface()); err != nil {
				if f.Kind() == reflect.Array || f.Kind() == reflect.Slice {
					if f.Type().Elem().Kind() == reflect.String {
						ss := strings.Split(val, ",")
						for i := range ss {
							ss[i] = strings.TrimSpace(ss[i])
						}
						f.Set(reflect.ValueOf(ss))
					} else {
						if err2 := json.Unmarshal([]byte("["+val+"]"), f.Addr().Interface()); err2 != nil {
							parseError.append(key, val, err) // append first error
						}
					}
				} else {
					parseError.append(key, val, err)
				}
			}
		}
	}

	if len(parseError.Items) > 0 {
		return &parseError
	}

	return nil
}
