package store_test

import (
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestCleaner(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "rolling",
			test: func(t *testing.T) {
				hook := test.NewGlobal()

				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithServer("foo", "kafka", "",
						asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{
							LogRetentionCheckIntervalMs: 500,
							LogRollMs:                   10,
							LogRetentionMs:              500,
						}),
					),
					asyncapi3test.WithChannel("foo"),
				)
				s := store.New(cfg, enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewKafka())

				topic := s.Topic("foo")
				require.NotNil(t, topic)
				rErr, err := topic.WritePartition(0, &kafka.Record{
					Offset: 0,
					Key:    kafka.NewBytes([]byte("key-1")),
				})
				require.NoError(t, err)
				require.Nil(t, rErr)

				time.Sleep(800 * time.Millisecond)
				p := topic.Partitions[0]
				require.NotNil(t, p)
				require.Len(t, p.Segments, 1)
				require.False(t, p.Segments[0].Closed.IsZero())

				time.Sleep(time.Second)
				require.Len(t, p.Segments, 0)

				require.Equal(t, "kafka: deleting segment with offset [0:1] from partition 0 topic 'foo'", hook.LastEntry().Message)
			},
		},
		{
			name: "retention bytes",
			test: func(t *testing.T) {
				hook := test.NewGlobal()

				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithServer("foo", "kafka", "",
						asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{
							LogRetentionCheckIntervalMs: 500,
							LogRollMs:                   10,
							LogRetentionBytes:           1,
						}),
					),
					asyncapi3test.WithChannel("foo"),
				)
				s := store.New(cfg, enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewKafka())

				topic := s.Topic("foo")
				require.NotNil(t, topic)
				rErr, err := topic.WritePartition(0, &kafka.Record{
					Offset: 0,
					Key:    kafka.NewBytes([]byte("key-1")),
				})
				require.NoError(t, err)
				require.Nil(t, rErr)

				time.Sleep(800 * time.Millisecond)
				p := topic.Partitions[0]
				require.NotNil(t, p)
				require.Len(t, p.Segments, 0)

				require.Equal(t, "kafka: maximum partition size reached. deleting segment [0:1] from partition 0 of topic 'foo'", hook.LastEntry().Message)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
