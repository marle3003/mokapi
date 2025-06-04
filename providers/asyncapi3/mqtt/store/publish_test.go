package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"testing"
)

func TestPublish(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "publish no consumers",
			test: func(t *testing.T, s *store.Store) {
				rr := mqtttest.NewRecorder()
				ctx, conn := mqtttest.NewTestClientContext()
				defer conn.Close()
				client := mqtt.ClientFromContext(ctx)
				client.ClientId = "foo"

				s.ServeMessage(rr, &mqtt.Request{
					Message: &mqtt.ConnectRequest{
						ClientId: client.ClientId,
					},
				})

				s.ServeMessage(rr, &mqtt.Request{
					Header: &mqtt.Header{
						QoS:    0,
						Retain: false,
					},
					Message: &mqtt.PublishRequest{
						MessageId: 11,
						Topic:     "/foo/bar",
						Data:      []byte("hello world"),
					},
					Context: ctx,
				})
				res := rr.Message.(*mqtt.PublishResponse)
				require.Equal(t, int16(11), res.MessageId)
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
