package store_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"net"
	"testing"
	"time"
)

func TestPublish(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "publish no consumers",
			test: func(t *testing.T, s *store.Store) {
				publisher := newClient("publisher", s)
				defer publisher.close()

				publisher.connect()

				rr := publisher.send(&mqtt.Message{
					Header: &mqtt.Header{
						QoS:    0,
						Retain: false,
					},
					Payload: &mqtt.PublishRequest{
						MessageId: 11,
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})
				res := rr.Message.Payload.(*mqtt.PublishResponse)
				require.Equal(t, int16(11), res.MessageId)
			},
		},
		{
			name: "publish with one consumer QoS=0",
			test: func(t *testing.T, s *store.Store) {
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
							},
						},
					},
					Context: consumer.ctx,
				})

				// publish
				publisher.send(&mqtt.Message{
					Header: &mqtt.Header{
						QoS:    0,
						Retain: false,
					},
					Payload: &mqtt.PublishRequest{
						MessageId: 11,
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})

				consumer.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				res := &mqtt.Message{}
				err := res.Read(consumer.conn)
				require.NoError(t, err)
				require.NotNil(t, res)
				pub := res.Payload.(*mqtt.PublishRequest)
				require.Equal(t, int16(1), pub.MessageId)
				require.Equal(t, "/foo/bar", pub.Topic)
				require.Equal(t, []byte("hello world"), pub.Data)
			},
		},
		{
			name: "consumer subscribes after published retain message QoS=0",
			test: func(t *testing.T, s *store.Store) {
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
						QoS:    0,
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
							},
						},
					},
					Context: consumer.ctx,
				})

				consumer.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				res := &mqtt.Message{}
				err := res.Read(consumer.conn)
				require.NoError(t, err)
				require.NotNil(t, res)
				pub := res.Payload.(*mqtt.PublishRequest)
				require.Equal(t, int16(1), pub.MessageId)
				require.Equal(t, "/foo/bar", pub.Topic)
				require.Equal(t, []byte("hello world"), pub.Data)
			},
		},
		{
			name: "consumer subscribes but is offline when publishing QoS=0",
			test: func(t *testing.T, s *store.Store) {
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
							},
						},
					},
					Context: consumer.ctx,
				})

				consumer.close()

				// publish
				publisher.send(&mqtt.Message{
					Header: &mqtt.Header{
						QoS: 0,
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
				consumer.connect()

				consumer.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				res := &mqtt.Message{}
				err := res.Read(consumer.conn)
				require.Error(t, err)
			},
		},
		{
			name: "consumer subscribes but is offline when publishing QoS=1",
			test: func(t *testing.T, s *store.Store) {
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
								QoS:  byte(1),
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
				err := res.Read(consumer.conn)
				require.NoError(t, err)
				require.NotNil(t, res)
				pub := res.Payload.(*mqtt.PublishRequest)
				require.Equal(t, int16(1), pub.MessageId)
				require.Equal(t, "/foo/bar", pub.Topic)
				require.Equal(t, []byte("hello world"), pub.Data)

				// broker should send the message again because no ACK was sent
				consumer.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				res = &mqtt.Message{}
				err = res.Read(consumer.conn)
				require.NoError(t, err)
				require.NotNil(t, res)
				pub = res.Payload.(*mqtt.PublishRequest)
				require.Equal(t, int16(1), pub.MessageId)
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
				err = res.Read(consumer.conn)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine())
			defer s.Close()

			tc.test(t, s)
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
