package runtime_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"testing"
)

func TestIndex_Http(t *testing.T) {
	toConfig := func(c *openapi.Config) *dynamic.Config {
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
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
				app.AddHttp(toConfig(cfg))
				r, err := app.Search("foo")
				require.NoError(t, err)
				require.Len(t, r, 1)
			},
		},
		{
			name: "Search by version",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "1.0", ""))
				app.AddHttp(toConfig(cfg))
				r, err := app.Search("1.0")
				require.NoError(t, err)
				require.Len(t, r, 1)
			},
		},
		{
			name: "Search by path",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0",
					openapitest.WithInfo("foo", "1.0", ""),
					openapitest.WithPath("/pets", openapitest.NewPath()),
				)
				app.AddHttp(toConfig(cfg))
				r, err := app.Search("pets")
				require.NoError(t, err)
				require.Len(t, r, 1)
			},
		},
		{
			name: "Search config name and path description have same text",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0",
					openapitest.WithInfo("foo", "1.0", "a description"),
					openapitest.WithPath("/pets", openapitest.NewPath(openapitest.WithPathInfo("", "a description"))),
				)
				app.AddHttp(toConfig(cfg))
				r, err := app.Search("description")
				require.NoError(t, err)
				require.Len(t, r, 2)
			},
		},
		{
			name: "Search operation",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0",
					openapitest.WithInfo("foo", "1.0", "a description"),
					openapitest.WithPath("/pets", openapitest.NewPath(
						openapitest.WithPathInfo("", "a description"),
						openapitest.WithOperation("get", openapitest.NewOperation(
							openapitest.WithHeaderParam("foo", true, openapitest.WithParamInfo("parameter description")),
						)),
					)),
				)
				app.AddHttp(toConfig(cfg))
				r, err := app.Search("\"parameter description\"")
				require.NoError(t, err)
				require.Len(t, r, 1)
				require.Equal(t,
					runtime.SearchResult{
						Type:       "HTTP",
						ConfigName: "foo",
						Title:      "GET /pets",
						Fragments:  []string{"<mark>parameter</mark> <mark>description</mark>"}},
					r[0])
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := runtime.New(&static.Config{})
			tc.test(t, app)
		})
	}
}
