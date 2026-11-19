package safe_test

import (
	"context"
	"mokapi/safe"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSend(t *testing.T) {
	t.Run("delivers to a waiting receiver", func(t *testing.T) {
		ch := make(chan int)
		got := make(chan int, 1)
		go func() { got <- <-ch }()

		require.True(t, safe.Send(context.Background(), ch, 42))
		require.Equal(t, 42, <-got)
	})

	t.Run("gives up when the context is done and nobody reads", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		require.False(t, safe.Send(ctx, make(chan int), 42))
	})
}
