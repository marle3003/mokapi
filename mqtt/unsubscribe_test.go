package mqtt_test

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/try"
	"testing"
)

func TestUnsubscribe_ReadRequest(t *testing.T) {
	testcases := []struct {
		name string
		in   []byte
		test func(t *testing.T, r *mqtt.Request, err error)
	}{
		{
			name: "unsubscribe from foo",
			in: []byte{
				0xA0,     // flags
				0x07,     // length
				0x0, 0xA, // MessageId
				0x00, 0x03, // topic length
				'f', 'o', 'o', // topic
			},
			test: func(t *testing.T, r *mqtt.Request, err error) {
				require.NoError(t, err)

				require.Equal(t, 7, r.Header.Size)

				require.IsType(t, &mqtt.UnsubscribeRequest{}, r.Message)
				msg := r.Message.(*mqtt.UnsubscribeRequest)

				require.Len(t, msg.Topics, 1)
				require.Equal(t, "foo", msg.Topics[0])
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &mqtt.Request{}
			err := r.Read(bytes.NewReader(tc.in))
			tc.test(t, r, err)
		})
	}
}

func TestUnsubscribe(t *testing.T) {
	testcases := []struct {
		name    string
		handler mqtt.Handler
		test    func(t *testing.T, s *mqtt.Server)
	}{
		{
			name: "unsubscribe from foo",
			handler: mqtt.HandlerFunc(func(rw mqtt.ResponseWriter, req *mqtt.Request) {
				rw.Write(mqtt.UNSUBACK, &mqtt.UnsubscribeResponse{
					MessageId: 10,
				})
			}),
			test: func(t *testing.T, s *mqtt.Server) {
				c := mqtttest.NewClient(s.Addr)
				defer c.Close()
				r := &mqtt.Request{
					Header: &mqtt.Header{
						Type: mqtt.UNSUBSCRIBE,
					},
					Message: &mqtt.UnsubscribeRequest{
						MessageId: 10,
						Topics:    []string{"foo"},
					},
				}
				res, err := c.Send(r)
				require.NoError(t, err)
				require.Equal(t, mqtt.UNSUBACK, res.Header.Type)
				msg := res.Message.(*mqtt.UnsubscribeResponse)
				require.Equal(t, int16(10), msg.MessageId)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			port := try.GetFreePort()
			addr := fmt.Sprintf("127.0.0.1:%v", port)
			s := &mqtt.Server{
				Addr:    addr,
				Handler: tc.handler,
			}
			go s.ListenAndServe()
			defer s.Close()

			tc.test(t, s)

		})
	}
}
