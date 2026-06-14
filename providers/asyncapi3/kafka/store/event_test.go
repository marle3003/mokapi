package store_test

import (
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/produce"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/schema/schematest"
	"mokapi/try"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProduceTriggersEvent(t *testing.T) {
	sm := &events.StoreManager{}

	triggerCount := 0
	s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngineWithHandler(func(event string, args ...interface{}) []*common.Action {
		triggerCount++
		return nil
	}), sm, monitor.NewKafka())
	defer s.Close()

	s.Update(asyncapi3test.NewConfig(
		asyncapi3test.WithServer("foo", "kafka", "127.0.0.1"),
		asyncapi3test.WithChannel("foo")))
	g := s.GetOrCreateGroup("foo", &store.Broker{})
	g.Commit("foo", 0, 0)
	sm.SetStore(5, events.NewTraits().WithNamespace("kafka"))

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
								Value:   kafka.NewBytes([]byte("bar-1")),
								Headers: nil,
							},
							{
								Offset:  1,
								Time:    time.Now(),
								Key:     kafka.NewBytes([]byte("foo-2")),
								Value:   kafka.NewBytes([]byte("bar-2")),
								Headers: nil,
							},
						},
					},
				},
			},
			}}})

	s.ServeMessage(rr, r)

	require.Equal(t, 2, triggerCount)
}

