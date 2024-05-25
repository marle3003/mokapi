package parser

import (
	"fmt"
	"mokapi/sortedmap"
	"reflect"
	"strings"
	"unicode"
)

func toString(i interface{}) string {
	var sb strings.Builder
	switch o := i.(type) {
	case []interface{}:
		sb.WriteRune('[')
		for i, v := range o {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(toString(v))
		}
		sb.WriteRune(']')
	case map[string]interface{}:
		sb.WriteRune('{')
		for key, val := range o {
			if sb.Len() > 1 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v: %v", key, toString(val)))
		}
		sb.WriteRune('}')
	case string, int, int32, int64, float32, float64:
		sb.WriteString(fmt.Sprintf("%v", o))
	case *sortedmap.LinkedHashMap[string, interface{}]:
		return o.String()
	default:
		v := reflect.ValueOf(i)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		t := reflect.TypeOf(i)
		switch v.Kind() {
		case reflect.Slice:
			sb.WriteRune('[')
			for i := 0; i < v.Len(); i++ {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(toString(v.Index(i).Interface()))
			}
			sb.WriteRune(']')
		case reflect.Struct:
			sb.WriteRune('{')
			for i := 0; i < v.NumField(); i++ {
				if i > 0 {
					sb.WriteString(", ")
				}
				name := t.Field(i).Name
				fv := v.Field(i).Interface()
				sb.WriteString(fmt.Sprintf("%v: %v", name, fv))
			}
			sb.WriteRune('}')
		}
	}
	return sb.String()
}

func compare(a, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	if av.Kind() != bv.Kind() {
		return false
	}

	k := av.Kind()
	_ = k

	switch av.Kind() {
	case reflect.Slice:
		return compareSlice(av, bv)
	case reflect.Map:
		return compareMap(av, bv)
	case reflect.Struct, reflect.Pointer:
		return compareStruct(av, bv)
	default:
		return a == b
	}
}

func compareSlice(a, b reflect.Value) bool {
	if a.Len() != b.Len() {
		return false
	}
	for i := 0; i < a.Len(); i++ {
		if !compare(a.Index(i).Interface(), b.Index(i).Interface()) {
			return false
		}
	}
	return true
}

func compareMap(a, b reflect.Value) bool {
	if a.Len() != b.Len() {
		return false
	}
	for _, k := range a.MapKeys() {
		av := a.MapIndex(k)
		bv := b.MapIndex(k)
		if !compare(av.Interface(), bv.Interface()) {
			return false
		}

	}

	return true
}

func compareStruct(a, b reflect.Value) bool {
	m1 := a.Interface().(*sortedmap.LinkedHashMap[string, interface{}])
	m2 := b.Interface().(*sortedmap.LinkedHashMap[string, interface{}])
	for it := m1.Iter(); it.Next(); {
		v, _ := m2.Get(it.Key())
		if !compare(it.Value(), v) {
			return false
		}
	}
	return true
}

func toFieldName(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
