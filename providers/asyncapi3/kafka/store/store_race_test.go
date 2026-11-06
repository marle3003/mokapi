package store_test

import (
	"fmt"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/metaData"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"sync"
	"testing"
	"time"
)

func brokerConfig(brokerName string) *asyncapi3.Config {
	return asyncapi3test.NewConfig(
		asyncapi3test.WithInfo("test", "", ""),
		asyncapi3test.WithServer(brokerName, "kafka", "127.0.0.1:9092"),
	)
}

func newStoreWithBroker(brokerName string) *store.Store {
	return store.New(brokerConfig(brokerName), enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewKafka())
}

// Clients connect through ServeMessage while the dashboard lists them. Run with -race.
func TestStore_ClientsAreReadWhileClientsConnect(t *testing.T) {
	s := newStoreWithBroker("a")
	defer s.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			s.ServeMessage(kafkatest.NewRecorder(), kafkatest.NewRequest(fmt.Sprintf("client-%d", i), 4, &metaData.Request{}))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = s.Clients()
			_, _ = s.Client(fmt.Sprintf("client-%d", i))
		}
	}()
	wg.Wait()
}

// Request handlers look up their broker while Update replaces the brokers. Run with -race.
func TestStore_BrokerIsLookedUpWhileBrokersChange(t *testing.T) {
	s := newStoreWithBroker("a")
	defer s.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			// a new name removes the old broker and adds a new one
			s.Update(brokerConfig(fmt.Sprintf("broker-%d", i)))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = s.BrokerByPort("127.0.0.1:9092")
		}
	}()
	wg.Wait()
}

// A script handler runs while the partition is locked and may look up topics through the store
// (mokapi/kafka produce). Update must therefore never wait for a partition lock while it holds
// the store lock, otherwise handler and Update block each other.
func TestStore_UpdateDoesNotWaitForPartitionWriteInHandler(t *testing.T) {
	config := func() *asyncapi3.Config {
		return asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("test", "", ""),
			asyncapi3test.WithServer("foo", "kafka", "127.0.0.1:9092"),
			asyncapi3test.WithChannel("foo"))
	}

	handlerEntered := make(chan struct{})
	releaseHandler := make(chan struct{})
	var s *store.Store
	emitter := enginetest.KafkaEventHandlerFunc(func(record *common.KafkaEventRecord) []*common.Action {
		close(handlerEntered)
		<-releaseHandler
		_ = s.Topic("foo")
		return nil
	})
	s = store.New(config(), emitter, &eventstest.Handler{}, monitor.NewKafka())

	writeDone := make(chan struct{})
	go func() {
		defer close(writeDone)
		_, _ = s.Topic("foo").Partition(0).Write(kafka.RecordBatch{Records: []*kafka.Record{
			{Key: kafka.NewBytes([]byte("k")), Value: kafka.NewBytes([]byte("v"))},
		}})
	}()
	<-handlerEntered

	updateDone := make(chan struct{})
	go func() {
		defer close(updateDone)
		s.Update(config())
	}()

	select {
	case <-updateDone:
	case <-time.After(5 * time.Second):
		// the handler's store lookup would now wait behind Update as well: a real deadlock,
		// so the goroutines are left behind
		t.Fatal("Update is blocked by a partition write that is waiting in a handler")
	}

	close(releaseHandler)
	select {
	case <-writeDone:
	case <-time.After(5 * time.Second):
		t.Fatal("partition write did not finish")
	}
	s.Close()
}
