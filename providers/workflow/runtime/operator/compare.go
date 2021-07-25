package operator

import "strings"

func Compare(x interface{}, y interface{}) (int, error) {
	if x == nil {
		if y == nil {
			return 0, nil
		}
		return -1, nil
	}
	if y == nil {
		return 1, nil
	}
	switch x := x.(type) {
	case int:
		switch y := y.(type) {
		case int:
			if x < y {
				return -1, nil
			} else if x > y {
				return 1, nil
			}
			return 0, nil
		case float64:
			if float64(x) < y {
				return -1, nil
			} else if float64(x) > y {
				return 1, nil
			}
			return 0, nil
		}
	case float64:
		switch y := y.(type) {
		case int:
			if x < float64(y) {
				return -1, nil
			} else if x > float64(y) {
				return 1, nil
			}
			return 0, nil
		case float64:
			if x < y {
				return -1, nil
			} else if x > y {
				return 1, nil
			}
			return 0, nil
		}
	case string:
		switch y := y.(type) {
		case string:
			return strings.Compare(x, y), nil
		}
	}

	return 0, &ErrInvalidOperation{x, y, "compare"}
}
