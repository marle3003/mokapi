package kafka_test

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"mokapi/kafka/fetch"
	"testing"
	"time"
)

func TestFetch_Response(t *testing.T) {
	testcases := []struct {
		name     string
		response *kafka.Response
		test     func(t *testing.T, b []byte, err error)
	}{
		{
			name: "v9",
			response: &kafka.Response{
				Header: &kafka.Header{
					ApiKey:        kafka.Fetch,
					ApiVersion:    int16(12),
					CorrelationId: int32(0),
				},
				Message: &fetch.Response{
					Topics: []fetch.ResponseTopic{
						{
							Name: "foo",
							Partitions: []fetch.ResponsePartition{
								{
									Index: 0,
									RecordSet: kafka.RecordBatch{
										Records: []*kafka.Record{
											{
												Offset:  0,
												Time:    time.Date(2025, time.Month(1), 16, 10, 10, 0, 0, time.UTC),
												Key:     kafka.NewBytes([]byte("12345")),
												Value:   kafka.NewBytes([]byte("msg")),
												Headers: nil,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Len(t, b, 140)
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x88}, b[0:4])                           // length 140
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, b[4:8])                            // correlation ID
				require.Equal(t, []byte{0x0}, b[8:9])                                           // tagged fields
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, b[9:13])                           // throttle time
				require.Equal(t, []byte{0x0, 0x0}, b[13:15])                                    // error code
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, b[15:19])                          // session id
				require.Equal(t, []byte{0x2}, b[19:20])                                         // topic length
				require.Equal(t, []byte{0x4}, b[20:21])                                         // topic name length
				require.Equal(t, []byte{0x66, 0x6f, 0x6f}, b[21:24])                            // topic name
				require.Equal(t, []byte{0x2}, b[24:25])                                         // partition length
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, b[25:29])                          // partition index
				require.Equal(t, []byte{0x0, 0x0}, b[29:31])                                    // error code
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, b[31:39])      // high watermark
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, b[39:47])      // last stable offset
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, b[47:55])      // log start offset
				require.Equal(t, []byte{0x1}, b[55:56])                                         // aborted transactions length
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, b[56:60])                          // preferred read replica
				require.Equal(t, []byte{0x4d}, b[60:61])                                        // message set size
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, b[61:69])      // base offset
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x40}, b[69:73])                         // size
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, b[73:77])                          // leader epoch
				require.Equal(t, []byte{0x2}, b[77:78])                                         // magic byte version
				require.Equal(t, []byte{0xc8, 0x9, 0xb5, 0x90}, b[78:82])                       // checksum CRC32
				require.Equal(t, []byte{0x0, 0x0}, b[82:84])                                    // attributes
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0}, b[84:88])                          // last offset delta
				require.Equal(t, []byte{0x0, 0x0, 0x1, 0x94, 0x6e, 0x97, 0x58, 0xc0}, b[88:96]) // first timestamp
				i := binary.BigEndian.Uint64(b[88:96])
				require.Equal(t, "2025-01-16 10:10:00 +0000 UTC", kafka.ToTime(int64(i)).UTC().String()) // max timestamp
				i = binary.BigEndian.Uint64(b[96:104])
				require.Equal(t, "2025-01-16 10:10:00 +0000 UTC", kafka.ToTime(int64(i)).UTC().String()) // max timestamp

				require.Equal(t, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, b[104:112]) // producer Id
				require.Equal(t, []byte{0xff, 0xff}, b[112:114])                                     // producer epoch
				require.Equal(t, []byte{0xff, 0xff, 0xff, 0xff}, b[114:118])                         // base sequence
				require.Equal(t, []byte{0x0, 0x0, 0x0, 0x1}, b[118:122])                             // number of records

				require.Equal(t, []byte{0x1c}, b[122:123]) // record size
				require.Equal(t, []byte{0x0}, b[123:124])  // attribute
				require.Equal(t, []byte{0x0}, b[124:125])  // timestamp delta
				require.Equal(t, []byte{0x0}, b[125:126])  // offset delta

				require.Equal(t, []byte{0xA}, b[126:127])     // key length
				require.Equal(t, []byte("12345"), b[127:132]) // key

				require.Equal(t, []byte{0x6}, b[132:133])   // value length
				require.Equal(t, []byte("msg"), b[133:136]) // value

				require.Equal(t, []byte{0x0}, b[136:137]) // headers

				require.Equal(t, []byte{0x0}, b[137:138]) // tagged fields partitions
				require.Equal(t, []byte{0x0}, b[138:139]) // tagged fields topics
				require.Equal(t, []byte{0x0}, b[139:140]) // tagged fields response
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			w := bufio.NewWriter(&b)
			err := tc.response.Write(w)
			require.NoError(t, err)
			err = w.Flush()
			require.NoError(t, err)
			tc.test(t, b.Bytes(), err)
		})
	}
}
