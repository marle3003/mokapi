package operator

import (
	"fmt"
)

type ErrInvalidOperation struct {
	X  interface{}
	Y  interface{}
	Op string
}

func (e ErrInvalidOperation) Error() string {
	return fmt.Sprintf("invalid %v %q (%T) and %q (%T) ", e.Op, e.X, e.X, e.Y, e.Y)
}
