package metrics

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCounter_Value(t *testing.T) {
	c := NewCounter()
	c.Add(1)
	c.Add(2.5)
	require.Equal(t, 3.5, c.Value())
}

func TestCounter_ConcurrentAddAndValue(t *testing.T) {
	c := NewCounter()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.Add(1)
		}()
		go func() {
			defer wg.Done()
			_ = c.Value()
		}()
	}
	wg.Wait()

	require.Equal(t, 100.0, c.Value())
}

func TestCounter_Reset(t *testing.T) {
	c := NewCounter()
	c.Add(3)
	c.Reset()
	require.Equal(t, 0.0, c.Value())
}

func TestCounter_ConcurrentAddAndReset(t *testing.T) {
	c := NewCounter()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.Add(1)
		}()
		go func() {
			defer wg.Done()
			c.Reset()
		}()
	}
	wg.Wait()

	c.Reset()
	require.Equal(t, 0.0, c.Value())
}
