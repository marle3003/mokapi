package runtime_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/search"
	"testing"
)

func TestIndex_Config(t *testing.T) {
	toConfig := func(c any) *dynamic.Config {
		cfg := &dynamic.Config{
			Info: dynamictest.NewConfigInfo(),
			Data: c,
		}
		return cfg
	}

	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "Search by name",
			test: func(t *testing.T, app *runtime.App) {
				info := dynamictest.NewConfigInfo()
				info.Provider = "file"

				cfg := &dynamic.Config{
					Info: info,
					Raw:  []byte(`{"name":"test"}`),
				}

				app.UpdateConfig(dynamic.ConfigEvent{
					Name:   "",
					Config: cfg,
					Event:  dynamic.Create,
				})
				r, err := app.Search(search.Request{QueryText: "name", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "Config",
						Domain:    "FILE",
						Title:     "file://foo.yml",
						Fragments: []string{"{&#34;<mark>name</mark>&#34;:&#34;test&#34;}"},
						Params: map[string]string{
							"type": "config",
							"id":   "64613435-3062-6462-3033-316532633233",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "kafka and http indexed",
			test: func(t *testing.T, app *runtime.App) {
				h := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
				app.AddHttp(toConfig(h))
				k := asyncapitest.NewConfig(asyncapitest.WithInfo("foo", "", ""))
				_, err := app.Kafka.Add(toConfig(k), enginetest.NewEngine())
				require.NoError(t, err)

				r, err := app.Search(search.Request{QueryText: "foo", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 2)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := runtime.New(&static.Config{Api: static.Api{
				Search: static.Search{
					Enabled: true,
				}}})
			tc.test(t, app)
		})
	}
}