func TestProduceEvents(t *testing.T) {
	testcases := []struct {
		name    string
		script  string
		records []*kafka.Record
		cfg     []asyncapi3test.ConfigOptions
		test    func(t *testing.T, app *runtime.App)
	}{
		{
			name: "produce event",
			cfg:  []asyncapi3test.ConfigOptions{asyncapi3test.WithChannel("foo")},
			script: `import { on } from 'mokapi'
export default function() {
  on('kafka', function(msg) {
    console.log(msg.key, msg.value)
  }, { track: true })
}
`,
			records: []*kafka.Record{
				{
					Offset:  0,
					Time:    time.Now(),
					Key:     kafka.NewBytes([]byte("foo-1")),
					Value:   kafka.NewBytes([]byte("bar-1")),
					Headers: nil,
				},
			},
			test: func(t *testing.T, app *runtime.App) {
				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("kafka"))
				require.Len(t, evts, 1)
				evt := evts[0]
				d := evt.Data.(*store.KafkaMessageLog)
				require.Equal(t, "foo-1 bar-1", d.Actions[0].Logs[0].Message)
			},
		},
		{
			name: "api is available",
			cfg: []asyncapi3test.ConfigOptions{
				asyncapi3test.WithInfo("EventTest", "", ""),
				asyncapi3test.WithChannel("foo"),
			},
			script: `import { on } from 'mokapi'
export default function() {
  on('kafka', function(msg) {
    console.log(msg.api)
  }, { track: true })
}
`,
			records: []*kafka.Record{
				{
					Offset:  0,
					Time:    time.Now(),
					Key:     kafka.NewBytes([]byte("foo-1")),
					Value:   kafka.NewBytes([]byte("bar-1")),
					Headers: nil,
				},
			},
			test: func(t *testing.T, app *runtime.App) {
				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("kafka"))
				require.Len(t, evts, 1)
				evt := evts[0]
				d := evt.Data.(*store.KafkaMessageLog)
				require.Equal(t, "EventTest", d.Actions[0].Logs[0].Message)
			},
		},
		{
			name: "topic is available",
			cfg: []asyncapi3test.ConfigOptions{
				asyncapi3test.WithChannel("foo"),
			},
			script: `import { on } from 'mokapi'
export default function() {
  on('kafka', function(msg) {
    console.log(msg.topic)
  }, { track: true })
}
`,
			records: []*kafka.Record{
				{
					Offset:  0,
					Time:    time.Now(),
					Key:     kafka.NewBytes([]byte("foo-1")),
					Value:   kafka.NewBytes([]byte("bar-1")),
					Headers: nil,
				},
			},
			test: func(t *testing.T, app *runtime.App) {
				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("kafka"))
				require.Len(t, evts, 1)
				evt := evts[0]
				d := evt.Data.(*store.KafkaMessageLog)
				require.Equal(t, "foo", d.Actions[0].Logs[0].Message)
			},
		},
		{
			name: "partition is available",
			cfg: []asyncapi3test.ConfigOptions{
				asyncapi3test.WithChannel("foo"),
			},
			script: `import { on } from 'mokapi'
export default function() {
  on('kafka', function(msg) {
    console.log(msg.partition)
  }, { track: true })
}
`,
			records: []*kafka.Record{
				{
					Offset:  0,
					Time:    time.Now(),
					Key:     kafka.NewBytes([]byte("foo-1")),
					Value:   kafka.NewBytes([]byte("bar-1")),
					Headers: nil,
				},
			},
			test: func(t *testing.T, app *runtime.App) {
				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("kafka"))
				require.Len(t, evts, 1)
				evt := evts[0]
				d := evt.Data.(*store.KafkaMessageLog)
				require.Equal(t, "0", d.Actions[0].Logs[0].Message)
			},
		},
		{
			name: "offset is available",
			cfg: []asyncapi3test.ConfigOptions{
				asyncapi3test.WithChannel("foo"),
			},
			script: `import { on } from 'mokapi'
export default function() {
  on('kafka', function(msg) {
    console.log(msg.offset)
  }, { track: true })
}
`,
			records: []*kafka.Record{
				{
					Offset:  12,
					Time:    time.Now(),
					Key:     kafka.NewBytes([]byte("foo-1")),
					Value:   kafka.NewBytes([]byte("bar-1")),
					Headers: nil,
				},
			},
			test: func(t *testing.T, app *runtime.App) {
				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("kafka"))
				require.Len(t, evts, 1)
				evt := evts[0]
				d := evt.Data.(*store.KafkaMessageLog)
				require.Equal(t, "12", d.Actions[0].Logs[0].Message)
			},
		},
		{
			name: "schemaId available in event",
			cfg: []asyncapi3test.ConfigOptions{
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithPayload(schematest.New("string")),
						asyncapi3test.WithContentType("application/json"),
						asyncapi3test.WithKafkaMessageBinding(asyncapi3.KafkaMessageBinding{
							SchemaIdLocation:        "payload",
							SchemaIdPayloadEncoding: "4",
						}),
					),
				),
			},
			script: `import { on } from 'mokapi'
export default function() {
  on('kafka', function(msg) {
    console.log(msg.schemaId)
  }, { track: true })
}
`,
			records: []*kafka.Record{
				{
					Offset:  0,
					Time:    time.Now(),
					Key:     kafka.NewBytes([]byte("foo-1")),
					Value:   kafka.NewBytes([]byte{0, 0, 0, 0, 8, '"', 'f', 'o', 'o', '"'}),
					Headers: nil,
				},
			},
			test: func(t *testing.T, app *runtime.App) {
				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("kafka"))
				require.Len(t, evts, 1)
				evt := evts[0]
				d := evt.Data.(*store.KafkaMessageLog)
				require.Equal(t, "8", d.Actions[0].Logs[0].Message)
			},
		},
		{
			name: "add header",
			cfg: []asyncapi3test.ConfigOptions{
				asyncapi3test.WithChannel("foo"),
			},
			script: `import { on } from 'mokapi'
export default function() {
  on('kafka', function(msg) {
    msg.headers['foo'] = 'bar'
  })
}
`,
			records: []*kafka.Record{
				{
					Offset:  0,
					Time:    time.Now(),
					Key:     kafka.NewBytes([]byte("foo-1")),
					Value:   kafka.NewBytes([]byte("bar-1")),
					Headers: nil,
				},
			},
			test: func(t *testing.T, app *runtime.App) {
				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("kafka"))
				require.Len(t, evts, 1)
				evt := evts[0]
				d := evt.Data.(*store.KafkaMessageLog)
				require.Equal(t, "bar", string(d.Headers["foo"].Binary))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app := runtimetest.NewApp()
			e := enginetest.NewEngine()
			app.Engine = e

			s := store.New(
				asyncapi3test.NewConfig(tc.cfg...),
				e, app.Events, app.Monitor.Kafka)
			defer s.Close()

			err := e.AddScript(dynamic.ConfigEvent{
				Config: &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.js")},
					Raw:  []byte(tc.script),
					Data: tc.script,
				},
				Event: dynamic.Create,
			})
			require.NoError(t, err)

			rr := kafkatest.NewRecorder()
			r := kafkatest.NewRequest("kafkatest", 3, &produce.Request{
				Topics: []produce.RequestTopic{
					{
						Name: "foo",
						Partitions: []produce.RequestPartition{
							{
								Record: kafka.RecordBatch{
									Records: tc.records,
								},
							},
						},
					},
				},
			})

			s.ServeMessage(rr, r)

			tc.test(t, app)
		})
	}
}
