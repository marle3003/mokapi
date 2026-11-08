package eventstest_test

import (
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandler_PushAndGetEventsAreSafeForConcurrentUse(t *testing.T) {
	h := &eventstest.Handler{}
	traits := events.NewTraits().WithNamespace("test")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			require.NoError(t, h.Push(&eventstest.Event{Name: "foo"}, traits))
		}()
		go func() {
			defer wg.Done()
			_ = h.GetEvents(traits)
		}()
	}
	wg.Wait()

	require.Len(t, h.GetEvents(traits), 10)
}

func TestHandler_GetEventsFiltersByTraits(t *testing.T) {
	h := &eventstest.Handler{}
	require.NoError(t, h.Push(&eventstest.Event{Name: "foo"}, events.NewTraits().WithNamespace("a")))
	require.NoError(t, h.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("b")))

	result := h.GetEvents(events.NewTraits().WithNamespace("b"))

	require.Len(t, result, 1)
	require.Equal(t, "bar", result[0].Data.(*eventstest.Event).Name)
}
