package runtime_test

import (
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/dynamic/provider/file"
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
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestApp_AddKafka(t *testing.T) {
	logrus.SetOutput(io.Discard)

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
				require.Equal(t, ki.Config().Servers.Len(), 1)
				require.Equal(t,
					&asyncapi3.Server{
						Host:     ":9092",
						Protocol: "kafka",
						Title:    "Mokapi Default Broker",
						Summary:  "Automatically added broker because no Kafka servers are defined in the AsyncAPI spec",
					},
					ki.Config().Servers.Lookup("mokapi").Value)

				require.Len(t, hook.Entries, 1)
				require.Equal(t, "kafka: using implicit default broker for 'foo' (inferred from bindings)", hook.Entries[0].Message)
			},
		},
		{
			name: "event store available",
			test: func(t *testing.T, app *runtime.App) {
				c := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""))
				_, err := app.Kafka.Add(getConfig("foo.bar", c), enginetest.NewEngine())
				require.NoError(t, err)

				require.NotNil(t, app.Kafka.Get("foo"))
				err = app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("kafka").WithName("foo"))
				require.NoError(t, err, "event store should be available")
			},
		},
		{
			name: "event store for topic available",
			test: func(t *testing.T, app *runtime.App) {
				c := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""), asyncapi3test.WithChannel("bar"))
				_, err := app.Kafka.Add(getConfig("foo.bar", c), enginetest.NewEngine())
				require.NoError(t, err)

				require.NotNil(t, app.Kafka.Get("foo"))
				err = app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("kafka").WithName("foo").With("path", "bar"))
				require.NoError(t, err, "event store should be available")
			},
		},
		{
			name: "event store for topic available after patching",
			test: func(t *testing.T, app *runtime.App) {
				c := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""), asyncapi3test.WithChannel("foo"))
				eng := enginetest.NewEngine()
				_, err := app.Kafka.Add(getConfig("foo.bar", c), eng)
				require.NoError(t, err)

				patch := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""), asyncapi3test.WithChannel("bar"))
				_, err = app.Kafka.Add(getConfig("foo.patch", patch), eng)
				require.NoError(t, err)

				require.NotNil(t, app.Kafka.Get("foo"))

				traits := events.NewTraits().WithNamespace("kafka").WithName("foo").With("topic", "foo")
				_ = app.Events.Push(&eventstest.Event{Name: "bar"}, traits)
				stores := app.Events.GetStores(traits)
				require.Len(t, stores, 4, "expected to find two stores for topic foo")
				require.Equal(t, stores[3].Traits, traits)
				require.Equal(t, 1, stores[3].NumEvents)

				traits = events.NewTraits().WithNamespace("kafka").WithName("foo").With("topic", "bar")
				_ = app.Events.Push(&eventstest.Event{Name: "bar"}, traits)
				stores = app.Events.GetStores(traits)
				require.Len(t, stores, 4, "expected to find two stores for topic bar")
				require.Equal(t, stores[3].Traits, traits)
				require.Equal(t, 1, stores[3].NumEvents)
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
				require.Equal(t, float64(1), m.Messages.Sum(metrics.NewQuery()))
			},
		},
		{
			name: "retrieve configs",
			test: func(t *testing.T, app *runtime.App) {
				c := asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""),
					asyncapi3test.WithChannel("bar"))
				info, err := app.Kafka.Add(getConfig("foo.bar", c), enginetest.NewEngine())
				require.NoError(t, err)

				configs := filesOf(info)
				require.Len(t, configs, 1)
				require.Equal(t, "foo.bar", configs[0].Info.Url.String())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg, &dynamictest.Reader{})
			tc.test(t, app)
		})
	}
}

