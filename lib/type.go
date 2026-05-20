package lib

import (
	"encoding/json"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
)

func TypeFrom(v interface{}) string {
	return TypeString(reflect.TypeOf(v))
}

func TypeString(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Array, reflect.Slice:
		return "Array"
	case reflect.Int64, reflect.Int:
		return "Integer"
	case reflect.Float64:
		return "Number"
	case reflect.Bool:
		return "Boolean"
	case reflect.Map, reflect.Struct:
		return "Object"
	case reflect.String:
		return "String"
	case reflect.Ptr:
		return "*" + TypeString(t.Elem())
	default:
		return "Unknown"
	}
}

func PrettyPrint(v any) string {
	if v == nil {
		return "<nil>"
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		return string(b)
	}
	log.Warnf("failed to pretty print %#v: %v", v, err)
	return fmt.Sprintf("%v", v)
}
