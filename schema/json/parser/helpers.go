package parser

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/sortedmap"
	"reflect"
	"sort"
	"strings"
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
		// order object by property names to avoid random output
		keys := make([]string, 0, len(o))
		for k := range o {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		sb.WriteRune('{')
		for _, key := range keys {
			val := o[key]
			if sb.Len() > 1 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v: %v", key, toString(val)))
		}
		sb.WriteRune('}')
	case string, int, int32, int64, float32, float64, bool:
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
		default:
			log.Errorf("JSON schema to string: unsupported type: %v", v.Kind())
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

func toType(i interface{}) string {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Struct, reflect.Map:
		return "object"
	case reflect.Int, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	default:
		log.Errorf("unable to resolve JSON type from value: %v", toString(i))
		return "string"
	}
}
