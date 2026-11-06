package common_test

import (
	"mokapi/engine/common"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventLog(t *testing.T) {
	t.Run("zero values", func(t *testing.T) {
		var l common.EventLog

		require.Nil(t, l.GetActions())
		require.Equal(t, int64(0), l.GetDuration())
	})

	t.Run("set and get", func(t *testing.T) {
		var l common.EventLog
		actions := []*common.Action{{Duration: 5}}

		l.SetActions(actions)
		l.SetDuration(42)

		require.Equal(t, actions, l.GetActions())
		require.Equal(t, int64(42), l.GetDuration())
	})

	t.Run("concurrent set and get", func(t *testing.T) {
		var l common.EventLog

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(4)
			go func() {
				defer wg.Done()
				l.SetActions([]*common.Action{{Duration: 1}})
			}()
			go func() {
				defer wg.Done()
				l.SetDuration(7)
			}()
			go func() {
				defer wg.Done()
				_ = l.GetActions()
			}()
			go func() {
				defer wg.Done()
				_ = l.GetDuration()
			}()
		}
		wg.Wait()

		require.Equal(t, int64(7), l.GetDuration())
	})
}
