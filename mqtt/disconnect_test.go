package mqtt_test

import (
	"bytes"
	"mokapi/mqtt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisconnect_ReadRequest(t *testing.T) {
	testcases := []struct {
		name string
		in   []byte
		ctx  *mqtt.ClientContext
		test func(t *testing.T, r *mqtt.Message, err error)
	}{
		{
			name: "disconnect no reason",
			in: []byte{
				0xE0, // Protocol Type
				0x0,  // length
			},
			ctx: &mqtt.ClientContext{},
			test: func(t *testing.T, r *mqtt.Message, err error) {
				require.NoError(t, err)
				require.IsType(t, &mqtt.DisconnectRequest{}, r.Payload)
				msg := r.Payload.(*mqtt.DisconnectRequest)
				require.Equal(t, mqtt.DisconnectNormal, msg.Reason)
			},
		},
		{
			name: "disconnect no reason",
			in: []byte{
				0xE0,      // Protocol Type
				0x8,       // Remaining Length: 1 (Reason) + 1 (PropLen) + 3 (StrLen) + 3 (Data) = 8
				0x80,      // Reason: Unspecified error
				0x06,      // Property length
				0x1F,      // Property ID: Reason String
				0x0, 0x03, // String length
				'f', 'o', 'o',
			},
			ctx: &mqtt.ClientContext{ProtocolVersion: 5},
			test: func(t *testing.T, r *mqtt.Message, err error) {
				require.NoError(t, err)
				require.IsType(t, &mqtt.DisconnectRequest{}, r.Payload)
				msg := r.Payload.(*mqtt.DisconnectRequest)
				require.Equal(t, mqtt.DisconnectReason(128), msg.Reason)
				require.Equal(t, "foo", msg.Properties.ReasonString())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &mqtt.Message{}
			err := r.Read(bytes.NewReader(tc.in), tc.ctx)
			tc.test(t, r, err)
		})
	}
}
