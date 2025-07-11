package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"strings"
	"testing"
)

func TestConnect(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "missing clientId",
			test: func(t *testing.T, s *store.Store) {
				rr := mqtttest.NewRecorder()
				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.ConnectRequest{},
				})
				res := rr.Message.Payload.(*mqtt.ConnectResponse)
				require.Equal(t, mqtt.ErrIdentifierRejected, res.ReturnCode)
			},
		},
		{
			name: "clientId too long",
			test: func(t *testing.T, s *store.Store) {
				rr := mqtttest.NewRecorder()
				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.ConnectRequest{
						ClientId: strings.Repeat("a", 24),
					},
				})
				res := rr.Message.Payload.(*mqtt.ConnectResponse)
				require.Equal(t, mqtt.ErrIdentifierRejected, res.ReturnCode)
			},
		},
		{
			name: "no session",
			test: func(t *testing.T, s *store.Store) {
				rr := mqtttest.NewRecorder()
				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.ConnectRequest{
						ClientId: "foo",
					},
				})
				res := rr.Message.Payload.(*mqtt.ConnectResponse)
				require.False(t, res.SessionPresent, "SessionPresent should be false")
			},
		},
		{
			name: "with session",
			test: func(t *testing.T, s *store.Store) {
				rr := mqtttest.NewRecorder()
				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.ConnectRequest{
						ClientId: "foo",
					},
				})
				res := rr.Message.Payload.(*mqtt.ConnectResponse)
				require.False(t, res.SessionPresent, "SessionPresent should be false")

				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.ConnectRequest{
						ClientId: "foo",
					},
				})
				res = rr.Message.Payload.(*mqtt.ConnectResponse)
				require.True(t, res.SessionPresent, "SessionPresent should be true")
			},
		},
		{
			name: "clean session",
			test: func(t *testing.T, s *store.Store) {
				rr := mqtttest.NewRecorder()
				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.ConnectRequest{
						ClientId: "foo",
					},
				})
				res := rr.Message.Payload.(*mqtt.ConnectResponse)
				require.False(t, res.SessionPresent, "SessionPresent should be false")

				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.ConnectRequest{
						CleanSession: true,
						ClientId:     "foo",
					},
				})
				res = rr.Message.Payload.(*mqtt.ConnectResponse)
				require.False(t, res.SessionPresent, "SessionPresent should be false")
			},
		},
		{
			name: "unknown topic",
			test: func(t *testing.T, s *store.Store) {
				rr := mqtttest.NewRecorder()
				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.ConnectRequest{
						ClientId: "foo",
						WillFlag: true,
						Topic:    "foo",
						Message:  []byte("bar"),
					},
				})
				res := rr.Message.Payload.(*mqtt.ConnectResponse)
				require.Equal(t, mqtt.ErrUnspecifiedError, res.ReturnCode)
			},
		},
		{
			name: "topic exists",
			test: func(t *testing.T, s *store.Store) {
				s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo")))

				rr := mqtttest.NewRecorder()
				s.ServeMessage(rr, &mqtt.Message{
					Payload: &mqtt.ConnectRequest{
						ClientId: "foo",
						WillFlag: true,
						Topic:    "foo",
						Message:  []byte("bar"),
					},
				})
				res := rr.Message.Payload.(*mqtt.ConnectResponse)
				require.Equal(t, mqtt.Accepted, res.ReturnCode)
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
