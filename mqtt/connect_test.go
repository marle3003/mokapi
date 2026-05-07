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

func TestConnect_ReadRequest(t *testing.T) {
	testcases := []struct {
		name string
		in   []byte
		test func(t *testing.T, r *mqtt.Message, err error)
	}{
		{
			name: "simple connect",
			in: []byte{
				0x10,       // Packet type
				0x10,       // length
				0x00, 0x04, // protocol length
				0x4d, 0x51, 0x54, 0x54, // protocol name MQTT
				0x04,       // version
				0x02,       // connect flags
				0x00, 0x3c, // keep alive
				0x00, 0x04, // client id length
				0x6d, 0x71, 0x74, 0x74, // client id
			},
			test: func(t *testing.T, r *mqtt.Message, err error) {
				require.NoError(t, err)

				require.Equal(t, 16, r.Header.Size)

				require.IsType(t, &mqtt.ConnectRequest{}, r.Payload)
				msg := r.Payload.(*mqtt.ConnectRequest)

				require.Equal(t, "MQTT", msg.Protocol)
				require.Equal(t, byte(4), msg.Version)

				require.False(t, msg.HasUsername)
				require.False(t, msg.HasPassword)
				require.False(t, msg.WillRetain)
				require.Equal(t, byte(0), msg.WillQoS)
				require.False(t, msg.WillFlag)
				require.True(t, msg.CleanSession)
				require.Equal(t, int16(60), msg.KeepAlive)
				require.Equal(t, "mqtt", msg.ClientId)
			},
		},
		{
			name: "connect with topic and message",
			in: []byte{
				0x10,       // Packet type
				0x1A,       // length
				0x00, 0x04, // protocol length
				0x4d, 0x51, 0x54, 0x54, // protocol name MQTT
				0x04,       // version
				0x0e,       // connect flags
				0x00, 0x3c, // keep alive
				0x00, 0x04, // client id length
				0x6d, 0x71, 0x74, 0x74, // client id
				0x00, 0x03, // topic length
				'f', 'o', 'o', // topic
				0x00, 0x03, // message length
				'b', 'a', 'r', // message
			},
			test: func(t *testing.T, r *mqtt.Message, err error) {
				require.NoError(t, err)

				require.Equal(t, 26, r.Header.Size)

				require.IsType(t, &mqtt.ConnectRequest{}, r.Payload)
				msg := r.Payload.(*mqtt.ConnectRequest)

				require.Equal(t, "MQTT", msg.Protocol)
				require.Equal(t, byte(4), msg.Version)

				require.False(t, msg.HasUsername)
				require.False(t, msg.HasPassword)
				require.False(t, msg.WillRetain)
				require.Equal(t, byte(1), msg.WillQoS)
				require.True(t, msg.WillFlag)
				require.True(t, msg.CleanSession)
				require.Equal(t, int16(60), msg.KeepAlive)
				require.Equal(t, "mqtt", msg.ClientId)
				require.Equal(t, "foo", msg.Topic)
				require.Equal(t, []byte("bar"), msg.Message)
			},
		},
		{
			name: "connect v5",
			in: []byte{
				0x10,       // Packet type
				0x18,       // length
				0x00, 0x04, // protocol length
				0x4d, 0x51, 0x54, 0x54, // protocol name MQTT
				0x05,       // version
				0x02,       // connect flags
				0x00, 0x3c, // keep alive
				0x5,                // properties length
				0x11,               // Session Expiry Interval
				0x0, 0x0, 0x0, 0x0, // value
				0x00, 0x06, // client id length
				'm', 'o', 'k', 'a', 'p', 'i', // client id
			},
			test: func(t *testing.T, r *mqtt.Message, err error) {
				require.NoError(t, err)

				require.Equal(t, 24, r.Header.Size)

				require.IsType(t, &mqtt.ConnectRequest{}, r.Payload)
				msg := r.Payload.(*mqtt.ConnectRequest)

				require.Equal(t, "MQTT", msg.Protocol)
				require.Equal(t, byte(5), msg.Version)

				require.False(t, msg.HasUsername)
				require.False(t, msg.HasPassword)
				require.False(t, msg.WillRetain)
				require.Equal(t, byte(0), msg.WillQoS)
				require.False(t, msg.WillFlag)
				require.True(t, msg.CleanSession)
				require.Equal(t, int16(60), msg.KeepAlive)
				require.Contains(t, msg.Properties, mqtt.SessionExpiryInterval)
				require.Equal(t, int32(0), msg.Properties[mqtt.SessionExpiryInterval])
				require.Equal(t, "mokapi", msg.ClientId)
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

func TestConnect_WriteRequest(t *testing.T) {
	testcases := []struct {
		name string
		msg  mqtt.Message
		out  []byte
	}{
		{
			name: "simple connect",
			msg: mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.CONNECT,
				},
				Payload: &mqtt.ConnectRequest{
					Protocol:     "MQTT",
					Version:      4,
					CleanSession: true,
					KeepAlive:    60,
					ClientId:     "mqtt",
				},
			},
			out: []byte{
				0x10,       // Packet type
				0x10,       // length
				0x00, 0x04, // protocol length
				0x4d, 0x51, 0x54, 0x54, // protocol name MQTT
				0x04,       // version
				0x02,       // connect flags
				0x00, 0x3c, // keep alive
				0x00, 0x04, // client id length
				0x6d, 0x71, 0x74, 0x74, // client id
			},
		},
		{
			name: "connect with topic and message",
			msg: mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.CONNECT,
				},
				Payload: &mqtt.ConnectRequest{
					Protocol:     "MQTT",
					Version:      4,
					WillQoS:      1,
					CleanSession: true,
					WillFlag:     true,
					KeepAlive:    60,
					ClientId:     "mqtt",
					Topic:        "foo",
					Message:      []byte("bar"),
				},
			},
			out: []byte{
				0x10,       // Packet type
				0x1A,       // length
				0x00, 0x04, // protocol length
				0x4d, 0x51, 0x54, 0x54, // protocol name MQTT
				0x04,       // version
				0x0e,       // connect flags
				0x00, 0x3c, // keep alive
				0x00, 0x04, // client id length
				0x6d, 0x71, 0x74, 0x74, // client id
				0x00, 0x03, // topic length
				'f', 'o', 'o', // topic
				0x00, 0x03, // message length
				'b', 'a', 'r', // message
			},
		},
		{
			name: "connect v5",
			msg: mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.CONNECT,
				},
				Payload: &mqtt.ConnectRequest{
					Protocol:     "MQTT",
					Version:      5,
					CleanSession: true,
					KeepAlive:    60,
					ClientId:     "mokapi",
					Properties: mqtt.Properties{
						mqtt.SessionExpiryInterval: int32(0),
					},
				},
			},
			out: []byte{
				0x10,       // Packet type
				0x18,       // length
				0x00, 0x04, // protocol length
				0x4d, 0x51, 0x54, 0x54, // protocol name MQTT
				0x05,       // version
				0x02,       // connect flags
				0x00, 0x3c, // keep alive
				0x5,                // properties length
				0x11,               // Session Expiry Interval
				0x0, 0x0, 0x0, 0x0, // value
				0x00, 0x06, // client id length
				'm', 'o', 'k', 'a', 'p', 'i', // client id
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var b bytes.Buffer
			err := tc.msg.Write(&b, &mqtt.ClientContext{})
			require.NoError(t, err)
			require.Equal(t, tc.out, b.Bytes())
		})
	}
}

func TestConnect(t *testing.T) {
	testcases := []struct {
		name    string
		handler mqtt.Handler
		test    func(t *testing.T, s *mqtt.Server)
	}{
		{
			name: "simple connect",
			handler: mqtt.HandlerFunc(func(w mqtt.MessageWriter, req *mqtt.Message) {
				w.Write(&mqtt.Message{
					Header: &mqtt.Header{
						Type: mqtt.CONNACK,
					},
					Payload: &mqtt.ConnectResponse{
						SessionPresent: false,
						ReasonCode:     mqtt.Success,
					},
				})
			}),
			test: func(t *testing.T, s *mqtt.Server) {
				c := mqtttest.NewClient(s.Addr)
				defer c.Close()
				r := &mqtt.Message{
					// The DUP, QoS, and RETAIN flags are not used in the CONNECT message.
					Header: &mqtt.Header{
						Type: mqtt.CONNECT,
					},
					Payload: &mqtt.ConnectRequest{
						Protocol:     "MQTT",
						Version:      4,
						CleanSession: true,
						KeepAlive:    60,
						ClientId:     "client-foo",
					},
				}
				res, err := c.Send(r)
				require.NoError(t, err)
				require.Equal(t, mqtt.CONNACK, res.Header.Type)
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
