package dynamic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"reflect"
	"strings"
	"unicode"
)

type decoder struct {
	d *json.Decoder
	b []byte
}

func UnmarshalJSON(b []byte, v interface{}) error {
	d := &decoder{
		d: json.NewDecoder(bytes.NewReader(b)),
		b: b,
	}

	err := unmarshalJSON(d, reflect.ValueOf(v))
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return fmt.Errorf("unexpected end of JSON input")
	}
	return err
}

func NextTokenIndex(b []byte) int64 {
	for n, c := range b {
		switch c {
		case ':', ' ':
			continue
		}
		return int64(n)
	}
	return 0
}

func unmarshalJSON(d *decoder, v reflect.Value) error {
	if unmarshaler(v) {
		v = indirect(v, false)
		p := reflect.New(v.Type())
		p.Elem().Set(v)
		err := d.d.Decode(p.Interface())
		if err != nil {
			return err
		}
		v.Set(p.Elem())
		return nil
	}

	token, err := d.d.Token()
	if err != nil {
		return err
	}

	return value(token, d, v)
}

func value(token json.Token, d *decoder, v reflect.Value) error {
	switch t := token.(type) {
	case json.Delim:
		if t == '{' {
			return object(d, v)
		} else {
			return array(d, v)
		}
	case string:
		v = indirect(v, false)
		switch v.Kind() {
		case reflect.String:
			v.SetString(t)
		case reflect.Interface:
			v.Set(reflect.ValueOf(t))
		default:
			return fmt.Errorf("unsupported cast string '%s' to %s", t, toTypeName(v))
		}

		return nil
	case float64:
		return number(t, v)
	case bool:
		v = indirect(v, false)
		v.Set(reflect.ValueOf(t))
		return nil
	case nil:
		switch v.Kind() {
		case reflect.Interface, reflect.Pointer, reflect.Map, reflect.Slice:
			v = indirect(v, true)
			v.Set(reflect.Zero(v.Type()))
			return nil
		default:
			// otherwise, ignore null for primitives/string
			return nil
		}
	default:
		return fmt.Errorf("unsupported token: %v, %T", token, token)
	}
}

func object(d *decoder, v reflect.Value) error {
	v = indirect(v, false)
	// check type
	switch v.Kind() {
	case reflect.Struct, reflect.Map:
		break
	case reflect.Interface:
		v.Set(reflect.ValueOf(map[string]interface{}{}))
		v = v.Elem()
	default:
		return fmt.Errorf("expected %s but received an object", toTypeName(v))
	}

	i := v.Interface()
	_ = i

	// The first byte of the object '{' has been read already.
	for {
		offset := d.d.InputOffset()

		token, err := d.d.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}
		key := token.(string)
		field, err := getField(v, key)
		if err != nil {
			return NewStructuralError(err, offset, d.d)
		} else if !field.IsValid() {
			err = skip(d.d)
			if err != nil {
				return err
			}
			continue
		}

		offset = d.d.InputOffset()
		err = unmarshalJSON(d, field)
		if err != nil {
			offset += NextTokenIndex(d.b[offset:])
			return NewStructuralErrorWithField(err, offset, d.d, key)
		}

		// write value back to map
		if v.Kind() == reflect.Map {
			v.SetMapIndex(reflect.ValueOf(key), field)
		}
	}
}

func array(d *decoder, v reflect.Value) error {
	v = indirect(v, false)
	isAny := false
	// check type
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		break
	case reflect.Interface:
		v.Set(reflect.ValueOf([]interface{}{}))
		isAny = true
	default:
		return fmt.Errorf("expected %s but received an array", toTypeName(v))
	}

	var err error
	var token json.Token
	// The first byte of the object '[' has been read already.
	for {
		if !d.d.More() {
			token, err = d.d.Token()
			if err != nil {
				return err
			}
			if delim, ok := token.(json.Delim); ok && delim == ']' {
				return nil
			}
		}

		if isAny {
			token, err = d.d.Token()
			if err != nil {
				return err
			}
			p := newValueForToken(token)
			err = value(token, d, p)
			if err != nil {
				return err
			}

			v.Set(reflect.Append(v.Elem(), p.Elem()))
		} else {
			p := reflect.New(v.Type().Elem())
			if unmarshaler(p) {
				err = d.d.Decode(p.Interface())
				if err != nil {
					return err
				}
			} else {
				token, err = d.d.Token()
				if err != nil {
					return err
				}

				err = value(token, d, p)
				if err != nil {
					return err
				}
			}

			v.Set(reflect.Append(v, p.Elem()))
		}

	}
}

