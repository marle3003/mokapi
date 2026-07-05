package monitor_test

import (
	"context"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"mokapi/safe"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMonitor_Start(t *testing.T) {
	t.Parallel()

	m := monitor.New()
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

	m := monitor.New()
	m.Http.RequestCounter.WithLabel("s", "e", "m").Add(1)
	r := m.FindAll(metrics.ByNamespace("http"))
	require.Len(t, r, 1)
}
