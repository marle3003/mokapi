package try

import (
	"mokapi/config/dynamic"
	"testing"
	"time"
)

func GetConfig(t *testing.T, ch chan dynamic.ConfigEvent, timeout time.Duration) *dynamic.Config {
	chTimeout := time.After(timeout)
	select {
	case c := <-ch:
		return c.Config
	case <-chTimeout:
		t.Fatal("timeout while waiting for config event")
	}

	return nil
}
