package functions

import (
	"time"
)

func Now(_ ...interface{}) (interface{}, error) {
	return time.Now(), nil
}
