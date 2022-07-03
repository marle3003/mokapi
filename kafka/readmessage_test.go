package kafka

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReadMessage_TwoMessages(t *testing.T) {
	b := []byte{
		0, 0, 0, 22, // length
		0, 3, // metadata
		0, 4, // version
		0, 0, 0, 2, // correlation id
		0, 3, 102, 111, 111, // client id: foo
		0, 0, 0, 0, // topics empty
		0,           // allow auto topic creation
		0, 0, 0, 23, // length
		0, 10, // find coordinator
		0, 2, // version
		0, 0, 0, 4, // correlation id
		0, 3, 102, 111, 111, // client id: foo
		0, 3, 98, 97, 114, // coordinator key: bar
		0, // coordinator type: group
	}
	r := bytes.NewReader(b)
	n := r.Len()
	_ = n
	h, msg, err := ReadMessage(r)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.Equal(t, Metadata, h.ApiKey)
	require.Equal(t, int16(4), h.ApiVersion)
	require.Equal(t, int32(2), h.CorrelationId)
	require.NotNil(t, msg)
	require.NotEqual(t, 0, r.Len())

	h, msg, err = ReadMessage(r)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.Equal(t, int32(23), h.Size)
	require.Equal(t, FindCoordinator, h.ApiKey)
	require.Equal(t, int16(2), h.ApiVersion)
	require.Equal(t, int32(4), h.CorrelationId)
	require.NotNil(t, msg)
}
