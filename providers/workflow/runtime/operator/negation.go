package operator

import "fmt"

func Negation(x interface{}) (bool, error) {
	switch x := x.(type) {
	case bool:
		return !x, nil
	}

	return false, fmt.Errorf("invalid operator '!' usage on %T", x)
}
