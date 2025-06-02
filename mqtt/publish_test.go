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

func TestPublish_ReadRequest(t *testing.T) {
	testcases := []struct {
		name string
		in   []byte
		test func(t *testing.T, r *mqtt.Request, err error)
	}{
		{
			name: "publish to foo",
			in: []byte{
				0x30,       // flags
				0xA,        // length
				0x00, 0x03, // topic length
				'f', 'o', 'o', // Payload
				0x0, 0xA, // MessageId
				'b', 'a', 'r', // Payload
			},
			test: func(t *testing.T, r *mqtt.Request, err error) {
				require.NoError(t, err)

				require.Equal(t, 10, r.Header.Size)

				require.IsType(t, &mqtt.PublishRequest{}, r.Message)
				msg := r.Message.(*mqtt.PublishRequest)

				require.Equal(t, "foo", msg.Topic)
				require.Equal(t, int16(10), msg.MessageId)
				require.Equal(t, []byte("bar"), msg.Data)
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

func TestPublish(t *testing.T) {
	testcases := []struct {
		name    string
		handler mqtt.Handler
		test    func(t *testing.T, s *mqtt.Server)
	}{
		{
			name: "publish to foo and PUBACK",
			handler: mqtt.HandlerFunc(func(rw mqtt.ResponseWriter, req *mqtt.Request) {
				rw.Write(mqtt.PUBACK, &mqtt.PublishResponse{
					MessageId: 10,
				})
			}),
			test: func(t *testing.T, s *mqtt.Server) {
				c := mqtttest.NewClient(s.Addr)
				defer c.Close()
				r := &mqtt.Request{
					Header: &mqtt.Header{
						Type: mqtt.PUBLISH,
					},
					Message: &mqtt.PublishRequest{
						Topic:     "foo",
						MessageId: 10,
						Data:      []byte("bar"),
					},
				}
				res, err := c.Send(r)
				require.NoError(t, err)
				require.Equal(t, mqtt.PUBACK, res.Header.Type)
				msg := res.Message.(*mqtt.PublishResponse)
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
