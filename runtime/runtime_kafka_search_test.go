package runtime_test

import (
	"context"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime"
	"mokapi/runtime/search"
	"mokapi/safe"
	"testing"

	"github.com/stretchr/testify/require"
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

				var r search.Result
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "Test", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "Kafka",
						Title:     "Kafka Test server",
						Fragments: []string{"Kafka <mark>Test</mark> server"},
						Params: map[string]string{
							"type":    "kafka",
							"service": "Kafka Test server",
							"topics":  "0",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "config should be removed from index",
			test: func(t *testing.T, app *runtime.App) {
				cfg := asyncapi3test.NewConfig(asyncapi3test.WithInfo("Kafka Test server", "", ""))
				_, err := app.Kafka.Add(toConfig(cfg), enginetest.NewEngine())
				require.NoError(t, err)

				var r search.Result
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "Test", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)

				app.Kafka.Remove(toConfig(cfg))
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
					asyncapi3test.WithInfo("Kafka Test server", "", ""),
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

				_, err := app.Kafka.Add(toConfig(cfg), enginetest.NewEngine())
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
						Type:      "Kafka",
						Domain:    "Kafka Test server",
						Title:     "Topic foo",
						Fragments: []string{"<mark>first</mark>"},
						Params: map[string]string{
							"type":    "kafka",
							"service": "Kafka Test server",
							"topic":   "foo",
						},
					},
					r.Results[0])

				r, err = app.Search(search.Request{QueryText: "second", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "Kafka",
						Domain:    "Kafka Test server",
						Title:     "Topic address-name",
						Fragments: []string{"<mark>second</mark>"},
						Params: map[string]string{
							"type":    "kafka",
							"service": "Kafka Test server",
							"topic":   "address-name",
						},
					},
					r.Results[0])

				r, err = app.Search(search.Request{QueryText: "third", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "Kafka",
						Domain:    "Kafka Test server",
						Title:     "Topic yuh",
						Fragments: []string{"<mark>third</mark>"},
						Params: map[string]string{
							"type":    "kafka",
							"service": "Kafka Test server",
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

func TestIndex_Kafka_Event(t *testing.T) {
	api := asyncapi3test.NewConfig(
		asyncapi3test.WithInfo("Kafka Test Events", "", ""),
		asyncapi3test.WithChannel("events"),
	)
	cfg := &dynamic.Config{
		Info: dynamictest.NewConfigInfo(),
		Data: api,
	}

	testcases := []struct {
		name string
		test func(t *testing.T, s *store.Store, app *runtime.App)
	}{
		{
			name: "search event by key",
			test: func(t *testing.T, s *store.Store, app *runtime.App) {

				wr, err := s.Topic("events").Partitions[0].Write(kafka.RecordBatch{Records: []*kafka.Record{
					{
						Key: kafka.NewBytes([]byte("foo")),
					},
				}})
				require.NoError(t, err)
				require.Len(t, wr.Records, 0)

				r, err := waitSearchResult(t, func() (search.Result, error) {
					return app.Search(search.Request{QueryText: "+key:foo +type:event", Limit: 10})
				}, 1)

				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t, "Event", r.Results[0].Type)
				require.Equal(t, "Kafka Test Events", r.Results[0].Domain)
				require.Equal(t, "foo", r.Results[0].Title)
				require.Len(t, r.Results[0].Fragments, 2)
				require.Contains(t, r.Results[0].Fragments, "<mark>foo</mark>")
				require.Contains(t, r.Results[0].Fragments, "<mark>event</mark>")
				require.Len(t, r.Results[0].Params, 7)
				require.Equal(t, "event", r.Results[0].Params["type"])
				require.Equal(t, "kafka", r.Results[0].Params["traits.namespace"])
				require.Equal(t, "Kafka Test Events", r.Results[0].Params["traits.name"])
				require.Equal(t, "message", r.Results[0].Params["traits.type"])
				require.Equal(t, "events", r.Results[0].Params["traits.topic"])
				require.Equal(t, "0", r.Results[0].Params["traits.partition"])
				require.Contains(t, r.Results[0].Params, "id")
				require.NotEmpty(t, r.Results[0].Time)
			},
		},
		{
			name: "search event by header",
			test: func(t *testing.T, s *store.Store, app *runtime.App) {

				wr, err := s.Topic("events").Partitions[0].Write(kafka.RecordBatch{Records: []*kafka.Record{
					{
						Key: kafka.NewBytes([]byte("foo")),
						Headers: []kafka.RecordHeader{
							{
								Key:   "header-key",
								Value: []byte("bar"),
							},
						},
					},
				}})
				require.NoError(t, err)
				require.Len(t, wr.Records, 0)

				r, err := waitSearchResult(t, func() (search.Result, error) {
					return app.Search(search.Request{QueryText: `+"header-key" +type:event`, Limit: 10})
				}, 1)

				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t, "Event", r.Results[0].Type)
				require.Equal(t, "Kafka Test Events", r.Results[0].Domain)
				require.Equal(t, "foo", r.Results[0].Title)
				require.Len(t, r.Results[0].Fragments, 2)
				require.Contains(t, r.Results[0].Fragments, "<mark>header</mark>-<mark>key</mark>")
				require.Contains(t, r.Results[0].Fragments, "<mark>event</mark>")
				require.Len(t, r.Results[0].Params, 7)
				require.Equal(t, "event", r.Results[0].Params["type"])
				require.Equal(t, "kafka", r.Results[0].Params["traits.namespace"])
				require.Equal(t, "Kafka Test Events", r.Results[0].Params["traits.name"])
				require.Equal(t, "message", r.Results[0].Params["traits.type"])
				require.Equal(t, "events", r.Results[0].Params["traits.topic"])
				require.Equal(t, "0", r.Results[0].Params["traits.partition"])
				require.Contains(t, r.Results[0].Params, "id")
				require.NotEmpty(t, r.Results[0].Time)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app := runtime.New(
				&static.Config{
					Api: static.Api{
						Search: static.Search{
							Enabled:  true,
							InMemory: true,
						},
					},
				}, &dynamictest.Reader{})

			info, err := app.Kafka.Add(cfg, enginetest.NewEngine())
			require.NoError(t, err)

			pool := safe.NewPool(context.Background())
			app.Start(pool)
			defer pool.Stop()

			tc.test(t, info.Store, app)
		})
	}
}
