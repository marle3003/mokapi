package schema

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/sortedmap"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

func ToType(i interface{}) string {
	if i == nil {
		return "null"
	}

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
		return "int"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	default:
		log.Errorf("unable to resolve JSON type from value %v kind %v", ToString(i), v.Kind())
		return "string"
	}
}

func ToString(i interface{}) string {
	var sb strings.Builder
	switch o := i.(type) {
	case []interface{}:
		sb.WriteRune('[')
		for i, v := range o {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(ToString(v))
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
			sb.WriteString(fmt.Sprintf("%v: %v", key, ToString(val)))
		}
		sb.WriteRune('}')
	case string, int, int32, int64, float32, float64, bool:
		sb.WriteString(fmt.Sprintf("%v", o))
	case *sortedmap.LinkedHashMap[string, interface{}]:
		return o.String()
	default:
		if i == nil {
			return "null"
		}
		v := reflect.ValueOf(i)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if !v.IsValid() {
			return "null"
		}
		t := v.Type()
		switch v.Kind() {
		case reflect.Slice:
			sb.WriteRune('[')
			for i := 0; i < v.Len(); i++ {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(ToString(v.Index(i).Interface()))
			}
			sb.WriteRune(']')
		case reflect.Struct:
			fields := strings.Builder{}
			exportedFields := 0
			for i := 0; i < t.NumField(); i++ {
				if i > 0 {
					fields.WriteString(", ")
				}
				name := t.Field(i).Name
				if unicode.IsUpper(rune(name[0])) {
					exportedFields++
					fv := v.Field(i).Interface()
					fields.WriteString(fmt.Sprintf("%v: %v", firstLetterToLower(name), fv))
				}
			}
			if exportedFields == 0 {
				sb.WriteString(fmt.Sprintf("%v", i))
			} else {
				sb.WriteString(fmt.Sprintf("{%s}", fields.String()))
			}
		default:
			log.Errorf("JSON schema to string: unsupported type: %v", v.Kind())
		}
	}
	return sb.String()
}

func firstLetterToLower(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToLower(r[0])

	return string(r)
}
