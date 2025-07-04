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
				cfg := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(),
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
						Type:       "Config",
						ConfigName: "",
						Title:      "foo.yml",
						Fragments:  []string{"{&#34;<mark>name</mark>&#34;:&#34;test&#34;}"},
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
			app := runtime.New(&static.Config{})
			tc.test(t, app)
		})
	}
}
