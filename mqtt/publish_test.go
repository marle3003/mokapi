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

func TestPublish_ReadRequest(t *testing.T) {
	testcases := []struct {
		name string
		in   []byte
		test func(t *testing.T, r *mqtt.Message, err error)
	}{
		{
			name: "publish to foo",
			in: []byte{
				0x30,       // Protocol Type
				0x8,        // length
				0x00, 0x03, // topic length
				'f', 'o', 'o', // topic
				'b', 'a', 'r', // Payload
			},
			test: func(t *testing.T, r *mqtt.Message, err error) {
				require.NoError(t, err)

				require.Equal(t, 8, r.Header.Size)

				require.IsType(t, &mqtt.PublishRequest{}, r.Payload)
				msg := r.Payload.(*mqtt.PublishRequest)

				require.Equal(t, "foo", msg.Topic)
				require.Equal(t, "bar", string(msg.Data))
			},
		},
		{
			name: "publish to foo with QoS",
			in: []byte{
				0x32,       // Protocol Type
				0xA,        // length
				0x00, 0x03, // topic length
				'f', 'o', 'o', // Payload
				0x0, 0xA, // MessageId
				'b', 'a', 'r', // Payload
			},
			test: func(t *testing.T, r *mqtt.Message, err error) {
				require.NoError(t, err)

				require.Equal(t, 10, r.Header.Size)

				require.IsType(t, &mqtt.PublishRequest{}, r.Payload)
				msg := r.Payload.(*mqtt.PublishRequest)

				require.Equal(t, "foo", msg.Topic)
				//require.Equal(t, int16(10), msg.MessageId)
				require.Equal(t, "bar", string(msg.Data))
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

func TestPublish_Write(t *testing.T) {
	testcases := []struct {
		name string
		msg  mqtt.Message
		ctx  *mqtt.ClientContext
		out  []byte
	}{
		{
			name: "request QoS=0",
			msg: mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.PUBLISH,
				},
				Payload: &mqtt.PublishRequest{
					Topic:     "foo",
					MessageId: uint16(123),
					Data:      []byte("bar"),
				},
			},
			ctx: &mqtt.ClientContext{},
			out: []byte{
				0x30,      // Packet type
				0x08,      // length
				0x0, 0x03, // topic length
				'f', 'o', 'o',
				'b', 'a', 'r', // data
			},
		},
		{
			name: "request QoS=1",
			msg: mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.PUBLISH,
					QoS:  1,
				},
				Payload: &mqtt.PublishRequest{
					Topic:     "foo",
					MessageId: uint16(123),
					Data:      []byte("bar"),
				},
			},
			ctx: &mqtt.ClientContext{},
			out: []byte{
				0x32,      // Packet type
				0x0a,      // length
				0x0, 0x03, // topic length
				'f', 'o', 'o',
				0x0, 0x7b, // message id
				'b', 'a', 'r', // data
			},
		},
		{
			name: "request v5",
			msg: mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.PUBLISH,
					QoS:  1,
				},
				Payload: &mqtt.PublishRequest{
					Topic:     "foo",
					MessageId: uint16(123),
					Data:      []byte("bar"),
				},
			},
			ctx: &mqtt.ClientContext{ProtocolVersion: 5},
			out: []byte{
				0x32,      // Packet type
				0x0b,      // length
				0x0, 0x03, // topic length
				'f', 'o', 'o',
				0x0, 0x7b, // message id
				0x0,           // properties
				'b', 'a', 'r', // data
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

func TestPublish(t *testing.T) {
	testcases := []struct {
		name    string
		handler mqtt.Handler
		test    func(t *testing.T, s *mqtt.Server)
	}{
		{
			name: "publish to foo and PUBACK",
			handler: mqtt.HandlerFunc(func(rw mqtt.MessageWriter, req *mqtt.Message) {
				rw.Write(&mqtt.Message{
					Header: &mqtt.Header{
						Type: mqtt.PUBACK,
					},
					Payload: &mqtt.PublishResponse{
						MessageId: 10,
					},
				})
			}),
			test: func(t *testing.T, s *mqtt.Server) {
				c := mqtttest.NewClient(s.Addr)
				defer c.Close()
				r := &mqtt.Message{
					Header: &mqtt.Header{
						Type: mqtt.PUBLISH,
						QoS:  1,
					},
					Payload: &mqtt.PublishRequest{
						Topic:     "foo",
						MessageId: 10,
						Data:      []byte("bar"),
					},
				}
				res, err := c.Send(r)
				require.NoError(t, err)
				require.Equal(t, mqtt.PUBACK, res.Header.Type)
				msg := res.Payload.(*mqtt.PublishResponse)
				require.Equal(t, uint16(10), msg.MessageId)
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
