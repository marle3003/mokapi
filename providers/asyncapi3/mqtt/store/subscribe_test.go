package store_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"testing"
)

func TestSubscribe(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "subscribe",
			test: func(t *testing.T, s *store.Store) {
				rr := mqtttest.NewRecorder()
				ctx := mqtt.NewClientContext(context.Background(), "localhost")
				client := mqtt.ClientFromContext(ctx)
				client.ClientId = "foo"

				s.ServeMessage(rr, &mqtt.Request{
					Message: &mqtt.ConnectRequest{
						ClientId: client.ClientId,
					},
				})

				s.ServeMessage(rr, &mqtt.Request{
					Message: &mqtt.SubscribeRequest{
						MessageId: 11,
						Topics: []mqtt.SubscribeTopic{
							{Name: "foo"},
						},
					},
					Context: ctx,
				})
				res := rr.Message.(*mqtt.SubscribeResponse)
				require.Equal(t, int16(11), res.MessageId)
				require.Len(t, res.TopicQoS, 1)
				require.Equal(t, byte(0), res.TopicQoS[0])
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
