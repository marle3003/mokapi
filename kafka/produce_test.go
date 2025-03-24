package kafka_test

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"mokapi/kafka/produce"
	"testing"
)

func TestProduce_Request(t *testing.T) {
	testcases := []struct {
		name  string
		input []byte
		test  func(t *testing.T, r *kafka.Request, unread int, err error)
	}{
		{
			name: "v9",
			input: []byte{
				0x0, 0x0, 0x0, 0x73, // length: 115
				0x0, 0x0, // api key (produce)
				0x0, 0x9, // version
				0x0, 0x0, 0x0, 0x3, // correlation ID: 3
				0x0, 0x7, 0x72, 0x64, 0x6b, 0x61, 0x66, 0x6b, 0x61, // client ID: rdkafka
				0x0,        // tagged fields: header
				0x0,        // transactional ID
				0xff, 0xff, // required acks: Full ISR
				0x0, 0x0, 0x75, 0x30, // timeout: 30000
				0x2,              // topic array length
				0x4,              // topic name length
				0x66, 0x6f, 0x6f, // topic: foo
				0x2,                // partition array length
				0x0, 0x0, 0x0, 0x0, // partition ID: 0
				0x4d,                                   // message set size
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, // base offset
				0x0, 0x0, 0x0, 0x40, // message size
				0x0, 0x0, 0x0, 0x0, // leader epoch
				0x2,                   // magic byte: version 2
				0xe2, 0x9f, 0xf, 0x5f, // CRC32
				0x0, 0x0, // attributes
				0x0, 0x0, 0x0, 0x0, // last offset delta
				0x0, 0x0, 0x1, 0x94, 0x6e, 0x33, 0x13, 0x6e, // First Timestamp
				0x0, 0x0, 0x1, 0x94, 0x6e, 0x33, 0x13, 0x6e, // Last Timestamp
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // producer ID: -1
				0xff, 0xff, // producer epoch: -1
				0xff, 0xff, 0xff, 0xff, // base sequence
				0x0, 0x0, 0x0, 0x1, // number of records
				0x1c,                         // record length
				0x0,                          // attributes
				0x0,                          // timestamp delta
				0x0,                          // offset delta
				0xA,                          // key length
				0x31, 0x32, 0x33, 0x34, 0x35, // key: 12345
				0x6,              // value length
				0x6d, 0x73, 0x67, // value: msg
				0x0, // headers
				0x0, // tagged fields partitions
				0x0, // tagged fields topics
				0x0, // tagged fields request
			},
			test: func(t *testing.T, r *kafka.Request, unread int, err error) {
				require.NoError(t, err)
				require.Equal(t, 0, unread, "should read all bytes")
				p := r.Message.(*produce.Request)
				messageSet := p.Topics[0].Partitions[0].Record
				require.Equal(t, "12345", kafka.BytesToString(messageSet.Records[0].Key))
				require.Equal(t, "msg", kafka.BytesToString(messageSet.Records[0].Value))
				require.Equal(t, 14, messageSet.Records[0].Size(messageSet.Records[0].Offset, messageSet.Records[0].Time))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r := &kafka.Request{}
			reader := bytes.NewReader(tc.input)
			err := r.Read(reader)
			unread := reader.Len()
			tc.test(t, r, unread, err)
		})
	}
}
