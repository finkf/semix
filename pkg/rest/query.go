package rest

import (
	"errors"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// DecodeQuery decodes a query into the given struct.
// It can only be used to decode bool, int, float32, float64 and string values.
// All other typed fileds are ignored.
func DecodeQuery(vals url.Values, out interface{}) error {
	t := reflect.TypeOf(out)
	if t.Kind() != reflect.Ptr {
		return errors.New("cannot decode: not a pointer")
	}
	if t.Elem().Kind() != reflect.Struct {
		return errors.New("cannot decode: not a struct")
	}
	for i := 0; i < t.Elem().NumField(); i++ {
		v := vals.Get(strings.ToLower(t.Elem().Field(i).Name))
		if v == "" {
			continue
		}
		switch t.Elem().Field(i).Type.Kind() {
		case reflect.Bool:
			x, err := decodeBool(v)
			if err != nil {
				return err
			}
			reflect.ValueOf(out).Elem().Field(i).SetBool(x)
		case reflect.Float64:
			x, err := decodeFloat64(v)
			if err != nil {
				return err
			}
			reflect.ValueOf(out).Elem().Field(i).SetFloat(x)
		case reflect.Float32:
			x, err := decodeFloat32(v)
			if err != nil {
				return err
			}
			reflect.ValueOf(out).Elem().Field(i).SetFloat(float64(x))
		case reflect.Int:
			x, err := decodeInt(v)
			if err != nil {
				return err
			}
			reflect.ValueOf(out).Elem().Field(i).SetInt(int64(x))
		case reflect.String:
			reflect.ValueOf(out).Elem().Field(i).SetString(v)
		case reflect.Slice:
			for _, v := range vals[strings.ToLower(t.Elem().Field(i).Name)] {
				switch t.Elem().Field(i).Type.Elem().Kind() {
				case reflect.Bool:
					x, err := decodeBool(v)
					if err != nil {
						return err
					}
					reflect.ValueOf(out).Elem().Field(i).Set(
						reflect.Append(reflect.ValueOf(out).Elem().Field(i), reflect.ValueOf(x)),
					)
				case reflect.Float64:
					x, err := decodeFloat64(v)
					if err != nil {
						return err
					}
					reflect.ValueOf(out).Elem().Field(i).Set(
						reflect.Append(reflect.ValueOf(out).Elem().Field(i), reflect.ValueOf(x)),
					)
				case reflect.Float32:
					x, err := decodeFloat32(v)
					if err != nil {
						return err
					}
					reflect.ValueOf(out).Elem().Field(i).Set(
						reflect.Append(reflect.ValueOf(out).Elem().Field(i), reflect.ValueOf(x)),
					)
				case reflect.Int:
					x, err := decodeInt(v)
					if err != nil {
						return err
					}
					reflect.ValueOf(out).Elem().Field(i).Set(
						reflect.Append(reflect.ValueOf(out).Elem().Field(i), reflect.ValueOf(x)),
					)
				case reflect.String:
					reflect.ValueOf(out).Elem().Field(i).Set(
						reflect.Append(reflect.ValueOf(out).Elem().Field(i), reflect.ValueOf(v)),
					)
				}
			}
		}
	}
	return nil
}

func decodeInt(val string) (int, error) {
	return strconv.Atoi(val)
}

func decodeFloat64(val string) (float64, error) {
	x, err := strconv.ParseFloat(val, 64)
	return float64(x), err
}

func decodeFloat32(val string) (float32, error) {
	x, err := strconv.ParseFloat(val, 32)
	return float32(x), err
}

func decodeBool(val string) (bool, error) {
	return strconv.ParseBool(val)
}
