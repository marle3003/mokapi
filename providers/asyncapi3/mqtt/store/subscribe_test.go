package store_test

import (
	"mokapi/engine/enginetest"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"testing"

	"github.com/stretchr/testify/require"
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
				require.Equal(t, uint16(11), res.MessageId)
				require.Len(t, res.ReasonCodes, 1)
				require.Equal(t, mqtt.GrantedQoS0, res.ReasonCodes[0])
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
		{
			name: "unsubscribe",
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

				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.UnsubscribeRequest{
						MessageId: 11,
						Topics:    []string{"foo"},
					},
					Context: ctx,
				})
				require.Equal(t, mqtt.UNSUBACK, rr.Message.Header.Type)
				res := rr.Message.Payload.(*mqtt.UnsubscribeResponse)
				require.Equal(t, uint16(11), res.MessageId)
				require.Len(t, res.ReasonCodes, 1)
				require.Equal(t, mqtt.UnsubscribeSuccess, res.ReasonCodes[0])
			},
		},
		{
			name: "unsubscribe but not existing",
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
					Payload: &mqtt.UnsubscribeRequest{
						MessageId: 11,
						Topics:    []string{"foo"},
					},
					Context: ctx,
				})
				res := rr.Message.Payload.(*mqtt.UnsubscribeResponse)
				require.Equal(t, uint16(11), res.MessageId)
				require.Len(t, res.ReasonCodes, 1)
				require.Equal(t, mqtt.NoSubscriptionExisted, res.ReasonCodes[0])
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
