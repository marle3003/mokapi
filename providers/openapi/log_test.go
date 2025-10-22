package openapi_test

import (
	"context"
	"mokapi/providers/openapi"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLog(t *testing.T) {
	testcases := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "accept header parameter",
			test: func(t *testing.T) {
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("accept", "application/json")

				r = r.WithContext(openapi.NewContext(context.Background(), &openapi.RequestParameters{}))

				ctx, err := openapi.NewLogEventContext(r, false, &eventstest.Handler{}, events.NewTraits())
				require.NoError(t, err)
				require.NotNil(t, ctx)

				log, ok := openapi.LogEventFromContext(ctx)
				require.True(t, ok)
				require.NotNil(t, log)
				require.Len(t, log.Request.Parameters, 1)
				require.Equal(t, "Accept", log.Request.Parameters[0].Name)
				require.Equal(t, "header", log.Request.Parameters[0].Type)
				require.Equal(t, "application/json", *log.Request.Parameters[0].Raw)
				require.Equal(t, "", log.Request.Parameters[0].Value)
			},
		},
		{
			name: "parameter defined in spec",
			test: func(t *testing.T) {
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("Foo", "bar")

				params := &openapi.RequestParameters{Header: map[string]openapi.RequestParameterValue{}}
				v := "bar"
				params.Header["Foo"] = openapi.RequestParameterValue{
					Value: v,
					Raw:   &v,
				}
				r = r.WithContext(openapi.NewContext(context.Background(), params))

				ctx, err := openapi.NewLogEventContext(r, false, &eventstest.Handler{}, events.NewTraits())
				require.NoError(t, err)
				require.NotNil(t, ctx)

				log, ok := openapi.LogEventFromContext(ctx)
				require.True(t, ok)
				require.NotNil(t, log)
				require.Len(t, log.Request.Parameters, 1)
				require.Equal(t, "Foo", log.Request.Parameters[0].Name)
				require.Equal(t, "header", log.Request.Parameters[0].Type)
				require.Equal(t, "bar", *log.Request.Parameters[0].Raw)
				require.Equal(t, `"bar"`, log.Request.Parameters[0].Value)
			},
		},
		{
			name: "parameter defined in spec case insensitive",
			test: func(t *testing.T) {
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("Foo", "bar")

				params := &openapi.RequestParameters{Header: map[string]openapi.RequestParameterValue{}}
				v := "bar"
				params.Header["foo"] = openapi.RequestParameterValue{
					Value: v,
					Raw:   &v,
				}
				r = r.WithContext(openapi.NewContext(context.Background(), params))

				ctx, err := openapi.NewLogEventContext(r, false, &eventstest.Handler{}, events.NewTraits())
				require.NoError(t, err)
				require.NotNil(t, ctx)

				log, ok := openapi.LogEventFromContext(ctx)
				require.True(t, ok)
				require.NotNil(t, log)
				require.Len(t, log.Request.Parameters, 1)
				// header parameter name is lower case because name is used from specification
				require.Equal(t, "foo", log.Request.Parameters[0].Name)
				require.Equal(t, "header", log.Request.Parameters[0].Type)
				require.Equal(t, "bar", *log.Request.Parameters[0].Raw)
				require.Equal(t, `"bar"`, log.Request.Parameters[0].Value)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.test(t)
		})
	}
}