func TestApp_AddKafka_Patching(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*dynamic.Config
		reader  dynamic.Reader
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
				require.Equal(t, "bar", info.Config().Info.Description)
				configs := filesOf(info)
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
				require.Equal(t, "foo", info.Config().Info.Description)
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
				require.Equal(t, "foo", info.Config().Info.Description)
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
		{
			name: "patching a referenced object",
			configs: []*dynamic.Config{
				getConfig("https://a.io/a", asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "foo", ""),
					asyncapi3test.WithChannel("bar",
						asyncapi3test.UseMessage("bar", &asyncapi3.MessageRef{Reference: dynamic.Reference[*asyncapi3.MessageRef]{Ref: "#/components/messages/bar"}}),
					),
					asyncapi3test.WithComponentMessage("bar", &asyncapi3.Message{
						Summary: "original",
					}),
				)),
				getConfig("https://mokapi.io/b", asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "bar", ""),
					asyncapi3test.WithComponentMessage("bar", &asyncapi3.Message{
						Summary: "patch",
					}),
				)),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Kafka.Get("foo")
				ch := info.Config().Channels["bar"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["bar"]
				require.NotNil(t, msg)
				require.Equal(t, "patch", msg.Value.Summary)
			},
		},
		{
			name: "using relative file path",
			reader: dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
				require.True(t, strings.HasSuffix(u.Path, "path/ref.yaml"))
				return dynamictest.NewConfig(&asyncapi3.MessageRef{
					Value: asyncapi3test.NewMessage(
						asyncapi3test.WithMessageTitle("original"),
					),
				}), nil
			}),
			configs: []*dynamic.Config{
				getFileConfig("path/foo.yaml", asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "foo", ""),
					asyncapi3test.WithChannel("bar",
						asyncapi3test.UseMessage("bar",
							&asyncapi3.MessageRef{Reference: dynamic.Reference[*asyncapi3.MessageRef]{Ref: "ref.yaml"}},
						),
					),
				)),
				getFileConfig("path/bar.yaml", asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "bar", ""),
				)),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Kafka.Get("foo")
				ch := info.Config().Channels["bar"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["bar"]
				require.NotNil(t, msg)
				require.NotNil(t, msg.Value)
				require.Equal(t, "original", msg.Value.Title)
			},
		},
		{
			name: "using relative URL path",
			reader: dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
				require.Equal(t, "https://a.io/bar.yaml", u.String())
				return dynamictest.NewConfig(&asyncapi3.MessageRef{
					Value: asyncapi3test.NewMessage(
						asyncapi3test.WithMessageTitle("original"),
					),
				}), nil
			}),
			configs: []*dynamic.Config{
				getConfig("https://a.io/a", asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "foo", ""),
					asyncapi3test.WithChannel("bar",
						asyncapi3test.UseMessage("bar",
							&asyncapi3.MessageRef{Reference: dynamic.Reference[*asyncapi3.MessageRef]{Ref: "bar.yaml"}},
						),
					),
				)),
				getConfig("https://mokapi.io/b", asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "bar", ""),
				)),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Kafka.Get("foo")
				ch := info.Config().Channels["bar"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["bar"]
				require.NotNil(t, msg)
				require.Equal(t, "original", msg.Value.Title)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r := tc.reader
			if r == nil {
				r = &dynamictest.Reader{}
			}
			cfg := &static.Config{}
			app := runtime.New(cfg, r)
			for _, c := range tc.configs {
				err := c.Data().(dynamic.Parser).Parse(c, r)
				require.NoError(t, err)
				_, err = app.Kafka.Add(c, enginetest.NewEngine())
				require.NoError(t, err)
			}
			tc.test(t, app)
		})
	}
}

func TestIsAsyncApiConfig(t *testing.T) {
	_, ok := runtime.IsAsyncApiConfig(dynamictest.NewConfig(asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", ""))))
	require.True(t, ok)
	_, ok = runtime.IsAsyncApiConfig(dynamictest.NewConfig("foo"))
	require.False(t, ok)
}

func getConfig(name string, c *asyncapi3.Config) *dynamic.Config {
	u, _ := url.Parse(name)
	cfg := dynamictest.NewConfig(c)
	cfg.Info.Url = u
	return cfg
}

func getFileConfig(name string, c *asyncapi3.Config) *dynamic.Config {
	u, _ := file.ParseUrl(name)
	cfg := dynamictest.NewConfig(c)
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

func TestKafkaStore_MergedConfigs(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "removing one of two files leaves exactly the config of the other file",
			test: func(t *testing.T, app *runtime.App) {
				a, b := twoFilesWithSameTitle("kafka")
				_, err := app.Kafka.Add(a, enginetest.NewEngine())
				require.NoError(t, err)
				info, err := app.Kafka.Add(b, enginetest.NewEngine())
				require.NoError(t, err)
				requireMergedWithB(t, info.Config())

				app.Kafka.Remove(b)

				info = app.Kafka.Get("foo")
				require.Len(t, filesOf(info), 1)
				requireOnlyFileA(t, "kafka", info.Config())
			},
		},
		{
			name: "merging does not change the files",
			test: func(t *testing.T, app *runtime.App) {
				a, b := twoFilesWithSameTitle("kafka")
				beforeA, beforeB := marshal(t, a.Data()), marshal(t, b.Data())

				_, err := app.Kafka.Add(a, enginetest.NewEngine())
				require.NoError(t, err)
				_, err = app.Kafka.Add(b, enginetest.NewEngine())
				require.NoError(t, err)

				require.Equal(t, beforeA, marshal(t, a.Data()))
				require.Equal(t, beforeB, marshal(t, b.Data()))
			},
		},
		{
			name: "the default broker is not added to the file",
			test: func(t *testing.T, app *runtime.App) {
				c := asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo"), asyncapi3test.WithChannel("bar"))

				info, err := app.Kafka.Add(getConfig("https://mokapi.io/a.yml", c), enginetest.NewEngine())

				require.NoError(t, err)
				require.NotNil(t, info.Config().Servers.Lookup("mokapi"))
				require.Equal(t, 0, c.Servers.Len())
			},
		},
		{
			name: "merging AsyncAPI 2.x files does not change them",
			test: func(t *testing.T, app *runtime.App) {
				a, b := twoAsyncApi2FilesWithSameTitle(t)

				_, err := app.Kafka.Add(a, enginetest.NewEngine())
				require.NoError(t, err)
				info, err := app.Kafka.Add(b, enginetest.NewEngine())
				require.NoError(t, err)

				requireExtraPropertyOnlyIn(t, info.Config(), a, b)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, runtime.New(&static.Config{}, &dynamictest.Reader{}))
		})
	}
}

