package rest

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// EncodeQuery encodes a query using the given struct.
// It can only encode bool, int, float32, float64 and string values or
// slices of one of these types.
// All other typed fields in the struct are ignored.
func EncodeQuery(in interface{}) (string, error) {
	v := reflect.ValueOf(in)
	t := reflect.TypeOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", errors.New("cannot encode: not a struct")
	}
	b := new(bytes.Buffer)
	for i := 0; i < v.NumField(); i++ {
		name := strings.ToLower(t.Field(i).Name)
		switch v.Field(i).Kind() {
		case reflect.Bool, reflect.Float64, reflect.Float32, reflect.Int, reflect.String:
			if err := encode(b, name, v.Field(i)); err != nil {
				return "", err
			}
		case reflect.Slice:
			for j := 0; j < v.Field(i).Len(); j++ {
				if err := encode(b, name, v.Field(i).Index(j)); err != nil {
					return "", err
				}
			}
		}
	}
	return b.String(), nil
}

func encode(b *bytes.Buffer, name string, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Bool:
		return appendQuery(b, name, fmt.Sprintf("%t", v.Bool()))
	case reflect.Float64, reflect.Float32:
		return appendQuery(b, name, fmt.Sprintf("%f", v.Float()))
	case reflect.Int:
		return appendQuery(b, name, fmt.Sprintf("%d", v.Int()))
	case reflect.String:
		return appendQuery(b, name, v.String())
	}
	return nil
}

func appendQuery(b *bytes.Buffer, name, value string) error {
	if b.Len() == 0 {
		if err := b.WriteByte('?'); err != nil {
			return err
		}
	} else {
		if err := b.WriteByte('&'); err != nil {
			return err
		}
	}
	str := name + "=" + url.QueryEscape(value)
	_, err := b.WriteString(str)
	return err
}

// DecodeQuery decodes a query into the given struct.
// It can only be used to decode bool, int, float32, float64 and string values or
// slices of one of these types.
// All other typed fields of the given struct are ignored.
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
