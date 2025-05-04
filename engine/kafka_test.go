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
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/schema/json/schema/schematest"
	"net/url"
	"testing"
	"time"
)

func TestKafkaClient(t *testing.T) {
	createCfg := func(topic string, msg *asyncapi3.Message) *asyncapi3.Config {
		ch := asyncapi3test.NewChannel(
			asyncapi3test.UseMessage("foo",
				&asyncapi3.MessageRef{Value: msg},
			),
		)

		return asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("foo", "", ""),
			asyncapi3test.AddChannel("foo", ch),
			asyncapi3test.WithComponentMessage("foo", msg),
			asyncapi3test.WithOperation("sendAction",
				asyncapi3test.WithOperationAction("send"),
				asyncapi3test.WithOperationChannel(ch),
				asyncapi3test.UseOperationMessage(msg),
			),
		)
	}

	testcases := []struct {
		name string
		cfg  func() *asyncapi3.Config
		test func(t *testing.T, e *engine.Engine, app *runtime.App)
	}{
		{
			name: "produce empty parameters",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ })
					}
				`))

				require.NoError(t, err)
				b, errCode := app.Kafka.Get("foo").Store.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Equal(t, "XidZuoWq ", kafka.BytesToString(b.Records[0].Key))
				require.Equal(t, "\"\"", kafka.BytesToString(b.Records[0].Value))
			},
		},
		{
			name: "produce with topic and cluster set",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'foo', cluster: 'foo' })
					}
				`))

				require.NoError(t, err)
				b, errCode := app.Kafka.Get("foo").Store.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Equal(t, "XidZuoWq ", kafka.BytesToString(b.Records[0].Key))
				require.Equal(t, "\"\"", kafka.BytesToString(b.Records[0].Value))
			},
		},
		{
			name: "produce but cluster not found",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({
							cluster: 'foo2',
							retry: { retries: 0 }
						})
					}
				`))

				require.EqualError(t, err, "kafka cluster 'foo2' not found at mokapi/js/kafka.(*Module).Produce-fm (native)")
			},
		},
		{
			name: "produce but topic not found",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ 
							topic: 'foo2',
							cluster: 'foo',
							retry: { retries: 0 }
						})
					}
				`))

				require.EqualError(t, err, "kafka topic 'foo2' not found at mokapi/js/kafka.(*Module).Produce-fm (native)")
			},
		},
		{
			name: "produce with specific message",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ 
							messages: [
								{
									key: 'foo',
									data: 'bar',
								}
							]
						})
					}
				`))

				require.NoError(t, err)
				b, errCode := app.Kafka.Get("foo").Store.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Equal(t, "foo", kafka.BytesToString(b.Records[0].Key))
				require.Equal(t, `"bar"`, kafka.BytesToString(b.Records[0].Value))
			},
		},
		{
			name: "produce with specific message value not validating against schema",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ 
							messages: [
								{
									key: 'foo',
									// value is not validate by Mokapi
									value: int32ToBytes(123),
								}
							]
						})
					}
					function int32ToBytes (int) {
					  return [
						int & 0xff,
						(int >> 8) & 0xff,
						(int >> 16) & 0xff,
						(int >> 24) & 0xff
					  ]
					}
				`))

				require.NoError(t, err)
				b, errCode := app.Kafka.Get("foo").Store.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Equal(t, "foo", kafka.BytesToString(b.Records[0].Key))
				val := make([]byte, 4)
				_, _ = b.Records[0].Value.Seek(0, 0)
				_, err = b.Records[0].Value.Read(val)
				require.NoError(t, err)
				require.Equal(t, []byte{123, 0, 0, 0}, val)
			},
		},
		{
			name: "produce with partition",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				cfg := createCfg("foo", msg)
				cfg.Channels["foo"].Value.Bindings.Kafka.Partitions = 10

				return cfg
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({
							messages: [
								{
									key: 'foo',
									data: 'bar',
									partition: 5
								}
							]
						})
					}
				`))

				require.NoError(t, err)
				b, errCode := app.Kafka.Get("foo").Store.Topic("foo").Partition(5).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Equal(t, "foo", kafka.BytesToString(b.Records[0].Key))
				require.Equal(t, `"bar"`, kafka.BytesToString(b.Records[0].Value))
			},
		},
		{
			name: "produce with header",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({
							messages: [
								{
									headers: {
										foo: 'bar'
									}
								}
							]
						})
					}
				`))

				require.NoError(t, err)
				b, errCode := app.Kafka.Get("foo").Store.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Len(t, b.Records[0].Headers, 1)
				require.Equal(t, "foo", b.Records[0].Headers[0].Key)
				require.Equal(t, "bar", string(b.Records[0].Headers[0].Value))
			},
		},
		{
			name: "multiple messages",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						const result = produce({
							topic: 'foo',
							messages: [
								{ key: 'key1', data: 'foo'},
								{ key: 'key2', data: 'bar'}
							],
						})
						console.log(result)
					}
				`))
				require.NoError(t, err)
				b, errCode := app.Kafka.Get("foo").Store.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Equal(t, "key1", kafka.BytesToString(b.Records[0].Key))
				require.Equal(t, `"foo"`, kafka.BytesToString(b.Records[0].Value))
				require.Equal(t, "key2", kafka.BytesToString(b.Records[1].Key))
				require.Equal(t, `"bar"`, kafka.BytesToString(b.Records[1].Value))

				require.Equal(t, float64(2), app.Monitor.Kafka.Messages.Sum())
			},
		},
		{
			name: "multiple clusters only topic is set",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				for i := 0; i < 3; i++ {
					_, _ = app.Kafka.Add(getConfig(
						asyncapi3test.NewConfig(asyncapi3test.WithInfo(fmt.Sprintf("x%v", i), "", ""))), e)
				}

				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'foo' })
					}
				`))
				require.NoError(t, err)

				b, errCode := app.Kafka.Get("foo").Store.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Len(t, b.Records, 1)
			},
		},
		{
			name: "two cluster same topic",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				cfg := createCfg("foo", nil)
				cfg.Info.Name = "Other Cluster"
				_, _ = app.Kafka.Add(getConfig(cfg), e)

				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'foo' })
					}
				`))
				require.EqualError(t, err, "ambiguous topic foo. Specify the cluster at mokapi/js/kafka.(*Module).Produce-fm (native)")
			},
		},
		{
			name: "trigger event",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				hook := test.NewGlobal()

				err := e.AddScript(newScript("test.js", `
					import { on } from 'mokapi'
					import { produceAsync } from 'mokapi/kafka'
					export default async function() {
						on('kafka', function(message) {
							console.log(message)
							message.value = '"mokapi"'
							message.headers = { version: '1.0' }
							return true
						})
						await produceAsync({ topic: 'foo', messages: [ { data: 'bar' } ] })
					}
				`))
				require.NoError(t, err)

				require.Equal(t, `{"offset":0,"key":"XidZuoWq ","value":"\"bar\"","schemaId":0,"headers":{}}`, hook.LastEntry().Message)

				b, errCode := app.Kafka.Get("foo").Store.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Equal(t, "\"mokapi\"", string(readBytes(b.Records[0].Value)))
				require.Len(t, b.Records[0].Headers, 1)
				version, found := getHeader("version", b.Records[0].Headers)
				require.True(t, found, "version header not found")
				require.Equal(t, []byte("1.0"), version.Value)
			},
		},
		{
			name: "trigger event add header",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				err := e.AddScript(newScript("test.js", `
					import { on } from 'mokapi'
					import { produceAsync } from 'mokapi/kafka'
					export default async function() {
						on('kafka', function(message) {
							message.headers = { version: '1.0' }
							return true
						})
						await produceAsync({ topic: 'foo', messages: [ { headers: { foo: 'bar' } } ] })
					}
				`))
				require.NoError(t, err)

				b, errCode := app.Kafka.Get("foo").Store.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
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
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				err := e.AddScript(newScript("test.js", `
					import { on } from 'mokapi'
					import { produceAsync } from 'mokapi/kafka'
					export default async function() {
						on('kafka', function(message) {
							message.headers = null
							return true
						})
						await produceAsync({ topic: 'foo', messages: [ { headers: { foo: 'bar' } } ] })
					}
				`))
				require.NoError(t, err)

				b, _ := app.Kafka.Get("foo").Store.Topic("foo").Partition(0).Read(0, 1000)
				require.Len(t, b.Records[0].Headers, 0)
			},
		},
		{
			name: "validation error",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")),
					asyncapi3test.WithKey(schematest.New("string")),
				)
				return createCfg("foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				logrus.SetOutput(io.Discard)
				hook := test.NewGlobal()

				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'foo', messages: [{ data: 12 }] })
					}
				`))
				require.EqualError(t, err, "producing kafka message to 'foo' failed: no matching 'send' or 'receive' operation found for value: 12 at mokapi/js/kafka.(*Module).Produce-fm (native)")

				b, errCode := app.Kafka.Get("foo").Store.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Len(t, b.Records, 0, "no record should be written")

				// logs
				require.Len(t, hook.Entries, 2)
				require.Equal(t, "js error: producing kafka message to 'foo' failed: no matching 'send' or 'receive' operation found for value: 12 in test.js", hook.LastEntry().Message)
			},
		},
		{
			name: "test retry",
			cfg: func() *asyncapi3.Config {
				return nil
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App) {
				logrus.SetOutput(io.Discard)
				logrus.SetLevel(logrus.DebugLevel)
				hook := test.NewGlobal()

				go func() {
					time.Sleep(time.Second * 1)

					msg := asyncapi3test.NewMessage(
						asyncapi3test.WithPayload(schematest.New("string")),
						asyncapi3test.WithKey(schematest.New("string")),
					)

					ch := asyncapi3test.NewChannel(
						asyncapi3test.UseMessage("foo",
							&asyncapi3.MessageRef{Value: msg},
						),
					)

					cfg := asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("retry", "", ""),
						asyncapi3test.AddChannel("retry", ch),
						asyncapi3test.WithComponentMessage("foo", msg),
						asyncapi3test.WithOperation("sendAction",
							asyncapi3test.WithOperationAction("send"),
							asyncapi3test.WithOperationChannel(ch),
							asyncapi3test.UseOperationMessage(msg),
						),
					)
					_, _ = app.Kafka.Add(getConfig(cfg), e)
				}()

				err := e.AddScript(newScript("test.js", `
					import { produce } from 'mokapi/kafka'
					export default function() {
						produce({ topic: 'retry', messages: [{ data: 'foo' }], retry: { retries: 4 } })
					}
				`))
				require.NoError(t, err)

				b, errCode := app.Kafka.Get("retry").Store.Topic("retry").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, errCode)
				require.NotNil(t, b)
				require.Len(t, b.Records, 1, "message should be written despite validation error")
				require.Equal(t, `"foo"`, kafka.BytesToString(b.Records[0].Value))
				msg := getMessages(hook)
				require.Contains(t, msg, "kafka topic 'retry' not found. Retry in 500ms")
				require.Contains(t, msg, "kafka topic 'retry' not found. Retry in 1s")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			app := runtime.New(&static.Config{})
			e := enginetest.NewEngine(
				engine.WithKafkaClient(engine.NewKafkaClient(app)),
				engine.WithDefaultLogger(),
			)

			cfg := tc.cfg()
			if cfg != nil {
				_, err := app.Kafka.Add(getConfig(cfg), e)
				require.NoError(t, err)
			}

			tc.test(t, e, app)
		})
	}
}

func readBytes(b kafka.Bytes) []byte {
	_, _ = b.Seek(0, io.SeekStart)
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
