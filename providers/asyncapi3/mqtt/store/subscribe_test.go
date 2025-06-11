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

func TestSubscribe(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "subscribe",
			test: func(t *testing.T, s *store.Store) {
				rr := mqtttest.NewRecorder()
				ctx, conn := mqtttest.NewTestClientContext()
				defer conn.Close()
				client := mqtt.ClientFromContext(ctx)
				client.ClientId = "foo"

				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.ConnectRequest{
						ClientId: client.ClientId,
					},
				})

				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.SubscribeRequest{
						MessageId: 11,
						Topics: []mqtt.SubscribeTopic{
							{Name: "foo"},
						},
					},
					Context: ctx,
				})
				res := rr.Message.Payload.(*mqtt.SubscribeResponse)
				require.Equal(t, int16(11), res.MessageId)
				require.Len(t, res.TopicQoS, 1)
				require.Equal(t, byte(0), res.TopicQoS[0])
			},
		},
		{
			name: "subscribe but no connect previously",
			test: func(t *testing.T, s *store.Store) {
				rr := mqtttest.NewRecorder()
				ctx, conn := mqtttest.NewTestClientContext()
				defer conn.Close()
				client := mqtt.ClientFromContext(ctx)
				client.ClientId = "foo"

				defer func() {
					r := recover()
					require.NotNil(t, r, "Test passed, panic was caught")
				}()

				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.SubscribeRequest{
						MessageId: 11,
						Topics: []mqtt.SubscribeTopic{
							{Name: "foo"},
						},
					},
					Context: ctx,
				})
				t.Error("Test failed, panic was expected")
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
