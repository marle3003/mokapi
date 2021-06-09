package functions

import "math/rand"

func RandInt(_ ...interface{}) (interface{}, error) {
	return rand.Int(), nil
}

func RandFloat(_ ...interface{}) (interface{}, error) {
	return rand.Float64(), nil
}
