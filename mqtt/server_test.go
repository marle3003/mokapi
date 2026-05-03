package mqtt_test

import (
	"errors"
	"fmt"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/try"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	testcases := []struct {
		name    string
		handler func(t *testing.T) mqtt.HandlerFunc
		test    func(t *testing.T, c *mqtttest.Client)
	}{
		{
			name: "Ping",
			handler: func(t *testing.T) mqtt.HandlerFunc {
				return func(rw mqtt.MessageWriter, m *mqtt.Message) {
					err := rw.Write(&mqtt.Message{
						Header: &mqtt.Header{
							Type: mqtt.PINGRESP,
						},
						Payload: &mqtt.PingResponse{},
					})
					require.NoError(t, err)
				}
			},
			test: func(t *testing.T, c *mqtttest.Client) {
				res, err := c.Send(&mqtt.Message{
					Header: &mqtt.Header{
						Type: mqtt.PINGREQ,
					},
					Payload: &mqtt.PingRequest{},
				})
				require.NoError(t, err)
				require.NotNil(t, res)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			addr := fmt.Sprintf(":%d", try.GetFreePort())
			s := &mqtt.Server{
				Addr:    addr,
				Handler: tc.handler(t),
			}
			go func() {
				err := s.ListenAndServe()
				if err != nil && !errors.Is(err, mqtt.ErrServerClosed) {
					panic(err)
				}
			}()
			defer s.Close()

			c := mqtttest.NewClient(addr)
			tc.test(t, c)
		})
	}
}
