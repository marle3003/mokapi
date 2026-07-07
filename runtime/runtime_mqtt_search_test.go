package runtime_test

import (
	"context"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/runtime/search"
	"mokapi/safe"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestIndex_Mqtt(t *testing.T) {
	logrus.SetOutput(io.Discard)

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
				cfg := asyncapi3test.NewConfig(asyncapi3test.WithInfo("MQTT Test server", "", ""))
				_, err := app.Mqtt.Add(toConfig(cfg), enginetest.NewEngine())
				require.NoError(t, err)

				var r search.Result
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "Test", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "MQTT",
						Title:     "MQTT Test server",
						Fragments: []string{"MQTT <mark>Test</mark> server"},
						Params: map[string]string{
							"type":    "mqtt",
							"service": "MQTT Test server",
							"topics":  "0",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "config should be removed from index",
			test: func(t *testing.T, app *runtime.App) {
				cfg := asyncapi3test.NewConfig(asyncapi3test.WithInfo("MQTT Test server", "", ""))
				_, err := app.Mqtt.Add(toConfig(cfg), enginetest.NewEngine())
				require.NoError(t, err)

				var r search.Result
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "Test", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)

				app.Mqtt.Remove(toConfig(cfg))
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "Test", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 0
				})
				require.Len(t, r.Results, 0)
			},
		},
		{
			name: "Search topic",
			test: func(t *testing.T, app *runtime.App) {
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("MQTT Test server", "", ""),
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithChannelDescription("first"),
					),
				)

				second := asyncapi3test.NewChannel(
					asyncapi3test.WithChannelDescription("second"),
					asyncapi3test.WithChannelAddress("address-name"),
				)
				cfg.Channels["bar"] = &asyncapi3.ChannelRef{Value: second}

				third := asyncapi3test.NewChannel(
					asyncapi3test.WithChannelDescription("third"),
				)
				cfg.Channels["yuh"] = &asyncapi3.ChannelRef{Value: third}

				_, err := app.Mqtt.Add(toConfig(cfg), enginetest.NewEngine())
				require.NoError(t, err)

				var r search.Result
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "first", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:        "MQTT",
						Domain:      "MQTT Test server",
						Title:       "Topic foo",
						Description: `first`,
						Fragments:   []string{"<mark>first</mark>"},
						Params: map[string]string{
							"type":    "mqtt",
							"service": "MQTT Test server",
							"topic":   "foo",
						},
					},
					r.Results[0])

				r, err = app.Search(search.Request{QueryText: "second", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:        "MQTT",
						Domain:      "MQTT Test server",
						Title:       "Topic address-name",
						Description: `second`,
						Fragments:   []string{"<mark>second</mark>"},
						Params: map[string]string{
							"type":    "mqtt",
							"service": "MQTT Test server",
							"topic":   "address-name",
						},
					},
					r.Results[0])

				r, err = app.Search(search.Request{QueryText: "third", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:        "MQTT",
						Domain:      "MQTT Test server",
						Title:       "Topic yuh",
						Description: `third`,
						Fragments:   []string{"<mark>third</mark>"},
						Params: map[string]string{
							"type":    "mqtt",
							"service": "MQTT Test server",
							"topic":   "yuh",
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
							Enabled:  true,
							InMemory: true,
						},
					},
				}, &dynamictest.Reader{})

			pool := safe.NewPool(context.Background())
			app.Start(pool)
			defer pool.Stop()

			tc.test(t, app)
		})
	}
}
