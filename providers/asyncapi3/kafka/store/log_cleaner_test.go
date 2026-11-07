package store_test

import (
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"sync"
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
				require.Equal(t, 1, p.NumSegments())
				require.True(t, p.IsSegmentClosed(0))

				time.Sleep(time.Second)
				require.Equal(t, 0, p.NumSegments())

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
				require.Equal(t, 0, p.NumSegments())

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

func aggressiveCleanerConfig(partitions int) *asyncapi3.Config {
	return asyncapi3test.NewConfig(
		asyncapi3test.WithServer("foo", "kafka", "",
			asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{
				LogRetentionCheckIntervalMs: 1,
				LogRollMs:                   1,
				LogRetentionMs:              1,
			}),
		),
		asyncapi3test.WithChannel("foo",
			asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: partitions}),
		),
	)
}

func newStoreWithAggressiveCleaner(partitions int) *store.Store {
	return store.New(aggressiveCleanerConfig(partitions), enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewKafka())
}

func TestCleaner_RunsWhileRecordsAreWrittenAndRead(t *testing.T) {
	s := newStoreWithAggressiveCleaner(1)
	defer s.Close()
	p := s.Topic("foo").Partition(0)

	var wg sync.WaitGroup
	stop := make(chan struct{})
	for _, reader := range []func(){
		func() { _, _ = p.Read(p.StartOffset(), 1000) },
		func() { _ = p.OffsetTimestamp(p.StartOffset()) },
		func() { _ = p.Offset() },
	} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					reader()
				}
			}
		}()
	}

	for i := 0; i < 300; i++ {
		_, err := p.Write(kafka.RecordBatch{Records: []*kafka.Record{{Key: kafka.NewBytes([]byte("key"))}}})
		require.NoError(t, err)
		time.Sleep(time.Millisecond)
	}
	close(stop)
	wg.Wait()

	require.Equal(t, int64(300), p.Offset())
}

func TestCleaner_RunsWhileStoreIsUpdated(t *testing.T) {
	s := newStoreWithAggressiveCleaner(1)
	defer s.Close()
	for i := 0; i < 50; i++ {
		_, rErr := s.Topic("foo").Write(&kafka.Record{Key: kafka.NewBytes([]byte("key"))})
		require.Nil(t, rErr)
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				// the cleaner works on partitions that are added and removed by Update
				for _, p := range s.Topic("foo").PartitionList() {
					_ = p.Offset()
				}
				_ = s.Topic("foo").NumPartitions()
				_ = s.Topics()
			}
		}
	}()

	for i := 0; i < 100; i++ {
		s.Update(aggressiveCleanerConfig(1 + i%3))
		time.Sleep(time.Millisecond)
	}
	close(stop)
	wg.Wait()

	require.Equal(t, 1+99%3, s.Topic("foo").NumPartitions())
}
