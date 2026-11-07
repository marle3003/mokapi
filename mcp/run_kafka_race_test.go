package mcp_test

import (
	"context"
	"io"
	"mokapi/kafka"
	"mokapi/mcp"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime/runtimetest"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// An AI client consumes records while a mock script keeps producing into the same topic. Run with -race.
func TestService_Run_Kafka_ConsumeWhileProducing(t *testing.T) {
	logrus.SetOutput(io.Discard)

	app := runtimetest.NewKafkaApp(
		asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("foo", "", ""),
			asyncapi3test.WithServer("broker", "kafka", "127.0.0.1:9092"),
			asyncapi3test.WithChannel("channel-1"),
		),
	)
	partition := app.Kafka.Get("foo").Store.Topic("channel-1").Partition(0)
	s := mcp.NewService(app)
	r, err := s.GetRunResponse(context.Background(), mcp.RunInput{
		Code: `mokapi.getApi('foo').getTopic('channel-1')`,
	})
	require.NoError(t, err)
	topic := r.Result.(mcp.Topic)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 300; i++ {
			_, err := partition.Write(kafka.RecordBatch{Records: []*kafka.Record{
				{Key: kafka.NewBytes([]byte("k")), Value: kafka.NewBytes([]byte("v"))},
			}})
			require.NoError(t, err)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 300; i++ {
			_, err := topic.Consume(0, 0, 1000)
			require.NoError(t, err)
		}
	}()
	wg.Wait()
}
