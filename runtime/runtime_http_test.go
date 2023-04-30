package runtime

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp_NoHttp(t *testing.T) {
	defer events.Reset()

	err := events.Push("bar", events.NewTraits().WithNamespace("http").WithName("foo"))
	require.EqualError(t, err, "no store found for namespace=http, name=foo")
}

func TestApp_AddHttp(t *testing.T) {
	defer events.Reset()

	app := New()
	app.AddHttp(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "")))

	require.Contains(t, app.Http, "foo")
	err := events.Push("bar", events.NewTraits().WithNamespace("http").WithName("foo"))
	require.NoError(t, err, "event store should be available")
}

func TestApp_AddHttp_Path(t *testing.T) {
	defer events.Reset()

	app := New()
	app.AddHttp(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
		openapitest.WithEndpoint("bar", openapitest.NewEndpoint())))

	require.Contains(t, app.Http, "foo")
	err := events.Push("bar", events.NewTraits().WithNamespace("http").WithName("foo").With("path", "bar"))
	require.NoError(t, err, "event store should be available")
}

func TestHttpHandler(t *testing.T) {
	hf := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		v, ok := monitor.HttpFromContext(request.Context())
		require.True(t, ok)
		require.NotNil(t, v)
	})
	h := NewHttpHandler(New().Monitor.Http, hf)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "https://foo.bar", nil))
}
