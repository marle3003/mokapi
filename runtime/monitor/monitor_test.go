package monitor

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/runtime/metrics"
	"mokapi/safe"
	"testing"
	"time"
)

func TestMonitor_Start(t *testing.T) {
	t.Parallel()

	m := New()
	m.RefreshRateSeconds = 1
	p := safe.NewPool(context.Background())
	m.Start(p)
	defer p.Stop()

	time.Sleep(2 * time.Second)
	require.Greater(t, m.MemoryUsage.Value(), float64(0))
	require.Greater(t, m.StartTime.Value(), float64(0))
}

func TestMonitor_FindAll(t *testing.T) {
	t.Parallel()

	m := New()
	m.Http.RequestCounter.WithLabel("s", "e").Add(1)
	r := m.FindAll(metrics.ByNamespace("http"))
	require.Len(t, r, 1)
}
