package kafka_test

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"mokapi/kafka/offsetFetch"
	"mokapi/kafka/produce"
	"testing"
	"time"
)

func TestRequest_Read_OffsetFetch(t *testing.T) {
	b := []byte{
		0, 0, 0, 37, // length
		0, 9, // OffsetFetch
		0, 3, // version
		0, 0, 0, 2, // correlation id
		0, 3, 'f', 'o', 'o', // client id: foo
		0, 3, 'b', 'a', 'r', // consumer group: bar
		0, 0, 0, 1, // topics length
		0, 5, 't', 'o', 'p', 'i', 'c', // topic  name
		0, 0, 0, 1, // partition indexes length
		0, 0, 39, 15, // index: 9999
	}

	r := &kafka.Request{}
	reader := bytes.NewReader(b)
	err := r.Read(reader)
	require.NoError(t, err)

	require.NotNil(t, r.Header)
	require.Equal(t, int32(37), r.Header.Size)
	require.Equal(t, kafka.OffsetFetch, r.Header.ApiKey)
	require.Equal(t, int16(3), r.Header.ApiVersion)
	require.Equal(t, int32(2), r.Header.CorrelationId)
	require.Equal(t, "foo", r.Header.ClientId)
	require.NotNil(t, r.Message)
	msg := r.Message.(*offsetFetch.Request)
	require.Equal(t, "bar", msg.GroupId)
	require.Len(t, msg.Topics, 1)
	require.Equal(t, "topic", msg.Topics[0].Name)
	require.Len(t, msg.Topics[0].PartitionIndexes, 1)
	require.Equal(t, int32(9999), msg.Topics[0].PartitionIndexes[0])
	require.Equal(t, 0, reader.Len())
}

func TestRequest_Write_OffsetFetch(t *testing.T) {
	r := kafka.Request{
		Header: &kafka.Header{
			ApiKey:        kafka.OffsetFetch,
			ApiVersion:    3,
			CorrelationId: 2,
			ClientId:      "foo",
		},
		Message: &offsetFetch.Request{
			GroupId: "bar",
			Topics: []offsetFetch.RequestTopic{{
				Name:             "foo",
				PartitionIndexes: []int32{9999},
			}},
		},
	}
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := r.Write(w)
	require.NoError(t, err)
	err = w.Flush()
	require.NoError(t, err)

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

func TestRequest_Write_Produce(t *testing.T) {
	data := bytes.Repeat([]byte("foobar"), 65536)
	v := kafka.NewBytes(data)
	r := kafka.Request{
		Header: &kafka.Header{
			ApiKey:        kafka.Produce,
			ApiVersion:    3,
			CorrelationId: 2,
			ClientId:      "foo",
		},
		Message: &produce.Request{
			Topics: []produce.RequestTopic{{
				Name: "foo",
				Partitions: []produce.RequestPartition{
					{Index: 0,
						Record: kafka.RecordBatch{
							Records: []kafka.Record{
								{Offset: 0,
									Time:  time.Time{},
									Key:   kafka.NewBytes([]byte{}),
									Value: v,
								},
							},
						},
					},
				},
			}},
		},
	}
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := r.Write(w)
	require.NoError(t, err)
	err = w.Flush()
	require.NoError(t, err)

	require.Equal(t, 393334, b.Len())
}
