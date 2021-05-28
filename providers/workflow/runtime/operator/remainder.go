package operator

func Modulo(x interface{}, y interface{}) (interface{}, error) {
	switch x := x.(type) {
	case int:
		switch y := y.(type) {
		case int:
			return x % y, nil
		}
	}

	return 0, &ErrInvalidOperation{x, y, "remainder"}
}
