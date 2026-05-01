package mqtt_test

import (
	"bytes"
	"fmt"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/try"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscribe_ReadRequest(t *testing.T) {
	testcases := []struct {
		name string
		in   []byte
		test func(t *testing.T, r *mqtt.Message, err error)
	}{
		{
			name: "subscribe to foo",
			in: []byte{
				0x82,      // Protocol Type
				0x08,      // length
				0x0, 0x10, // Message Identifier
				0x00, 0x03, // topic length
				'f', 'o', 'o', // topic
				0x1, // QoS
			},
			test: func(t *testing.T, r *mqtt.Message, err error) {
				require.NoError(t, err)

				require.Equal(t, 8, r.Header.Size)

				require.IsType(t, &mqtt.SubscribeRequest{}, r.Payload)
				msg := r.Payload.(*mqtt.SubscribeRequest)

				require.Equal(t, uint16(16), msg.MessageId)
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

			r := &mqtt.Message{}
			err := r.Read(bytes.NewReader(tc.in), &mqtt.ClientContext{})
			tc.test(t, r, err)
		})
	}
}

func TestSubscribe_WriteResponse(t *testing.T) {
	testcases := []struct {
		name string
		msg  mqtt.Message
		ctx  *mqtt.ClientContext
		out  []byte
	}{
		{
			name: "simple subscribe",
			msg: mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.SUBACK,
				},
				Payload: &mqtt.SubscribeResponse{
					MessageId:   uint16(53902),
					ReasonCodes: []mqtt.SubscriptionReason{mqtt.GrantedQoS2},
				},
			},
			ctx: &mqtt.ClientContext{},
			out: []byte{
				0x90,       // Packet type
				0x03,       // length
				0xd2, 0x8e, // message id
				0x02, // reason
			},
		},
		{
			name: "subscribe v5",
			msg: mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.SUBACK,
				},
				Payload: &mqtt.SubscribeResponse{
					MessageId:   uint16(53902),
					ReasonCodes: []mqtt.SubscriptionReason{mqtt.GrantedQoS2},
				},
			},
			ctx: &mqtt.ClientContext{ProtocolVersion: 5},
			out: []byte{
				0x90,       // Packet type
				0x04,       // length
				0xd2, 0x8e, // message id
				0x00, // properties
				0x02, // reason
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var b bytes.Buffer
			err := tc.msg.Write(&b, tc.ctx)
			require.NoError(t, err)
			require.Equal(t, tc.out, b.Bytes())
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
			handler: mqtt.HandlerFunc(func(rw mqtt.MessageWriter, req *mqtt.Message) {
				rw.Write(&mqtt.Message{
					Header: &mqtt.Header{
						Type: mqtt.SUBACK,
					},
					Payload: &mqtt.SubscribeResponse{
						MessageId: 1,
						ReasonCodes: []mqtt.SubscriptionReason{
							mqtt.GrantedQoS1,
						},
					},
				})
			}),
			test: func(t *testing.T, s *mqtt.Server) {
				c := mqtttest.NewClient(s.Addr)
				defer c.Close()
				r := &mqtt.Message{
					Header: &mqtt.Header{
						Type: mqtt.SUBSCRIBE,
					},
					Payload: &mqtt.SubscribeRequest{
						Topics: []mqtt.SubscribeTopic{
							{
								Name: "foo",
								QoS:  1,
							},
						},
					},
				}
				res, err := c.Send(r)
				require.NoError(t, err)
				require.Equal(t, mqtt.SUBACK, res.Header.Type)
				msg := res.Payload.(*mqtt.SubscribeResponse)
				require.Equal(t, uint16(1), msg.MessageId)
				require.Len(t, msg.ReasonCodes, 1)
				require.Equal(t, mqtt.GrantedQoS1, msg.ReasonCodes[0])
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
