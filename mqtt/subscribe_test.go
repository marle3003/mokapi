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

func TestSubscribe_ReadRequest(t *testing.T) {
	testcases := []struct {
		name string
		in   []byte
		test func(t *testing.T, r *mqtt.Request, err error)
	}{
		{
			name: "subscribe to foo",
			in: []byte{
				0x82,     // flags
				0x08,     // length
				0x0, 0xA, // MessageId
				0x00, 0x03, // topic length
				'f', 'o', 'o', // topic
				0x1, // QoS
			},
			test: func(t *testing.T, r *mqtt.Request, err error) {
				require.NoError(t, err)

				require.Equal(t, 8, r.Header.Size)

				require.IsType(t, &mqtt.SubscribeRequest{}, r.Message)
				msg := r.Message.(*mqtt.SubscribeRequest)

				require.Len(t, msg.Topics, 1)
				require.Equal(t, "foo", msg.Topics[0].Name)
				require.Equal(t, byte(1), msg.Topics[0].QoS)
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

func TestSubscribe(t *testing.T) {
	testcases := []struct {
		name    string
		handler mqtt.Handler
		test    func(t *testing.T, s *mqtt.Server)
	}{
		{
			name: "subscribe to foo",
			handler: mqtt.HandlerFunc(func(rw mqtt.ResponseWriter, req *mqtt.Request) {
				rw.Write(mqtt.SUBACK, &mqtt.SubscribeResponse{
					MessageId: 10,
					TopicQoS: []byte{
						byte(1),
					},
				})
			}),
			test: func(t *testing.T, s *mqtt.Server) {
				c := mqtttest.NewClient(s.Addr)
				defer c.Close()
				r := &mqtt.Request{
					// The DUP, QoS, and RETAIN flags are not used in the CONNECT message.
					Header: &mqtt.Header{
						Type: mqtt.SUBSCRIBE,
					},
					Message: &mqtt.SubscribeRequest{
						MessageId: 10,
						Topics: []mqtt.SubscribeTopic{
							{
								Name: "foo",
								QoS:  1,
							},
						},
					},
					Context: nil,
				}
				res, err := c.Send(r)
				require.NoError(t, err)
				require.Equal(t, mqtt.SUBACK, res.Header.Type)
				msg := res.Message.(*mqtt.SubscribeResponse)
				require.Equal(t, int16(10), msg.MessageId)
				require.Len(t, msg.TopicQoS, 1)
				require.Equal(t, byte(1), msg.TopicQoS[0])
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
