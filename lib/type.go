package lib

import "reflect"

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
	case reflect.Map:
		return "Object"
	case reflect.String:
		return "String"
	default:
		return "Unknown"
	}
}
