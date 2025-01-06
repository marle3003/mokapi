package util

import "reflect"

func JsType(v interface{}) string {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array, reflect.Slice:
		return "Array"
	case reflect.Int64:
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
