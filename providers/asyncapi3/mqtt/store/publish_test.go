package store_test

import (
	"context"
	"mokapi/engine/enginetest"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"mokapi/schema/json/schema/schematest"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPublish(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, s *store.Store, eh events.Handler, m *monitor.Mqtt)
	}{
		{
			name: "publish QoS=0 topic not specified",
			test: func(t *testing.T, s *store.Store, eh events.Handler, m *monitor.Mqtt) {
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
			},
		},
		{
			name: "publish QoS=1 topic not specified",
			test: func(t *testing.T, s *store.Store, eh events.Handler, m *monitor.Mqtt) {
				publisher := newClient("publisher", s)
				defer publisher.close()

				publisher.connect()

				rr := publisher.send(&mqtt.Message{
					Header: &mqtt.Header{
						QoS:    1,
						Retain: false,
					},
					Payload: &mqtt.PublishRequest{
						MessageId: uint16(123),
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})
				require.NotNil(t, rr.Message)
				res := rr.Message.Payload.(*mqtt.PublishResponse)
				require.Equal(t, uint16(123), res.MessageId)
				require.Equal(t, mqtt.TopicNameInvalid, res.ReasonCode)
			},
		},
		{
			name: "publish QoS=1 topic specified",
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
						QoS:    1,
						Retain: false,
					},
					Payload: &mqtt.PublishRequest{
						MessageId: uint16(123),
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})
				require.NotNil(t, rr.Message)
				res := rr.Message.Payload.(*mqtt.PublishResponse)
				require.Equal(t, uint16(123), res.MessageId)
				require.Equal(t, mqtt.PublishSuccess, res.ReasonCode)

				evts := eh.GetEvents(events.NewTraits().WithNamespace("mqtt").With("type", "message"))
				require.Len(t, evts, 1)
				d := evts[0].Data.(*store.LogMessage)
				require.Equal(t, "/foo/bar", d.Topic)
				require.Equal(t, "publisher", d.ClientId)
				require.Equal(t, "msg-name", d.MessageId)
				require.Equal(t, "hello world", d.Message.Value)
				require.Equal(t, "publisher", d.ClientId)

				require.Equal(t, float64(1), m.Messages.Sum(metrics.NewQuery()))
				require.Equal(t, float64(1), m.Messages.WithLabel("test-server", "/foo/bar").Value())
				require.Greater(t, m.LastMessage.WithLabel("test-server", "/foo/bar").Value(), float64(1))
			},
		},
		{
			name: "publish QoS=1 topic specified message not valid",
			test: func(t *testing.T, s *store.Store, eh events.Handler, m *monitor.Mqtt) {
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("test-server", "", ""),
					asyncapi3test.WithChannel("/foo/bar",
						asyncapi3test.WithMessage("bar", asyncapi3test.WithPayload(schematest.New("integer"))),
					),
				))

				publisher := newClient("publisher", s)
				defer publisher.close()

				publisher.connect()

				rr := publisher.send(&mqtt.Message{
					Header: &mqtt.Header{
						QoS:    1,
						Retain: false,
					},
					Payload: &mqtt.PublishRequest{
						MessageId: uint16(123),
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})
				require.NotNil(t, rr.Message)
				res := rr.Message.Payload.(*mqtt.PublishResponse)
				require.Equal(t, uint16(123), res.MessageId)
				require.Equal(t, mqtt.PayloadFormatInvalid, res.ReasonCode)
			},
		},
		{
			name: "publish with one consumer QoS=1",
			test: func(t *testing.T, s *store.Store, eh events.Handler, m *monitor.Mqtt) {
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("test-server", "", ""),
					asyncapi3test.WithChannel("/foo/bar",
						asyncapi3test.WithMessage("bar", asyncapi3test.WithPayload(schematest.New("string"))),
					),
				))

				publisher := newClient("publisher", s)
				defer publisher.close()
				consumer := newClient("consumer", s)
				defer consumer.close()

				publisher.connect()
				consumer.connect()

				// subscribe
				consumer.send(&mqtt.Message{
					Payload: &mqtt.SubscribeRequest{
						MessageId: 1,
						Topics: []mqtt.SubscribeTopic{
							{
								Name: "/foo/bar",
								QoS:  1,
							},
						},
					},
					Context: consumer.ctx,
				})

				// publish
				publisher.send(&mqtt.Message{
					Header: &mqtt.Header{
						QoS:    1,
						Retain: false,
					},
					Payload: &mqtt.PublishRequest{
						MessageId: 11,
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})

				_ = consumer.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				res := &mqtt.Message{}
				err := res.Read(consumer.conn, consumer.clientCtx)
				require.NoError(t, err)
				require.NotNil(t, res)
				pub := res.Payload.(*mqtt.PublishRequest)
				require.Equal(t, uint16(1), pub.MessageId)
				require.Equal(t, "/foo/bar", pub.Topic)
				require.Equal(t, []byte("hello world"), pub.Data)
			},
		},
		{
			name: "consumer subscribes after published retain message QoS=1",
			test: func(t *testing.T, s *store.Store, eh events.Handler, m *monitor.Mqtt) {
				s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/foo/bar")))

				publisher := newClient("publisher", s)
				defer publisher.close()
				consumer := newClient("consumer", s)
				defer consumer.close()

				publisher.connect()
				consumer.connect()

				// publish
				publisher.send(&mqtt.Message{
					Header: &mqtt.Header{
						QoS:    1,
						Retain: true,
					},
					Payload: &mqtt.PublishRequest{
						MessageId: 11,
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})

				time.Sleep(500 * time.Millisecond)
				// subscribe
				consumer.send(&mqtt.Message{
					Payload: &mqtt.SubscribeRequest{
						MessageId: 1,
						Topics: []mqtt.SubscribeTopic{
							{
								Name: "/foo/bar",
								QoS:  1,
							},
						},
					},
					Context: consumer.ctx,
				})

				consumer.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				res := &mqtt.Message{}
				err := res.Read(consumer.conn, consumer.clientCtx)
				require.NoError(t, err)
				require.NotNil(t, res)
				pub := res.Payload.(*mqtt.PublishRequest)
				require.Equal(t, uint16(1), pub.MessageId)
				require.Equal(t, "/foo/bar", pub.Topic)
				require.Equal(t, []byte("hello world"), pub.Data)
			},
		},
		{
			name: "consumer subscribes but is offline when publishing QoS=1",
			test: func(t *testing.T, s *store.Store, eh events.Handler, m *monitor.Mqtt) {
				s.RetryInterval = 500 * time.Millisecond
				s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/foo/bar")))

				publisher := newClient("publisher", s)
				defer publisher.close()
				consumer := newClient("consumer", s)
				defer consumer.close()

				publisher.connect()
				consumer.connect()

				// subscribe
				consumer.send(&mqtt.Message{
					Payload: &mqtt.SubscribeRequest{
						MessageId: 1,
						Topics: []mqtt.SubscribeTopic{
							{
								Name: "/foo/bar",
								QoS:  1,
							},
						},
					},
					Context: consumer.ctx,
				})

				consumer.close()
				time.Sleep(500 * time.Millisecond)

				// publish
				publisher.send(&mqtt.Message{
					Header: &mqtt.Header{
						QoS:    1,
						Retain: true,
					},
					Payload: &mqtt.PublishRequest{
						MessageId: 11,
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})

				time.Sleep(500 * time.Millisecond)

				consumer = newClient("consumer", s)
				defer consumer.close()
				consumer.connect()

				consumer.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				res := &mqtt.Message{}
				err := res.Read(consumer.conn, consumer.clientCtx)
				require.NoError(t, err)
				require.NotNil(t, res)
				pub := res.Payload.(*mqtt.PublishRequest)
				require.Equal(t, uint16(1), pub.MessageId)
				require.Equal(t, "/foo/bar", pub.Topic)
				require.Equal(t, []byte("hello world"), pub.Data)

				// broker should send the message again because no ACK was sent
				consumer.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				res = &mqtt.Message{}
				err = res.Read(consumer.conn, consumer.clientCtx)
				require.NoError(t, err)
				require.NotNil(t, res)
				pub = res.Payload.(*mqtt.PublishRequest)
				require.Equal(t, uint16(1), pub.MessageId)
				require.Equal(t, "/foo/bar", pub.Topic)
				require.Equal(t, []byte("hello world"), pub.Data)

				consumer.send(&mqtt.Message{
					Header: &mqtt.Header{
						Type: mqtt.PUBACK,
					},
					Payload: &mqtt.PublishResponse{MessageId: pub.MessageId},
				})

				// no further message should be received
				consumer.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				res = &mqtt.Message{}
				err = res.Read(consumer.conn, consumer.clientCtx)
			},
		},
		{
			name: "publish topic with parameter",
			test: func(t *testing.T, s *store.Store, eh events.Handler, m *monitor.Mqtt) {
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("test-server", "", ""),
					asyncapi3test.WithChannel("sensors/{sensorId}/data",
						asyncapi3test.WithParameter(
							"sensorId", &asyncapi3.Parameter{},
						),
					)),
				)

				publisher := newClient("publisher", s)
				defer publisher.close()

				publisher.connect()

				rr := publisher.send(&mqtt.Message{
					Header: &mqtt.Header{
						QoS:    1,
						Retain: false,
					},
					Payload: &mqtt.PublishRequest{
						MessageId: uint16(123),
						Topic:     "sensors/1234z/data",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})
				require.NotNil(t, rr.Message)
				res := rr.Message.Payload.(*mqtt.PublishResponse)
				require.Equal(t, uint16(123), res.MessageId)
				require.Equal(t, mqtt.PublishSuccess, res.ReasonCode)

				evts := eh.GetEvents(events.
					NewTraits().
					WithNamespace("mqtt").
					WithName("test-server").
					With("topic", "sensors/{sensorId}/data").
					With("type", "message"),
				)
				require.Len(t, evts, 1)
				d := evts[0].Data.(*store.LogMessage)
				require.Equal(t, "sensors/1234z/data", d.Topic)
				require.Equal(t, "publisher", d.ClientId)
				require.Equal(t, "hello world", d.Message.Value)
				require.Equal(t, "namespace=mqtt, name=test-server, clientId=publisher, sensorId=1234z, topic=sensors/{sensorId}/data, type=message", evts[0].Traits.String())

				require.Equal(t, float64(1), m.Messages.Sum(metrics.NewQuery()))
				require.Equal(t, float64(1), m.Messages.WithLabel("test-server", "sensors/{sensorId}/data").Value())
				require.Greater(t, m.LastMessage.WithLabel("test-server", "sensors/{sensorId}/data").Value(), float64(1))
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
				enginetest.NewEngine(),
				eh,
				m,
			)
			defer s.Close()

			tc.test(t, s, eh, m)
		})
	}
}

type client struct {
	ctx       context.Context
	clientCtx *mqtt.ClientContext
	conn      net.Conn
	handler   mqtt.Handler
}

func newClient(clientId string, handler mqtt.Handler) *client {
	ctx, conn := mqtttest.NewTestClientContext()
	c := &client{
		ctx:       ctx,
		clientCtx: mqtt.ClientFromContext(ctx),
		conn:      conn,
		handler:   handler,
	}
	c.clientCtx.ClientId = clientId
	return c
}

func (c *client) close() {
	c.conn.Close()
}

func (c *client) connect() *mqtttest.MessageRecorder {
	rr := mqtttest.NewRecorder()
	c.handler.ServeMessage(rr, &mqtt.Message{
		Payload: &mqtt.ConnectRequest{
			ClientId: c.clientCtx.ClientId,
		},
		Context: c.ctx,
	})
	return rr
}

func (c *client) send(r *mqtt.Message) *mqtttest.MessageRecorder {
	rr := mqtttest.NewRecorder()
	c.handler.ServeMessage(rr, r)
	return rr
}
