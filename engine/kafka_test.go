package engine_test

import (
	"bytes"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"io"
	"mokapi/config/dynamic"
	"mokapi/engine"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/produce"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/schema/json/schematest"
	"net/url"
	"testing"
	"time"
)

func TestKafkaClient_Produce(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App, s *store.Store, engine *engine.Engine)
	}{
		{
			name: "random message",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *engine.Engine) {
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
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *engine.Engine) {
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
			},
		},
		{
			name: "to partition 5",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *engine.Engine) {
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
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *engine.Engine) {
				for i := 0; i < 10; i++ {
					app.AddKafka(getConfig(
						asyncapi3test.NewConfig(asyncapi3test.WithInfo(fmt.Sprintf("x%v", i), "", ""))), enginetest.NewEngine())
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
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *engine.Engine) {
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
				require.Equal(t, `{"offset":0,"key":"foo-1","value":"\"bar-1\"","headers":{"x-specification-message-id":"foo"}}`, hook.LastEntry().Message)

				b, errCode := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Equal(t, "mokapi", string(readBytes(b.Records[0].Value)))
				require.Len(t, b.Records[0].Headers, 2)
				version, found := getHeader("version", b.Records[0].Headers)
				require.True(t, found, "version header not found")
				require.Equal(t, []byte("1.0"), version.Value)
			},
		},
		{
			name: "add header",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *engine.Engine) {
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
				require.Len(t, b.Records[0].Headers, 3)
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
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *engine.Engine) {
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
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *engine.Engine) {
				logrus.SetOutput(io.Discard)
				hook := test.NewGlobal()

				err := engine.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'foo', messages: [{ data: 12 }] })
					}
				`))
				require.EqualError(t, err, "produce kafka message to 'foo' failed: encoding data to 'application/json' failed: found 1 error:\ninvalid type, expected string but got integer\nschema path #/type at mokapi/js/kafka.(*Module).Produce-fm (native)")

				b, errCode := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Len(t, b.Records, 0, "no record should be written")

				// logs
				require.Len(t, hook.Entries, 2)
				require.Equal(t, "js error: produce kafka message to 'foo' failed: encoding data to 'application/json' failed: found 1 error:\ninvalid type, expected string but got integer\nschema path #/type in test.js", hook.LastEntry().Message)
			},
		},
		{
			name: "using value instead of data (skip validation)",
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *engine.Engine) {
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
			test: func(t *testing.T, app *runtime.App, s *store.Store, engine *engine.Engine) {
				logrus.SetOutput(io.Discard)
				logrus.SetLevel(logrus.DebugLevel)
				hook := test.NewGlobal()

				go func() {
					time.Sleep(time.Second * 1)

					config := asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "", ""),
						asyncapi3test.WithChannel("retry",
							asyncapi3test.WithMessage("foo",
								asyncapi3test.WithContentType("application/json"),
								asyncapi3test.WithPayload(schematest.New("string")),
								asyncapi3test.WithKey(schematest.New("string")))),
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
					"executing script test.js",
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

			config := asyncapi3test.NewConfig(
				asyncapi3test.WithInfo("foo", "", ""),
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithContentType("application/json"),
						asyncapi3test.WithPayload(schematest.New("string")),
						asyncapi3test.WithKey(schematest.New("string")),
					),
					asyncapi3test.WithTopicBinding(asyncapi3.TopicBindings{ValueSchemaValidation: true}),
				),
				asyncapi3test.WithChannel("bar",
					asyncapi3test.WithChannelKafka(asyncapi3.TopicBindings{Partitions: 10}),
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithContentType("application/json"),
						asyncapi3test.WithPayload(schematest.New("string")),
						asyncapi3test.WithKey(schematest.New("string")))),
			)
			app := runtime.New()
			e := enginetest.NewEngine(
				engine.WithKafkaClient(engine.NewKafkaClient(app)),
				engine.WithLogger(logrus.StandardLogger()),
			)

			info, err := app.AddKafka(getConfig(config), e)
			require.NoError(t, err)
			tc.test(t, app, info.Store, e)
		})
	}
}

func readBytes(b kafka.Bytes) []byte {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(b)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func getConfig(c *asyncapi3.Config) *dynamic.Config {
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
						Records: []*kafka.Record{
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

func getHeader(name string, headers []kafka.RecordHeader) (kafka.RecordHeader, bool) {
	for _, h := range headers {
		if h.Key == name {
			return h, true
		}
	}
	return kafka.RecordHeader{}, false
}
