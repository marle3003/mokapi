package operator

func Substract(x interface{}, y interface{}) (interface{}, error) {
	switch x := x.(type) {
	case int:
		switch y := y.(type) {
		case int:
			return x - y, nil
		case float64:
			return float64(x) - y, nil
		}
	case float64:
		switch y := y.(type) {
		case int:
			return x - float64(y), nil
		case float64:
			return x - y, nil
		}
	}

	return 0, &ErrInvalidOperation{x, y, "add"}
}
