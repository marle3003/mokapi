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

				rr := publisher.send(&mqtt.Request{
					Header: &mqtt.Header{
						QoS:    0,
						Retain: false,
					},
					Message: &mqtt.PublishRequest{
						MessageId: 11,
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})
				res := rr.Message.(*mqtt.PublishResponse)
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
				consumer.send(&mqtt.Request{
					Message: &mqtt.SubscribeRequest{
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
				publisher.send(&mqtt.Request{
					Header: &mqtt.Header{
						QoS:    0,
						Retain: false,
					},
					Message: &mqtt.PublishRequest{
						MessageId: 11,
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})

				consumer.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				res, err := mqtt.ReadResponse(consumer.conn)
				require.NoError(t, err)
				require.NotNil(t, res)
				pub := res.Message.(*mqtt.PublishRequest)
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
				publisher.send(&mqtt.Request{
					Header: &mqtt.Header{
						QoS:    0,
						Retain: true,
					},
					Message: &mqtt.PublishRequest{
						MessageId: 11,
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: publisher.ctx,
				})

				time.Sleep(500 * time.Millisecond)
				// subscribe
				consumer.send(&mqtt.Request{
					Message: &mqtt.SubscribeRequest{
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
				res, err := mqtt.ReadResponse(consumer.conn)
				require.NoError(t, err)
				require.NotNil(t, res)
				pub := res.Message.(*mqtt.PublishRequest)
				require.Equal(t, int16(1), pub.MessageId)
				require.Equal(t, "/foo/bar", pub.Topic)
				require.Equal(t, []byte("hello world"), pub.Data)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine())
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

func (c *client) connect() *mqtttest.ResponseRecorder {
	rr := mqtttest.NewRecorder()
	c.handler.ServeMessage(rr, &mqtt.Request{
		Message: &mqtt.ConnectRequest{
			ClientId: c.clientCtx.ClientId,
		},
		Context: c.ctx,
	})
	return rr
}

func (c *client) send(r *mqtt.Request) *mqtttest.ResponseRecorder {
	rr := mqtttest.NewRecorder()
	c.handler.ServeMessage(rr, r)
	return rr
}
