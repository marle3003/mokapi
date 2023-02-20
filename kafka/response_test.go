package kafka_test

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"mokapi/kafka/offset"
	"testing"
)

func TestResponse_Write(t *testing.T) {
	msg := &offset.Response{
		ThrottleTimeMs: 0,
		Topics: []offset.ResponseTopic{
			{
				Name: "foo",
				Partitions: []offset.ResponsePartition{
					{
						Index:           0,
						ErrorCode:       0,
						OldStyleOffsets: []int64{1},
					},
				},
			},
		},
	}

	res := &kafka.Response{
		Header: &kafka.Header{
			ApiKey:        kafka.Offset,
			ApiVersion:    int16(0),
			CorrelationId: int32(0),
		},
		Message: msg,
	}

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := res.Write(w)
	require.NoError(t, err)
	err = w.Flush()
	require.NoError(t, err)

	expected := []byte{
		0, 0, 0, 35, // length
		0, 0, 0, 0, // correlation id
		0, 0, 0, 1, // topics length
		0, 3, 102, 111, 111, // topic name: foo
		0, 0, 0, 1, // partitions length
		0, 0, 0, 0, // partition index
		0, 0, // error code
		0, 0, 0, 1, // offsets length
		0, 0, 0, 0, 0, 0, 0, 1, // old style offset
	}
	require.Equal(t, expected, b.Bytes())
}