func number(f float64, v reflect.Value) error {
	v = indirect(v, false)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := int64(f)
		if float64(n) != f {
			return fmt.Errorf("expected %s but received a floating number", toTypeName(v))
		}
		if v.OverflowInt(n) {
			return fmt.Errorf("overflow number %v", n)
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n := uint64(f)
		if float64(n) != f {
			return fmt.Errorf("expected %s but received a floating number", toTypeName(v))
		}
		if v.OverflowUint(n) {
			return fmt.Errorf("overflow number %v", n)
		}
		v.SetUint(n)

	case reflect.Float32, reflect.Float64:
		if v.OverflowFloat(f) {
			return fmt.Errorf("overflow number %v", f)
		}
		v.SetFloat(f)
	case reflect.Interface:
		v.Set(reflect.ValueOf(f))
	default:
		return fmt.Errorf("unsupported cast number '%v' to %s", f, toTypeName(v))
	}
	return nil
}

func unmarshaler(v reflect.Value) bool {
	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Pointer && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}

	for {
		if v.Kind() != reflect.Pointer {
			break
		}

		if _, ok := v.Interface().(json.Unmarshaler); ok {
			return true
		}

		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return false
}

func indirect(v reflect.Value, isNil bool) reflect.Value {
	for {
		if v.Kind() != reflect.Pointer {
			break
		}

		if isNil && v.CanSet() {
			break
		}

		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func getField(v reflect.Value, name string) (reflect.Value, error) {
	if v.Kind() == reflect.Map {
		if v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}

		key := reflect.ValueOf(name)
		mv := v.MapIndex(key)
		elemType := v.Type().Elem()
		if !mv.IsValid() {
			mv = reflect.New(elemType).Elem()
			v.SetMapIndex(key, mv)
		}
		return mv, nil
	}

	fieldName := firstLetterToUpper(name)
	field := v.FieldByNameFunc(func(f string) bool {
		return f == fieldName
	})
	if field.IsValid() {
		if field.Kind() == reflect.Pointer && field.IsNil() {
			fv := reflect.New(field.Type())
			if field.Type().Kind() != reflect.Struct {
				fv = fv.Elem()
			}
			field.Set(fv)
		}

		return field, nil
	}

	i := v.Interface()
	_ = i
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)

		tag := f.Tag.Get("json")
		if len(tag) == 0 {
			continue
		}
		info := strings.Split(tag, ",")
		tagName := info[0]
		if tagName == name {
			return v.Field(i), nil
		}
	}

	log.Debugf("field %v not found", name)
	return reflect.Value{}, nil
}

func firstLetterToUpper(s string) string {

	if len(s) == 0 {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}

func toTypeName(v reflect.Value) string {
	switch v.Type().Kind() {
	case reflect.Slice:
		return "array"
	case reflect.Struct, reflect.Map:
		return "object"
	default:
		return v.Type().Kind().String()
	}
}

func skip(d *json.Decoder) error {
	openBraces := 0
	openSquare := 0
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}
		switch v := token.(type) {
		case json.Delim:
			switch v {
			case '{':
				openBraces++
			case '}':
				openBraces--
			case '[':
				openSquare++
			case ']':
				openSquare--
			}
		}
		if openBraces == 0 && openSquare == 0 {
			break
		}
	}
	return nil
}

func newValueForToken(token json.Token) reflect.Value {
	switch t := token.(type) {
	case json.Delim:
		if t == '{' {
			return reflect.New(reflect.TypeOf(map[string]interface{}{}))
		} else {
			return reflect.New(reflect.TypeOf([]interface{}{}))
		}
	case string:
		return reflect.ValueOf("")
	case float64:
		return reflect.New(reflect.TypeOf(float64(0)))
	}

	return reflect.Value{}
}
