package runtime_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestApp_AddHttp(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "event store available",
			test: func(t *testing.T, app *runtime.App) {
				app.AddHttp(newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))))

				require.NotNil(t, app.GetHttp("foo"))
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("http").WithName("foo"))
				require.NoError(t, err, "event store should be available")
			},
		},
		{
			name: "event store for endpoint available",
			test: func(t *testing.T, app *runtime.App) {
				app.AddHttp(newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("bar", openapitest.NewPath()))))

				require.NotNil(t, app.GetHttp("foo"))
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("http").WithName("foo").With("path", "bar"))
				require.NoError(t, err, "event store should be available")
			},
		},
		{
			name: "request is counted in monitor",
			test: func(t *testing.T, app *runtime.App) {
				info := app.AddHttp(newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK),
						)))),
				)))
				m := monitor.NewHttp()
				h := info.Handler(m, enginetest.NewEngine(), &events.StoreManager{})

				r := httptest.NewRequest(http.MethodGet, "https://mokapi.io/foo", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)

				require.Equal(t, float64(1), m.RequestCounter.Sum())
			},
		},
		{
			name: "retrieve configs",
			test: func(t *testing.T, app *runtime.App) {
				info := app.AddHttp(newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("bar", openapitest.NewPath()))))

				configs := info.Configs()
				require.Len(t, configs, 1)
				require.Equal(t, "https://mokapi.io", configs[0].Info.Url.String())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg)
			tc.test(t, app)
		})
	}
}

func TestApp_AddHttp_Patching(t *testing.T) {
	newConfig := func(name string, c *openapi.Config) *dynamic.Config {
		cfg := &dynamic.Config{Data: c}
		u, _ := url.Parse(name)
		cfg.Info.Url = u
		return cfg
	}

	testcases := []struct {
		name    string
		configs []*dynamic.Config
		static  *static.Config
		test    func(t *testing.T, app *runtime.App)
	}{
		{
			name: "overwrite value",
			configs: []*dynamic.Config{
				newConfig("https://mokapi.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
				newConfig("https://mokapi.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "bar"))),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.GetHttp("foo")
				require.Equal(t, "bar", info.Info.Description)
				configs := info.Configs()
				require.Len(t, configs, 2)
			},
		},
		{
			name: "a is patched with b",
			configs: []*dynamic.Config{
				newConfig("https://mokapi.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
				newConfig("https://mokapi.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "bar"))),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.GetHttp("foo")
				require.Equal(t, "foo", info.Info.Description)
			},
		},
		{
			name: "order only by filename",
			configs: []*dynamic.Config{
				newConfig("https://a.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
				newConfig("https://mokapi.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "bar"))),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.GetHttp("foo")
				require.Equal(t, "foo", info.Info.Description)
			},
		},
		{
			name: "patch does not reset events",
			configs: []*dynamic.Config{
				newConfig("https://a.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
			},
			test: func(t *testing.T, app *runtime.App) {
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("http").WithName("foo"))
				require.NoError(t, err)

				app.AddHttp(
					newConfig("https://mokapi.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "bar"))),
				)

				e := app.Events.GetEvents(events.NewTraits().WithNamespace("http").WithName("foo"))
				require.Len(t, e, 1)
			},
		},
		{
			name: "patch does not reset events",
			configs: []*dynamic.Config{
				newConfig("https://a.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
			},
			static: &static.Config{Event: static.Event{Store: map[string]static.Store{
				"foo": {Size: 104},
			}}},
			test: func(t *testing.T, app *runtime.App) {
				for i := 0; i < 105; i++ {
					err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("http").WithName("foo"))
					require.NoError(t, err)
				}

				e := app.Events.GetEvents(events.NewTraits().WithNamespace("http").WithName("foo"))
				require.Len(t, e, 104)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.static
			if cfg == nil {
				cfg = &static.Config{}
			}
			app := runtime.New(cfg)
			for _, c := range tc.configs {
				app.AddHttp(c)
			}
			tc.test(t, app)
		})
	}
}

func TestIsHttpConfig(t *testing.T) {
	_, b := runtime.IsHttpConfig(&dynamic.Config{Data: openapitest.NewConfig("3.0")})
	require.True(t, b)
	_, b = runtime.IsHttpConfig(&dynamic.Config{Data: "foo"})
	require.False(t, b)
}

func newConfig(config *openapi.Config) *dynamic.Config {
	c := &dynamic.Config{Data: config}
	u, _ := url.Parse("https://mokapi.io")
	c.Info.Url = u
	return c
}
