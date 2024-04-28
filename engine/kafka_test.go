package engine

import (
	"bytes"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	bindings "mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/produce"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"net/url"
	"testing"
	"time"
)

func TestKafkaClient_Produce(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App, s *store.Store, engine *Engine)
	}{
		{
			name: "random message",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *Engine) {
				err := engine.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'foo', cluster: 'foo' })
					}
				`))
				require.NoError(t, err)
				b, errCode := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Equal(t, "XidZuoWq ", string(readBytes(b.Records[0].Key)))
				require.Equal(t, "\"\"", string(readBytes(b.Records[0].Value)))
			},
		},
		{
			name: "non random values",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *Engine) {
				hook := test.NewGlobal()

				err := engine.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						const result = produce({ 
							topic: 'foo',
							partition: 0,
							key: 'foo',
							value: 'bar',
							headers: { version: '1.0' },
						})
						console.log(result)
					}
				`))
				require.NoError(t, err)
				b, errCode := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Equal(t, "foo", string(readBytes(b.Records[0].Key)))
				require.Equal(t, `"bar"`, string(readBytes(b.Records[0].Value)))
				require.Equal(t, "version", b.Records[0].Headers[0].Key)
				require.Equal(t, []byte("1.0"), b.Records[0].Headers[0].Value)

				require.Equal(t, `{"cluster":"foo","topic":"foo","messages":[{"key":"foo","value":"\"bar\"","offset":0,"headers":{"version":"1.0"},"partition":0}]}`, hook.LastEntry().Message)
			},
		},
		{
			name: "to partition 5",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *Engine) {
				err := engine.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'bar', partition: 5 })
					}
				`))
				require.NoError(t, err)
				b, errCode := s.Topic("bar").Partition(5).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
			},
		},
		{
			name: "multiple clusters",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *Engine) {
				for i := 0; i < 10; i++ {
					app.AddKafka(getConfig(
						asyncapitest.NewConfig(asyncapitest.WithInfo(fmt.Sprintf("x%v", i), "", ""))), enginetest.NewEngine())
				}

				err := engine.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'foo' })
					}
				`))
				require.NoError(t, err)

				b, errCode := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Len(t, b.Records, 1)
			},
		},
		{
			name: "trigger event",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *Engine) {
				err := engine.AddScript(newScript("test.js", `
					import { on } from 'mokapi'
					export default function() {
						on('kafka', function(message) {
							console.log(message)
							message.value = 'mokapi'
							message.headers = { version: '1.0' }
						})
					}
				`))
				require.NoError(t, err)

				hook := test.NewGlobal()

				sendMessage(s, nil)
				require.Equal(t, `{"offset":0,"key":"foo-1","value":"\"bar-1\"","headers":{}}`, hook.LastEntry().Message)

				b, errCode := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Equal(t, "mokapi", string(readBytes(b.Records[0].Value)))
				require.Len(t, b.Records[0].Headers, 1)
				require.Equal(t, "version", b.Records[0].Headers[0].Key)
				require.Equal(t, []byte("1.0"), b.Records[0].Headers[0].Value)
			},
		},
		{
			name: "add header",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *Engine) {
				err := engine.AddScript(newScript("test.js", `
					import { on } from 'mokapi'
					export default function() {
						on('kafka', function(message) {
							message.headers = { version: '1.0' }
						})
					}
				`))
				require.NoError(t, err)

				sendMessage(s, map[string]string{"foo": "bar"})

				b, _ := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Len(t, b.Records[0].Headers, 2)
				require.Contains(t, b.Records[0].Headers, kafka.RecordHeader{
					Key:   "foo",
					Value: []byte("bar"),
				})
				require.Contains(t, b.Records[0].Headers, kafka.RecordHeader{
					Key:   "version",
					Value: []byte("1.0"),
				})
			},
		},
		{
			name: "remove all headers",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *Engine) {
				err := engine.AddScript(newScript("test.js", `
					import { on } from 'mokapi'
					export default function() {
						on('kafka', function(message) {
							message.headers = null
						})
					}
				`))
				require.NoError(t, err)

				sendMessage(s, map[string]string{"foo": "bar"})

				b, _ := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Len(t, b.Records[0].Headers, 0)
			},
		},
		{
			name: "validation error",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *Engine) {
				err := engine.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'foo', messages: [{ data: 12 }] })
					}
				`))
				require.EqualError(t, err, "produce kafka message to 'foo' failed: marshal data to 'application/json' failed: validation error on int64, expected schema type=string at reflect.methodValueCall (native)")

				b, errCode := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Len(t, b.Records, 0, "no record should be written")
			},
		},
		{
			name: "using value instead of data (skip validation)",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *Engine) {
				err := engine.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'foo', messages: [{ value: 12 }] })
					}
				`))
				require.NoError(t, err)

				b, errCode := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Len(t, b.Records, 1, "message should be written despite validation error")
				require.Equal(t, "12", kafka.BytesToString(b.Records[0].Value))
			},
		},
		{
			name: "test retry",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *Engine) {
				logrus.SetOutput(io.Discard)
				logrus.SetLevel(logrus.DebugLevel)
				hook := test.NewGlobal()

				go func() {
					time.Sleep(time.Second * 1)

					config := asyncapitest.NewConfig(
						asyncapitest.WithInfo("foo", "", ""),
						asyncapitest.WithChannel("retry",
							asyncapitest.WithSubscribeAndPublish(
								asyncapitest.WithMessage(
									asyncapitest.WithContentType("application/json"),
									asyncapitest.WithPayload(schematest.New("string")),
									asyncapitest.WithKey(schematest.New("string"))))),
					)
					app.AddKafka(getConfig(config), nil)
				}()

				err := engine.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'retry', messages: [{ value: 12 }] })
					}
				`))
				require.NoError(t, err)

				b, errCode := s.Topic("retry").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Len(t, b.Records, 1, "message should be written despite validation error")
				require.Equal(t, "12", kafka.BytesToString(b.Records[0].Value))
				require.Equal(t, []string([]string{
					"parsing script test.js",
					"kafka topic 'retry' not found. Retry in 200ms",
					"kafka topic 'retry' not found. Retry in 800ms",
				}), getMessages(hook))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			config := asyncapitest.NewConfig(
				asyncapitest.WithInfo("foo", "", ""),
				asyncapitest.WithChannel("foo",
					asyncapitest.WithSubscribeAndPublish(
						asyncapitest.WithMessage(
							asyncapitest.WithContentType("application/json"),
							asyncapitest.WithPayload(schematest.New("string")),
							asyncapitest.WithKey(schematest.New("string")),
						),
					),
					asyncapitest.WithTopicBinding(bindings.TopicBindings{ValueSchemaValidation: true}),
				),
				asyncapitest.WithChannel("bar",
					asyncapitest.WithChannelKafka(bindings.TopicBindings{Partitions: 10}),
					asyncapitest.WithSubscribeAndPublish(
						asyncapitest.WithMessage(
							asyncapitest.WithContentType("application/json"),
							asyncapitest.WithPayload(schematest.New("string")),
							asyncapitest.WithKey(schematest.New("string"))))),
			)
			app := runtime.New()
			engine := New(reader, app, static.JsConfig{}, false)
			info := app.AddKafka(getConfig(config), engine)
			tc.test(t, app, info.Store, engine)
		})
	}
}

func readBytes(b kafka.Bytes) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(b)
	return buf.Bytes()
}

func getConfig(c *asyncApi.Config) *dynamic.Config {
	u, _ := url.Parse("foo.bar")
	cfg := &dynamic.Config{Data: c}
	cfg.Info.Url = u
	return cfg
}

func sendMessage(s *store.Store, headers map[string]string) {
	var rHeaders []kafka.RecordHeader
	for k, v := range headers {
		rHeaders = append(rHeaders, kafka.RecordHeader{
			Key:   k,
			Value: []byte(v),
		})
	}

	rr := kafkatest.NewRecorder()
	r := kafkatest.NewRequest("kafkatest", 3, &produce.Request{
		Topics: []produce.RequestTopic{
			{Name: "foo", Partitions: []produce.RequestPartition{
				{
					Record: kafka.RecordBatch{
						Records: []kafka.Record{
							{
								Offset:  0,
								Time:    time.Now(),
								Key:     kafka.NewBytes([]byte("foo-1")),
								Value:   kafka.NewBytes([]byte(`"bar-1"`)),
								Headers: rHeaders,
							},
						},
					},
				},
			},
			}}})
	m := monitor.New()
	r.Context = monitor.NewKafkaContext(r.Context, m.Kafka)
	s.ServeMessage(rr, r)
}

func getMessages(hook *test.Hook) []string {
	var result []string
	for _, e := range hook.Entries {
		result = append(result, e.Message)
	}
	return result
}
