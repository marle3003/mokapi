package runtime_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/runtime/search"
	"testing"
)

func TestIndex_Kafka(t *testing.T) {
	toConfig := func(c *asyncapi3.Config) *dynamic.Config {
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
				cfg := asyncapi3test.NewConfig(asyncapi3test.WithInfo("Kafka Test server", "", ""))
				_, err := app.Kafka.Add(toConfig(cfg), enginetest.NewEngine())
				require.NoError(t, err)

				r, err := app.Search(search.Request{QueryText: "Test", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "Kafka",
						Title:     "Kafka Test server",
						Fragments: []string{"Kafka <mark>Test</mark> server"},
						Params: map[string]string{
							"type":    "kafka",
							"service": "Kafka Test server",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "Search topic",
			test: func(t *testing.T, app *runtime.App) {
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("Kafka Test server", "", ""),
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithChannelDescription("description"),
					),
				)
				_, err := app.Kafka.Add(toConfig(cfg), enginetest.NewEngine())
				require.NoError(t, err)

				r, err := app.Search(search.Request{QueryText: "description", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "Kafka",
						Domain:    "Kafka Test server",
						Title:     "Topic foo",
						Fragments: []string{"<mark>description</mark>"},
						Params: map[string]string{
							"type":    "kafka",
							"service": "Kafka Test server",
							"topic":   "foo",
						},
					},
					r.Results[0])
			},
		},
	}
	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := runtime.New(
				&static.Config{
					Api: static.Api{
						Search: static.Search{
							Enabled: true,
						},
					},
				})
			tc.test(t, app)
		})
	}
}
