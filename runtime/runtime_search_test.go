package runtime_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/runtime"
	"testing"
)

func TestIndex_Config(t *testing.T) {
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
				r, err := app.Search("name")
				require.NoError(t, err)
				require.Len(t, r, 1)
				require.Equal(t,
					runtime.SearchResult{
						Type:      "Config",
						Domain:    "FILE",
						Title:     "file://foo.yml",
						Fragments: []string{"{&#34;<mark>name</mark>&#34;:&#34;test&#34;}"},
						Params: map[string]string{
							"type": "config",
							"id":   "64613435-3062-6462-3033-316532633233",
						},
					},
					r[0])
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
					Enabled:  true,
					Analyzer: "ngram",
					Ngram: static.NgramAnalyzer{
						Min: 3,
						Max: 5,
					},
				}}})
			tc.test(t, app)
		})
	}
}