// Two files with the same title define the same recursive schema (Node.properties.child
// refers to Node). Merging them must not overflow the stack.
func TestApp_AddKafka_SameCyclicSchemaInTwoFiles(t *testing.T) {
	logrus.SetOutput(io.Discard)
	newNode := func() *schema.Schema {
		props := &schema.Schemas{}
		props.Set("name", schematest.New("string"))
		node := schematest.New("object", schematest.WithTitle("Node"))
		node.Properties = props
		props.Set("child", &schema.Schema{Type: schema.Types{"object"}, Properties: props})
		return node
	}
	file := func(name string) *dynamic.Config {
		return getConfig("https://mokapi.io/"+name, asyncapi3test.NewConfig(
			asyncapi3test.WithTitle("foo"),
			asyncapi3test.WithComponentSchema("Node", newNode()),
		))
	}

	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	_, err := app.Kafka.Add(file("a.yml"), enginetest.NewEngine())
	require.NoError(t, err)
	info, err := app.Kafka.Add(file("b.yml"), enginetest.NewEngine())
	require.NoError(t, err)

	require.Contains(t, info.Config().Components.Schemas, "Node")
}

func TestKafkaStore_Remove(t *testing.T) {
	logrus.SetOutput(io.Discard)

	t.Run("unknown config does not panic and keeps the store usable", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		unknown := getConfig("https://mokapi.io/unknown.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("unknown")))
		known := getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))

		requireStoreUsableAfterRemoveOfUnknownConfig(t,
			func() { app.Kafka.Remove(unknown) },
			func() { _, _ = app.Kafka.Add(known, enginetest.NewEngine()) },
		)
	})

	t.Run("two files removed at the same time", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a, b := twoFilesWithSameTitle("kafka")
		_, err := app.Kafka.Add(a, enginetest.NewEngine())
		require.NoError(t, err)
		_, err = app.Kafka.Add(b, enginetest.NewEngine())
		require.NoError(t, err)

		removeConcurrently(app.Kafka.Remove, a, b)

		require.Equal(t, 0, app.Kafka.Len())
	})
}

func TestKafkaStore_ConcurrentAccess(t *testing.T) {
	logrus.SetOutput(io.Discard)

	t.Run("reader of the info does not race with reload of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		reload := func(channel string) *dynamic.Config {
			return getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(
				asyncapi3test.WithTitle("foo"),
				asyncapi3test.WithServer("a", "kafka", ":1111"),
				asyncapi3test.WithChannel(channel),
			))
		}
		_, err := app.Kafka.Add(reload("first"), enginetest.NewEngine())
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if _, err := app.Kafka.Add(reload("second"), enginetest.NewEngine()); err != nil {
					t.Error(err)
					return
				}
			}
		}()
		go func() {
			defer wg.Done()
			// what the dashboard API handler does: get the info, then read the config
			for i := 0; i < 200; i++ {
				info := app.Kafka.Get("foo")
				cfg := info.Config()
				_ = cfg.Info.Name
				for range cfg.Channels {
				}
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})

	t.Run("reader does not race with remove", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a, b := twoFilesWithSameTitle("kafka")
		info, err := app.Kafka.Add(a, enginetest.NewEngine())
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if _, err := app.Kafka.Add(b, enginetest.NewEngine()); err != nil {
					t.Error(err)
					return
				}
				app.Kafka.Remove(b)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if cfg := info.Config(); cfg != nil {
					_ = cfg.Info.Name
				}
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})
}

