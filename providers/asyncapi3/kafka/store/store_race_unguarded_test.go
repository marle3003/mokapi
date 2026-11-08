package store_test

import (
	"fmt"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/heartbeat"
	"mokapi/kafka/initProducerId"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/metaData"
	"mokapi/kafka/produce"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"sync"
	"testing"
)

func brokerAndTopicConfig(host string) *asyncapi3.Config {
	return asyncapi3test.NewConfig(
		asyncapi3test.WithInfo("test", "", ""),
		asyncapi3test.WithServer("a", "kafka", host),
		asyncapi3test.WithChannel("foo"),
	)
}

// A client joins groups on its connection while other requests of the same client check their
// group membership. Run with -race.
func TestStore_ClientMembershipIsReadWhileClientJoinsGroups(t *testing.T) {
	s := newStoreWithBroker("a")
	defer s.Close()

	req := kafkatest.NewRequest("client", 4, &heartbeat.Request{GroupId: "group-0"})
	client := kafka.ClientFromContext(req.Context)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			client.AddGroup(fmt.Sprintf("group-%d", i), "member")
			_ = client.GetOrCreateMemberId(fmt.Sprintf("other-%d", i))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			s.ServeMessage(kafkatest.NewRecorder(), req)
		}
	}()
	wg.Wait()
}

// A produce request writes to the topic while a config reload replaces the topic's channel. Run with -race.
func TestStore_ProduceWhileConfigIsReloaded(t *testing.T) {
	s := newStoreFromConfig(brokerAndTopicConfig("127.0.0.1:9092"))
	defer s.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			s.Update(brokerAndTopicConfig("127.0.0.1:9092"))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			s.ServeMessage(kafkatest.NewRecorder(), kafkatest.NewRequest("client", 3, &produce.Request{
				Topics: []produce.RequestTopic{{Name: "foo", Partitions: []produce.RequestPartition{{
					Record: kafka.RecordBatch{Records: []*kafka.Record{
						{Key: kafka.NewBytes([]byte("k")), Value: kafka.NewBytes([]byte("v"))},
					}},
				}}}},
			}))
		}
	}()
	wg.Wait()
}

// A metadata request reads the store's config while a config reload replaces it. Run with -race.
func TestStore_MetadataWhileConfigIsReloaded(t *testing.T) {
	s := newStoreFromConfig(brokerAndTopicConfig("127.0.0.1:9092"))
	defer s.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			s.Update(brokerAndTopicConfig("127.0.0.1:9092"))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			s.ServeMessage(kafkatest.NewRecorder(), kafkatest.NewRequest("client", 4, &metaData.Request{}))
		}
	}()
	wg.Wait()
}

// A reload changes the host of an existing broker while metadata requests report the broker's host.
// Run with -race.
func TestStore_MetadataWhileBrokerHostChanges(t *testing.T) {
	s := newStoreFromConfig(brokerAndTopicConfig("127.0.0.1:9092"))
	defer s.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			host := "localhost:9092"
			if i%2 == 0 {
				host = "foo.local:9092"
			}
			s.Update(brokerAndTopicConfig(host))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			s.ServeMessage(kafkatest.NewRecorder(), kafkatest.NewRequest("client", 4, &metaData.Request{}))
		}
	}()
	wg.Wait()
}

// Producers are registered while other producers write idempotent records. Run with -race.
func TestStore_InitProducerIdWhileProducerWrites(t *testing.T) {
	s := newStoreFromConfig(brokerAndTopicConfig("127.0.0.1:9092"))
	defer s.Close()
	partition := s.Topic("foo").Partition(0)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			s.ServeMessage(kafkatest.NewRecorder(), kafkatest.NewRequest("client", 3, &initProducerId.Request{}))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_, _ = partition.Write(kafka.RecordBatch{Records: []*kafka.Record{
				{
					Key:            kafka.NewBytes([]byte("k")),
					Value:          kafka.NewBytes([]byte("v")),
					ProducerId:     1,
					SequenceNumber: int32(i),
				},
			}})
		}
	}()
	wg.Wait()
}

func newStoreFromConfig(c *asyncapi3.Config) *store.Store {
	return store.New(c, enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewKafka())
}
