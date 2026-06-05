package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/runtimetest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvents(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		code string
		test func(t *testing.T, result any, err error)
	}{
		{
			name: "without params should not error",
			code: "mokapi.getEvents()",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http"), &openapi.HttpLog{
					Api:  "API",
					Path: "/foo",
					Request: &openapi.HttpRequestLog{
						Method: "GET",
					},
					Response: &openapi.HttpResponseLog{
						StatusCode: 200,
					},
				}),
			),
			test: func(t *testing.T, result any, err error) {
				require.NoError(t, err)
				require.IsType(t, []any{}, result)
				evts := result.([]any)
				require.Len(t, evts, 1)
				require.IsType(t, &mcp.HttpEvent{}, evts[0])
				evt := evts[0].(*mcp.HttpEvent)
				require.NotEmpty(t, evt.Id)
				require.Equal(t, "http", evt.Type)
				require.NotNil(t, evt.Time)
				require.Equal(t, "API", evt.Api)
				require.Equal(t, "/foo", evt.Path)
				require.Equal(t, "GET", evt.Method)
				require.Equal(t, 200, evt.StatusCode)
			},
		},
		{
			name: "filter by API type http",
			code: "mokapi.getEvents({ type: 'http' })",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("kafka"), &store.KafkaMessageLog{}),
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http"), &openapi.HttpLog{}),
			),
			test: func(t *testing.T, result any, err error) {
				require.NoError(t, err)
				require.IsType(t, []any{}, result)
				evts := result.([]any)
				require.Len(t, evts, 1)
				evt := evts[0].(*mcp.HttpEvent)
				require.Equal(t, "http", evt.Type)
			},
		},
		{
			name: "filter by API type kafka",
			code: "mokapi.getEvents({ type: 'kafka' })",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("kafka").With("topic", "foo"), &store.KafkaMessageLog{
					Offset:    1234,
					Key:       store.LogValue{Value: "key"},
					Message:   store.LogValue{Value: "message"},
					Partition: 8,
					Api:       "cluster",
				}),
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http"), &openapi.HttpLog{}),
			),
			test: func(t *testing.T, result any, err error) {
				require.NoError(t, err)
				require.IsType(t, []any{}, result)
				evts := result.([]any)
				require.Len(t, evts, 1)
				evt := evts[0].(*mcp.KafkaEvent)
				require.NotEmpty(t, evt.Id)
				require.Equal(t, "kafka", evt.Type)
				require.NotNil(t, evt.Time)
				require.Equal(t, "cluster", evt.Api)
				require.Equal(t, "foo", evt.Topic)
				require.Equal(t, 8, evt.Partition)
				require.Equal(t, int64(1234), evt.Offset)
				require.Equal(t, "key", evt.Key)
				require.Equal(t, "message", evt.Message)
			},
		},
		{
			name: "filter by API name",
			code: "mokapi.getEvents({ name: 'bar' })",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").WithName("foo"), &openapi.HttpLog{Api: "foo"}),
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").WithName("bar"), &openapi.HttpLog{Api: "bar"}),
			),
			test: func(t *testing.T, result any, err error) {
				require.NoError(t, err)
				require.IsType(t, []any{}, result)
				evts := result.([]any)
				require.Len(t, evts, 1)
				require.Equal(t, "bar", evts[0].(*mcp.HttpEvent).Api)
			},
		},
		{
			name: "filter by path",
			code: "mokapi.getEvents({ path: '/pets' })",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").With("path", "/users"), &openapi.HttpLog{Api: "foo"}),
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").With("path", "/pets"), &openapi.HttpLog{Api: "bar"}),
			),
			test: func(t *testing.T, result any, err error) {
				require.NoError(t, err)
				require.IsType(t, []any{}, result)
				evts := result.([]any)
				require.Len(t, evts, 1)
				require.Equal(t, "bar", evts[0].(*mcp.HttpEvent).Api)
			},
		},
		{
			name: "filter by method",
			code: "mokapi.getEvents({ method: 'post' })",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").With("method", "GET"), &openapi.HttpLog{Api: "foo"}),
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").With("method", "POST"), &openapi.HttpLog{Api: "bar"}),
			),
			test: func(t *testing.T, result any, err error) {
				require.NoError(t, err)
				evts := result.([]any)
				require.Len(t, evts, 1)
				require.Equal(t, "bar", evts[0].(*mcp.HttpEvent).Api)
			},
		},
		{
			name: "get specific event",
			code: "mokapi.getEvent(mokapi.getEvents({ method: 'GET' })[0].id)",
			app: runtimetest.NewApp(
				runtimetest.WithEvent(events.NewTraits().WithNamespace("http").With("method", "GET"), &openapi.HttpLog{Api: "foo"}),
			),
			test: func(t *testing.T, result any, err error) {
				require.NoError(t, err)
				require.IsType(t, &mcp.HttpEvent{}, result)
				evt := result.(*mcp.HttpEvent)
				require.Equal(t, "foo", evt.Api)
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

			tc.test(t, r.Result, err)
		})
	}
}
