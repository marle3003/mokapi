package offsetFetch

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"testing"
)

func TestReadMessage(t *testing.T) {
	b := []byte{
		0, 0, 0, 35, // length
		0, 9, // OffsetFetch
		0, 3, // version
		0, 0, 0, 2, // correlation id
		0, 3, 102, 111, 111, // client id: foo
		0, 3, 98, 97, 114, // consumer group: bar
		0, 0, 0, 1, // topics length
		0, 3, 102, 111, 111, // topic  name
		0, 0, 0, 1, // partition indexes length
		0, 0, 39, 15, // index: 9999
	}
	r := bytes.NewReader(b)
	h, msg, err := kafka.ReadMessage(r)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.Equal(t, kafka.OffsetFetch, h.ApiKey)
	require.Equal(t, int16(3), h.ApiVersion)
	require.Equal(t, int32(2), h.CorrelationId)
	require.NotNil(t, msg)
	of := msg.(*Request)
	require.Equal(t, "bar", of.GroupId)
	require.Equal(t, 0, r.Len())
}

func TestWriteMessage(t *testing.T) {
	r := kafka.Request{
		Header: &kafka.Header{
			ApiKey:        kafka.OffsetFetch,
			ApiVersion:    3,
			CorrelationId: 2,
			ClientId:      "foo",
		},
		Message: &Request{
			GroupId: "bar",
			Topics: []RequestTopic{{
				Name:             "foo",
				PartitionIndexes: []int32{9999},
			}},
		},
	}
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := r.Write(w)
	require.NoError(t, err)
	w.Flush()

	expected := []byte{
		0, 0, 0, 35, // length
		0, 9, // OffsetFetch
		0, 3, // version
		0, 0, 0, 2, // correlation id
		0, 3, 102, 111, 111, // client id: foo
		0, 3, 98, 97, 114, // consumer group: bar
		0, 0, 0, 1, // topics length
		0, 3, 102, 111, 111, // topic  name
		0, 0, 0, 1, // partition indexes length
		0, 0, 39, 15, // index: 9999
	}
	require.Equal(t, expected, b.Bytes())
}
