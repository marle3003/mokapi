package safe_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/safe"
	"testing"
	"time"
)

func TestSafe(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "start and stop",
			test: func(t *testing.T) {
				ctx := context.Background()
				p := safe.NewPool(ctx)
				b := false
				p.Go(func(ctx context.Context) {
					select {
					case <-ctx.Done():
						b = true
					case <-time.After(10 * time.Second):
					}
				})
				time.Sleep(300 * time.Millisecond)
				p.Stop()
				require.True(t, b, "context should have been canceled")
			},
		},
		{
			name: "atomic",
			test: func(t *testing.T) {
				var b safe.AtomicBool
				require.False(t, b.IsSet())
				b.SetTrue()
				require.True(t, b.IsSet())
				b.SetFalse()
				require.False(t, b.IsSet())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
