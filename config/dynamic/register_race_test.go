package dynamic_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/try"
	"sync"
	"testing"
)

// Registering a config type while another goroutine parses must not race
// (e.g. a ConfigWatcher goroutine left over from an earlier test).
// The race detector is the assertion: without -race this test always passes.
func TestRegister_WhileParsing(t *testing.T) {
	type racing struct {
		Racing string `json:"racing"`
	}
	defer dynamic.Reset()

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				c := &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("racing.json")},
					Raw:  []byte(`{"racing": "1.0"}`),
				}
				_ = dynamic.Parse(c, &dynamictest.Reader{})
			}
		}
	}()

	for i := 0; i < 200; i++ {
		dynamic.Register("racing-unrelated", dynamic.AnyVersion, &racing{})
	}
	close(done)
	wg.Wait()
}
