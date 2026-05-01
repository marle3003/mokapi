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

func TestUnsubscribe_ReadRequest(t *testing.T) {
	testcases := []struct {
		name string
		in   []byte
		test func(t *testing.T, r *mqtt.Message, err error)
	}{
		{
			name: "unsubscribe from foo",
			in: []byte{
				0xA0,       // Protocol Type
				0x07,       // length
				0x00, 0x03, // message id
				0x00, 0x03, // topic length
				'f', 'o', 'o', // topic
			},
			test: func(t *testing.T, r *mqtt.Message, err error) {
				require.NoError(t, err)

				require.Equal(t, 7, r.Header.Size)

				require.IsType(t, &mqtt.UnsubscribeRequest{}, r.Payload)
				msg := r.Payload.(*mqtt.UnsubscribeRequest)

				require.Equal(t, uint16(3), msg.MessageId)
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

			r := &mqtt.Message{}
			err := r.Read(bytes.NewReader(tc.in), &mqtt.ClientContext{})
			tc.test(t, r, err)
		})
	}
}

func TestUnsubscribe_WriteResponse(t *testing.T) {
	testcases := []struct {
		name string
		msg  mqtt.Message
		ctx  *mqtt.ClientContext
		out  []byte
	}{
		{
			name: "simple unsubscribe",
			msg: mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.UNSUBACK,
				},
				Payload: &mqtt.UnsubscribeResponse{
					MessageId:   uint16(12),
					ReasonCodes: []mqtt.UnsubscriptionReason{mqtt.UnsubscribeSuccess},
				},
			},
			ctx: &mqtt.ClientContext{},
			out: []byte{
				0xB0,       // Packet type
				0x02,       // length
				0x00, 0x0C, // message id
				// reason available in v5
			},
		},
		{
			name: "subscribe v5",
			msg: mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.UNSUBACK,
				},
				Payload: &mqtt.UnsubscribeResponse{
					MessageId:   uint16(12),
					ReasonCodes: []mqtt.UnsubscriptionReason{mqtt.UnsubscribeSuccess},
				},
			},
			ctx: &mqtt.ClientContext{ProtocolVersion: 5},
			out: []byte{
				0xB0,       // Packet type
				0x04,       // length
				0x00, 0x0C, // message id
				0x00, // properties
				0x00, // reason available in v5
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

func TestUnsubscribe(t *testing.T) {
	testcases := []struct {
		name    string
		handler mqtt.Handler
		test    func(t *testing.T, s *mqtt.Server)
	}{
		{
			name: "unsubscribe from foo",
			handler: mqtt.HandlerFunc(func(rw mqtt.MessageWriter, req *mqtt.Message) {
				rw.Write(&mqtt.Message{
					Header: &mqtt.Header{
						Type: mqtt.UNSUBACK,
					},
					Payload: &mqtt.UnsubscribeResponse{
						MessageId: 7,
						ReasonCodes: []mqtt.UnsubscriptionReason{
							mqtt.UnsubscribeSuccess,
						},
					},
				})
			}),
			test: func(t *testing.T, s *mqtt.Server) {
				c := mqtttest.NewClient(s.Addr)
				defer c.Close()
				r := &mqtt.Message{
					Header: &mqtt.Header{
						Type: mqtt.UNSUBSCRIBE,
					},
					Payload: &mqtt.UnsubscribeRequest{
						Topics: []string{"foo"},
					},
				}
				res, err := c.Send(r)
				require.NoError(t, err)
				require.Equal(t, mqtt.UNSUBACK, res.Header.Type)
				msg := res.Payload.(*mqtt.UnsubscribeResponse)
				require.Equal(t, uint16(7), msg.MessageId)
				// only used from version 5 onwards
				require.Len(t, msg.ReasonCodes, 0)
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
