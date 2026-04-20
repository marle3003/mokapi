package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/runtimetest"
	"testing"

	"github.com/stretchr/testify/require"
)

type testEvent struct {
	Name string
}

func (t *testEvent) Title() string {
	return t.Name
}

func TestEvents(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		code string
		test func(t *testing.T, evts []events.Event, err error)
	}{
		{
			name: "without params should not error",
			code: "mokapi.getEvents()",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http"), &testEvent{Name: "test-1"}),
			),
			test: func(t *testing.T, evts []events.Event, err error) {
				require.NoError(t, err)
				require.Len(t, evts, 1)
				require.Equal(t, &testEvent{Name: "test-1"}, evts[0].Data)
			},
		},
		{
			name: "filter by API type",
			code: "mokapi.getEvents({ type: 'http' })",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("kafka"), &testEvent{Name: "test-1"}),
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http"), &testEvent{Name: "test-2"}),
			),
			test: func(t *testing.T, evts []events.Event, err error) {
				require.NoError(t, err)
				require.Len(t, evts, 1)
				require.Equal(t, &testEvent{Name: "test-2"}, evts[0].Data)
			},
		},
		{
			name: "filter by API name",
			code: "mokapi.getEvents({ name: 'bar' })",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").WithName("foo"), &testEvent{Name: "test-1"}),
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").WithName("bar"), &testEvent{Name: "test-2"}),
			),
			test: func(t *testing.T, evts []events.Event, err error) {
				require.NoError(t, err)
				require.Len(t, evts, 1)
				require.Equal(t, &testEvent{Name: "test-2"}, evts[0].Data)
			},
		},
		{
			name: "filter by path",
			code: "mokapi.getEvents({ path: '/pets' })",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").With("path", "/users"), &testEvent{Name: "test-1"}),
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").With("path", "/pets"), &testEvent{Name: "test-2"}),
			),
			test: func(t *testing.T, evts []events.Event, err error) {
				require.NoError(t, err)
				require.Len(t, evts, 1)
				require.Equal(t, &testEvent{Name: "test-2"}, evts[0].Data)
			},
		},
		{
			name: "filter by method",
			code: "mokapi.getEvents({ method: 'post' })",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").With("method", "GET"), &testEvent{Name: "test-1"}),
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").With("method", "POST"), &testEvent{Name: "test-2"}),
			),
			test: func(t *testing.T, evts []events.Event, err error) {
				require.NoError(t, err)
				require.Len(t, evts, 1)
				require.Equal(t, &testEvent{Name: "test-2"}, evts[0].Data)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := mcp.NewService(tc.app)

			r, err := s.GetRunResponse(
				context.Background(),
				mcp.RunInput{Code: tc.code},
			)
			require.IsType(t, []events.Event{}, r.Result)

			tc.test(t, r.Result.([]events.Event), err)
		})
	}
}
