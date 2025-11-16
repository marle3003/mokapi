package util

import "mokapi/lib"

func JsType(v interface{}) string {
	return lib.TypeFrom(v)
}