func TestKafkaInfo_Snapshot(t *testing.T) {
	logrus.SetOutput(io.Discard)

	fileWithChannel := func(channel string) *dynamic.Config {
		return getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(
			asyncapi3test.WithTitle("foo"),
			asyncapi3test.WithServer("a", "kafka", ":1111"),
			asyncapi3test.WithChannel(channel),
		))
	}

	t.Run("snapshot is not changed by a reload of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		info, err := app.Kafka.Add(fileWithChannel("first"), enginetest.NewEngine())
		require.NoError(t, err)
		snapshot := info.Config()

		_, err = app.Kafka.Add(fileWithChannel("second"), enginetest.NewEngine())
		require.NoError(t, err)

		require.Contains(t, snapshot.Channels, "first")
		require.NotContains(t, snapshot.Channels, "second")
		require.NotSame(t, snapshot, info.Config())
		require.Contains(t, info.Config().Channels, "second")
	})

	t.Run("published config is not the config of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo"), asyncapi3test.WithChannel("bar")))

		info, err := app.Kafka.Add(file, enginetest.NewEngine())

		require.NoError(t, err)
		require.NotSame(t, file.Data(), info.Config())
		require.NotNil(t, info.Config().Servers.Lookup("mokapi"), "default broker is added to the published config")
		require.Equal(t, 0, file.Data().(*asyncapi3.Config).Servers.Len(), "but not to the file")
	})

	t.Run("config is nil after the last file is removed", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := fileWithChannel("first")
		info, err := app.Kafka.Add(file, enginetest.NewEngine())
		require.NoError(t, err)

		app.Kafka.Remove(file)

		require.Nil(t, info.Config())
		require.NotNil(t, info.Store, "store is not removed with the last file")
		require.Nil(t, app.Kafka.Get("foo"))
	})

	t.Run("Configs returns a copy", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a, b := twoFilesWithSameTitle("kafka")
		_, err := app.Kafka.Add(a, enginetest.NewEngine())
		require.NoError(t, err)
		info, err := app.Kafka.Add(b, enginetest.NewEngine())
		require.NoError(t, err)

		configs := filesOf(info)
		require.Len(t, configs, 2)
		configs[0] = nil

		require.NotContains(t, filesOf(info), (*dynamic.Config)(nil))
		require.Len(t, filesOf(info), 2)
	})

	t.Run("Configs does not race with Add", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a, b := twoFilesWithSameTitle("kafka")
		info, err := app.Kafka.Add(a, enginetest.NewEngine())
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if _, err := app.Kafka.Add(b, enginetest.NewEngine()); err != nil {
					t.Error(err)
					return
				}
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})
}

func TestKafkaStore_ListSetAndLen(t *testing.T) {
	logrus.SetOutput(io.Discard)

	t.Run("List returns all infos and none after the last Remove", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		foo := getConfig("https://mokapi.io/foo.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))
		bar := getConfig("https://mokapi.io/bar.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("bar")))
		_, err := app.Kafka.Add(foo, enginetest.NewEngine())
		require.NoError(t, err)
		_, err = app.Kafka.Add(bar, enginetest.NewEngine())
		require.NoError(t, err)

		names := map[string]bool{}
		for _, info := range app.Kafka.List() {
			names[info.Config().Info.Name] = true
		}
		require.Equal(t, map[string]bool{"foo": true, "bar": true}, names)
		require.Equal(t, 2, app.Kafka.Len())

		app.Kafka.Remove(foo)
		app.Kafka.Remove(bar)

		require.Empty(t, app.Kafka.List())
		require.Equal(t, 0, app.Kafka.Len())
	})

	t.Run("nil store has no infos", func(t *testing.T) {
		var s *runtime.KafkaStore

		require.Nil(t, s.List())
		require.Equal(t, 0, s.Len())
	})

	t.Run("Set registers an info that Get returns", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		cfg := asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo"))
		info := runtime.NewKafkaInfo(cfg, nil)

		app.Kafka.Set("foo", info)

		require.Same(t, info, app.Kafka.Get("foo"))
		require.Same(t, cfg, info.Config())
		require.Len(t, app.Kafka.List(), 1)
	})

	t.Run("Add rejects a config that is not AsyncAPI", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})

		_, err := app.Kafka.Add(dynamictest.NewConfig("foo"), enginetest.NewEngine())

		require.EqualError(t, err, "no Kafka config found")
	})
}

func TestShouldStartKafka(t *testing.T) {
	testcases := []struct {
		name string
		data any
		mode runtime.KafkaMode
	}{
		{"not AsyncAPI", "foo", runtime.KafkaNone},
		{"kafka server", asyncapi3test.NewConfig(asyncapi3test.WithServer("a", "kafka", ":9092")), runtime.KafkaExplicit},
		{"only a server of another protocol", asyncapi3test.NewConfig(asyncapi3test.WithServer("a", "mqtt", ":1883"), asyncapi3test.WithChannel("foo")), runtime.KafkaNone},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, mode := runtime.ShouldStartKafka(dynamictest.NewConfig(tc.data))

			require.Equal(t, tc.mode, mode)
		})
	}
}
