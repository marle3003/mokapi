package runtime

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/runtime/events"
	"net/url"
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
	app.AddHttp(newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))))

	require.Contains(t, app.Http, "foo")
	err := events.Push("bar", events.NewTraits().WithNamespace("http").WithName("foo"))
	require.NoError(t, err, "event store should be available")
}

func TestApp_AddHttp_Path(t *testing.T) {
	defer events.Reset()

	app := New()
	app.AddHttp(newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
		openapitest.WithEndpoint("bar", openapitest.NewEndpoint()))))

	require.Contains(t, app.Http, "foo")
	err := events.Push("bar", events.NewTraits().WithNamespace("http").WithName("foo").With("path", "bar"))
	require.NoError(t, err, "event store should be available")
}

func newConfig(config *openapi.Config) *common.Config {
	c := &common.Config{Data: config}
	u, _ := url.Parse("https://mokapi.io")
	c.Info.Url = u
	return c
}
