package store_test

import (
	"mokapi/config/dynamic"
	"mokapi/engine"
	"mokapi/engine/enginetest"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"mokapi/try"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvent(t *testing.T) {
	testcases := []struct {
		name   string
		engine *engine.Engine
		test   func(t *testing.T, s *store.Store, eh events.Handler, m *monitor.Mqtt)
	}{
		{
			name: "publish event",
			engine: func() *engine.Engine {
				script := `import { on } from 'mokapi'
export default function() {
  on('mqtt', function(msg) {
    console.log(msg.api, msg.topic, msg.value)
  }, { track: true })
}
`

				e := enginetest.NewEngine()
				err := e.AddScript(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.js")},
						Raw:  []byte(script),
						Data: script,
					},
					Event: dynamic.Create,
				})
				require.NoError(t, err)
				return e
			}(),
			test: func(t *testing.T, s *store.Store, eh events.Handler, m *monitor.Mqtt) {
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("test-server", "", ""),
					asyncapi3test.WithChannel("/foo/bar",
						asyncapi3test.WithMessage("msg-name"),
					),
				))

				publisher := newClient("publisher", s)
				defer publisher.close()

				publisher.connect()

				rr := publisher.send(&mqtt.Message{
					Header: &mqtt.Header{
						QoS:    0,
						Retain: false,
					},
					Payload: &mqtt.PublishRequest{
						Topic: "/foo/bar",
						Data:  []byte("hello world"),
					},
					Context: publisher.ctx,
				})
				require.Nil(t, rr.Message)
				evts := eh.GetEvents(events.NewTraits().WithNamespace("mqtt"))
				require.Len(t, evts, 2)
				d := evts[1].Data.(*store.LogMessage)
				require.Equal(t, "test-server /foo/bar hello world", d.Actions[0].Logs[0].Message)
			},
		},
		{
			name: "update retain",
			engine: func() *engine.Engine {
				script := `import { on } from 'mokapi'
export default function() {
  on('mqtt', function(msg) {
    msg.retain = true
  }, { track: true })
}
`

				e := enginetest.NewEngine()
				err := e.AddScript(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.js")},
						Raw:  []byte(script),
						Data: script,
					},
					Event: dynamic.Create,
				})
				require.NoError(t, err)
				return e
			}(),
			test: func(t *testing.T, s *store.Store, eh events.Handler, m *monitor.Mqtt) {
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("test-server", "", ""),
					asyncapi3test.WithChannel("/foo/bar",
						asyncapi3test.WithMessage("msg-name"),
					),
				))

				publisher := newClient("publisher", s)
				defer publisher.close()

				publisher.connect()

				publisher.send(&mqtt.Message{
					Header: &mqtt.Header{
						QoS:    0,
						Retain: false,
					},
					Payload: &mqtt.PublishRequest{
						Topic: "/foo/bar",
						Data:  []byte("hello world"),
					},
					Context: publisher.ctx,
				})
				topic, _ := s.Topic("/foo/bar")
				require.NotNil(t, topic.Retained)

				evts := eh.GetEvents(events.NewTraits().WithNamespace("mqtt"))
				require.Len(t, evts, 2)
				d := evts[1].Data.(*store.LogMessage)
				require.True(t, d.Retain)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			eh := &eventstest.Handler{}
			m := monitor.NewMqtt()
			s := store.New(
				asyncapi3test.NewConfig(asyncapi3test.WithInfo("test-server", "", "")),
				tc.engine,
				eh,
				m,
			)
			defer s.Close()

			tc.test(t, s, eh, m)
		})
	}
}
