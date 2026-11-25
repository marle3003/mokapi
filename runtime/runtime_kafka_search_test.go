package runtime_test

import (
	"context"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	openapiTest "mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime"
	"mokapi/runtime/search"
	"mokapi/safe"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestIndex_Kafka(t *testing.T) {
	logrus.SetOutput(io.Discard)

	toConfig := func(c *asyncapi3.Config) *dynamic.Config {
		cfg := dynamictest.NewConfig(c, dynamictest.WithInfo(dynamictest.NewConfigInfo()))
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
						Type:        "Kafka",
						Domain:      "Kafka Test server",
						Title:       "Topic foo",
						Description: `first`,
						Fragments:   []string{"<mark>first</mark>"},
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
						Type:        "Kafka",
						Domain:      "Kafka Test server",
						Title:       "Topic address-name",
						Description: `second`,
						Fragments:   []string{"<mark>second</mark>"},
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
						Type:        "Kafka",
						Domain:      "Kafka Test server",
						Title:       "Topic yuh",
						Description: `third`,
						Fragments:   []string{"<mark>third</mark>"},
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
	logrus.SetOutput(io.Discard)

	api := asyncapi3test.NewConfig(
		asyncapi3test.WithInfo("Kafka Test Events", "", ""),
		asyncapi3test.WithChannel("events"),
	)
	cfg := dynamictest.NewConfig(api, dynamictest.WithInfo(dynamictest.NewConfigInfo()))

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

func TestIndex_Kafka_MergedFiles(t *testing.T) {
	logrus.SetOutput(io.Discard)

	file := func(url, channel string) *dynamic.Config {
		return dynamictest.NewConfig(asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("foo", "", ""),
			asyncapi3test.WithChannel(channel),
		), dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl(url))))
	}
	found := func(t *testing.T, app *runtime.App, text string) bool {
		r, err := app.Search(search.Request{QueryText: text, Limit: 10})
		require.NoError(t, err)
		return len(r.Results) > 0
	}

	app := runtime.New(&static.Config{Api: static.Api{Search: static.Search{Enabled: true, InMemory: true}}}, &dynamictest.Reader{})
	pool := safe.NewPool(context.Background())
	app.Start(pool)
	defer pool.Stop()

	_, err := app.Kafka.Add(file("https://mokapi.io/a.yml", "alpha"), enginetest.NewEngine())
	require.NoError(t, err)
	b := file("https://mokapi.io/b.yml", "beta")
	_, err = app.Kafka.Add(b, enginetest.NewEngine())
	require.NoError(t, err)
	waitSearchIndex(t, func() bool { return found(t, app, "alpha") && found(t, app, "beta") })

	app.Kafka.Remove(b)

	waitSearchIndex(t, func() bool {
		return found(t, app, "foo") && found(t, app, "alpha") && !found(t, app, "beta")
	})
}

func TestIndex_Kafka_MessagePayload(t *testing.T) {
	logrus.SetOutput(io.Discard)

	cfg := dynamictest.NewConfig(asyncapi3test.NewConfig(
		asyncapi3test.WithInfo("foo", "", ""),
		asyncapi3test.WithChannel("orders",
			asyncapi3test.WithMessage("order",
				asyncapi3test.WithPayload(schematest.New("object",
					schematest.WithProperty("id", schematest.New("string", schematest.WithDescription("quokkadescription"))),
				)),
			),
		),
	), dynamictest.WithInfo(dynamictest.NewConfigInfo()))

	app := runtime.New(&static.Config{Api: static.Api{Search: static.Search{Enabled: true, InMemory: true}}}, &dynamictest.Reader{})
	pool := safe.NewPool(context.Background())
	app.Start(pool)
	defer pool.Stop()

	_, err := app.Kafka.Add(cfg, enginetest.NewEngine())
	require.NoError(t, err)

	var r search.Result
	waitSearchIndex(t, func() bool {
		r, err = app.Search(search.Request{QueryText: "quokkadescription", Limit: 10})
		require.NoError(t, err)
		return len(r.Results) == 1
	})
	require.Len(t, r.Results, 1)
}

func TestIndex_Kafka_MessageSchemas(t *testing.T) {
	// a record that references itself, as Avro allows for linked structures
	recursiveAvro := &avro.Schema{Type: []interface{}{"record"}, Name: "Node"}
	recursiveAvro.Fields = []*avro.Schema{
		{Name: "id", Type: []interface{}{"string"}, Doc: "avrodescription"},
		{Name: "next", Type: []interface{}{recursiveAvro}},
	}
	avroFormat := "application/vnd.apache.avro;version=1.9.0"

	testcases := []struct {
		name    string
		message asyncapi3test.MessageOptions
		term    string
	}{
		{
			name: "avro payload",
			message: asyncapi3test.WithPayloadMulti(avroFormat, &asyncapi3.AvroRef{Schema: &avro.Schema{
				Type:   []interface{}{"record"},
				Fields: []*avro.Schema{{Name: "id", Type: []interface{}{"string"}, Doc: "avrodescription"}},
			}}),
			term: "avrodescription",
		},
		{
			name:    "recursive avro payload",
			message: asyncapi3test.WithPayloadMulti(avroFormat, &asyncapi3.AvroRef{Schema: recursiveAvro}),
			term:    "avrodescription",
		},
		{
			name:    "avro without schema",
			message: asyncapi3test.WithPayloadMulti(avroFormat, &asyncapi3.AvroRef{}),
			term:    "orders",
		},
		{
			name: "openapi payload",
			message: asyncapi3test.WithPayloadOpenAPI(openapiTest.New("object",
				openapiTest.WithProperty("id", openapiTest.New("string", openapiTest.WithDescription("openapidescription"))),
			)),
			term: "openapidescription",
		},
		{
			name: "headers",
			message: asyncapi3test.WithHeaders(schematest.New("object",
				schematest.WithProperty("trace", schematest.New("string", schematest.WithDescription("headerdescription"))),
			)),
			term: "headerdescription",
		},
		{
			// guards against a nil-payload panic while indexing
			name:    "message without payload",
			message: asyncapi3test.WithMessageName("order"),
			term:    "orders",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			logrus.SetOutput(io.Discard)
			hook := test.NewGlobal()
			t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks)) })

			cfg := dynamictest.NewConfig(asyncapi3test.NewConfig(
				asyncapi3test.WithInfo("foo", "", ""),
				asyncapi3test.WithChannel("orders", asyncapi3test.WithMessage("order", tc.message)),
			), dynamictest.WithInfo(dynamictest.NewConfigInfo()))

			app := runtime.New(&static.Config{Api: static.Api{Search: static.Search{Enabled: true, InMemory: true}}}, &dynamictest.Reader{})
			pool := safe.NewPool(context.Background())
			app.Start(pool)
			defer pool.Stop()

			_, err := app.Kafka.Add(cfg, enginetest.NewEngine())
			require.NoError(t, err)

			var r search.Result
			waitSearchIndex(t, func() bool {
				r, err = app.Search(search.Request{QueryText: tc.term, Limit: 10})
				require.NoError(t, err)
				return len(r.Results) == 1
			})
			require.Len(t, r.Results, 1)
			for _, e := range hook.AllEntries() {
				require.NotContains(t, e.Message, "indexing message")
			}
		})
	}
}
