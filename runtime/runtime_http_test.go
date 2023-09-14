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
		openapitest.WithPath("bar", openapitest.NewPath()))))

	require.Contains(t, app.Http, "foo")
	err := events.Push("bar", events.NewTraits().WithNamespace("http").WithName("foo").With("path", "bar"))
	require.NoError(t, err, "event store should be available")
}

func TestApp_AddHttp_Patching(t *testing.T) {
	defer events.Reset()

	newConfig := func(name string, c *openapi.Config) *common.Config {
		cfg := &common.Config{Data: c}
		u, _ := url.Parse(name)
		cfg.Info.Url = u
		return cfg
	}

	testcases := []struct {
		name    string
		configs []*common.Config
		test    func(t *testing.T, app *App)
	}{
		{
			name: "overwrite value",
			configs: []*common.Config{
				newConfig("https://mokapi.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
				newConfig("https://mokapi.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "bar"))),
			},
			test: func(t *testing.T, app *App) {
				info := app.Http["foo"]
				require.Equal(t, "bar", info.Info.Description)
			},
		},
		{
			name: "a is patched with b",
			configs: []*common.Config{
				newConfig("https://mokapi.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
				newConfig("https://mokapi.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "bar"))),
			},
			test: func(t *testing.T, app *App) {
				info := app.Http["foo"]
				require.Equal(t, "foo", info.Info.Description)
			},
		},
		{
			name: "order only by filename",
			configs: []*common.Config{
				newConfig("https://a.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
				newConfig("https://mokapi.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "bar"))),
			},
			test: func(t *testing.T, app *App) {
				info := app.Http["foo"]
				require.Equal(t, "foo", info.Info.Description)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := New()
			for _, c := range tc.configs {
				app.AddHttp(c)
			}
			tc.test(t, app)
		})
	}
}

func newConfig(config *openapi.Config) *common.Config {
	c := &common.Config{Data: config}
	u, _ := url.Parse("https://mokapi.io")
	c.Info.Url = u
	return c
}
