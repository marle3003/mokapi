package runtime_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/produce"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"net/url"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestApp_AddKafka(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "add default server if no is specified",
			test: func(t *testing.T, app *runtime.App) {
				hook := test.NewGlobal()

				c := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""))
				ki, err := app.Kafka.Add(getConfig("foo.bar", c), enginetest.NewEngine())
				require.Nil(t, err)
				require.Equal(t, ki.Servers.Len(), 1)
				require.Equal(t,
					&asyncapi3.Server{
						Host:     ":9092",
						Protocol: "kafka",
						Title:    "Mokapi Default Broker",
						Summary:  "Automatically added broker because no servers are defined in the AsyncAPI spec",
					},
					ki.Servers.Lookup("mokapi").Value)

				require.Len(t, hook.Entries, 1)
				require.Equal(t, "no servers defined in AsyncAPI spec — using default Mokapi broker for cluster 'foo'", hook.Entries[0].Message)
			},
		},
		{
			name: "event store available",
			test: func(t *testing.T, app *runtime.App) {
				c := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""))
				app.Kafka.Add(getConfig("foo.bar", c), enginetest.NewEngine())

				require.NotNil(t, app.Kafka.Get("foo"))
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("kafka").WithName("foo"))
				require.NoError(t, err, "event store should be available")
			},
		},
		{
			name: "event store for topic available",
			test: func(t *testing.T, app *runtime.App) {
				c := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""), asyncapi3test.WithChannel("bar"))
				app.Kafka.Add(getConfig("foo.bar", c), enginetest.NewEngine())

				require.NotNil(t, app.Kafka.Get("foo"))
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("kafka").WithName("foo").With("path", "bar"))
				require.NoError(t, err, "event store should be available")
			},
		},
		{
			name: "event store for topic available after patching",
			test: func(t *testing.T, app *runtime.App) {
				c := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""), asyncapi3test.WithChannel("foo"))
				app.Kafka.Add(getConfig("foo.bar", c), enginetest.NewEngine())

				patch := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""), asyncapi3test.WithChannel("bar"))
				ki := app.Kafka.Get(patch.Info.Name)
				ki.AddConfig(getConfig("foo.patch", patch))

				require.NotNil(t, app.Kafka.Get("foo"))

				traits := events.NewTraits().WithNamespace("kafka").WithName("foo").With("topic", "foo")
				_ = app.Events.Push(&eventstest.Event{Name: "bar"}, traits)
				stores := app.Events.GetStores(traits)
				require.Len(t, stores, 3, "expected to find two stores for topic foo")
				require.Equal(t, stores[2].Traits, traits)
				require.Equal(t, 1, stores[2].NumEvents)

				traits = events.NewTraits().WithNamespace("kafka").WithName("foo").With("topic", "bar")
				_ = app.Events.Push(&eventstest.Event{Name: "bar"}, traits)
				stores = app.Events.GetStores(traits)
				require.Len(t, stores, 3, "expected to find two stores for topic bar")
				require.Equal(t, stores[2].Traits, traits)
				require.Equal(t, 1, stores[2].NumEvents)
			},
		},
		{
			name: "monitor",
			test: func(t *testing.T, app *runtime.App) {
				c := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""),
					asyncapi3test.WithChannel("bar"))
				info, err := app.Kafka.Add(getConfig("foo.bar", c), enginetest.NewEngine())
				require.NoError(t, err)
				m := monitor.NewKafka()
				h := info.Handler(m)

				rr := kafkatest.NewRecorder()
				h.ServeMessage(rr, newProduceMessage("bar"))

				// wait for update monitor
				time.Sleep(500 * time.Millisecond)
				require.Equal(t, float64(1), m.Messages.Sum())
			},
		},
		{
			name: "retrieve configs",
			test: func(t *testing.T, app *runtime.App) {
				c := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""),
					asyncapi3test.WithChannel("bar"))
				info, err := app.Kafka.Add(getConfig("foo.bar", c), enginetest.NewEngine())
				require.NoError(t, err)

				configs := info.Configs()
				require.Len(t, configs, 1)
				require.Equal(t, "foo.bar", configs[0].Info.Url.String())
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

func TestApp_AddKafka_Patching(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*dynamic.Config
		test    func(t *testing.T, app *runtime.App)
	}{
		{
			name: "overwrite value",
			configs: []*dynamic.Config{
				getConfig("https://mokapi.io/a", asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "foo", ""))),
				getConfig("https://mokapi.io/b", asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "bar", ""))),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Kafka.Get("foo")
				require.Equal(t, "bar", info.Info.Description)
				configs := info.Configs()
				require.Len(t, configs, 2)
			},
		},
		{
			name: "a is patched with b",
			configs: []*dynamic.Config{
				getConfig("https://mokapi.io/b", asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "foo", ""))),
				getConfig("https://mokapi.io/a", asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "bar", ""))),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Kafka.Get("foo")
				require.Equal(t, "foo", info.Info.Description)
			},
		},
		{
			name: "order only by filename",
			configs: []*dynamic.Config{
				getConfig("https://a.io/b", asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "foo", ""))),
				getConfig("https://mokapi.io/a", asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "bar", ""))),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Kafka.Get("foo")
				require.Equal(t, "foo", info.Info.Description)
			},
		},
		{
			name: "patch does not reset events and metrics",
			configs: []*dynamic.Config{
				getConfig("https://a.io/a",
					asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "foo", ""),
						asyncapi3test.WithChannel("bar"),
					),
				),
			},
			test: func(t *testing.T, app *runtime.App) {
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("kafka").WithName("foo").With("topic", "bar"))
				require.NoError(t, err)
				e := app.Events.GetEvents(events.NewTraits().WithNamespace("kafka"))
				require.Len(t, e, 1)
				app.Monitor.Kafka.Messages.WithLabel("foo", "bar").Add(1)

				_, err = app.Kafka.Add(getConfig("https://a.io/b",
					asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "foo", ""),
						asyncapi3test.WithChannel("bar"),
					),
				), enginetest.NewEngine())
				require.NoError(t, err)

				e = app.Events.GetEvents(events.NewTraits().WithNamespace("kafka"))
				require.Len(t, e, 1)
				v := app.Monitor.Kafka.Messages.WithLabel("foo", "bar").Value()
				require.Equal(t, float64(1), v)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg)
			for _, c := range tc.configs {
				app.Kafka.Add(c, enginetest.NewEngine())
			}
			tc.test(t, app)
		})
	}
}

func TestIsAsyncApiConfig(t *testing.T) {
	_, ok := runtime.IsAsyncApiConfig(&dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", ""))})
	require.True(t, ok)
	_, ok = runtime.IsAsyncApiConfig(&dynamic.Config{Data: "foo"})
	require.False(t, ok)
}

func getConfig(name string, c *asyncapi3.Config) *dynamic.Config {
	u, _ := url.Parse(name)
	cfg := &dynamic.Config{Data: c}
	cfg.Info.Url = u
	return cfg
}

func newProduceMessage(topic string) *kafka.Request {
	return kafkatest.NewRequest("kafkatest", 3, &produce.Request{
		Topics: []produce.RequestTopic{
			{
				Name: topic, Partitions: []produce.RequestPartition{
					{
						Record: kafka.RecordBatch{
							Records: []*kafka.Record{
								{
									Offset:  0,
									Time:    time.Now(),
									Key:     kafka.NewBytes([]byte("foo-1")),
									Value:   kafka.NewBytes([]byte("bar-1")),
									Headers: nil,
								},
							},
						},
					},
				},
			},
		},
	})
}
