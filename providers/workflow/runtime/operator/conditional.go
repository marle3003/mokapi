package operator

func And(x interface{}, y interface{}) (bool, error) {
	switch x := x.(type) {
	case bool:
		switch y := y.(type) {
		case bool:
			return x && y, nil
		}
	}

	return false, &ErrInvalidOperation{x, y, "and"}
}

func Or(x interface{}, y interface{}) (bool, error) {
	switch x := x.(type) {
	case bool:
		switch y := y.(type) {
		case bool:
			return x || y, nil
		}
	}

	return false, &ErrInvalidOperation{x, y, "and"}
}
