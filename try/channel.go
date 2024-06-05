package try

import (
	"mokapi/config/dynamic"
	"testing"
	"time"
)

func GetConfig(t *testing.T, ch chan *dynamic.Config, timeout time.Duration) *dynamic.Config {
	chTimeout := time.After(timeout)
	select {
	case c := <-ch:
		return c
	case <-chTimeout:
		t.Fatal("timeout while waiting for config event")
	}

	return nil
}
