package acceptance

import (
	"fmt"
	"mokapi/config/static"
	"mokapi/kafka"
	"mokapi/kafka/fetch"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/metaData"
	"mokapi/kafka/produce"
	"mokapi/schema/json/generator"
	"mokapi/try"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type KafkaSuite struct{ BaseSuite }

func (suite *KafkaSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Health.Port = cfg.Api.Port
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./kafka"}}
	cfg.Api.Search.Enabled = true
	cfg.Api.Search.InMemory = true
	cfg.Api.Search.NumIndexWorker = 1
	suite.initCmd(cfg)
}

func (suite *KafkaSuite) SetupTest() {
	generator.Seed(11)
}

func (suite *KafkaSuite) TestApi() {
	suite.T().Run("get AsyncAPI service", func(t *testing.T) {
		expected := map[string]any{
			"version": "1.0.0",
			"name":    "A sample AsyncApi Kafka streaming api",
			"servers": []any{
				map[string]any{
					"host":     "127.0.0.1:19092",
					"name":     "broker",
					"protocol": "kafka",
				},
			},

			"topics": []any{map[string]any{
				"name": "petstore.order-event",
				"metrics": map[string]any{
					// skip timestamp check "kafka_message_timestamp"
					"kafka_messages_total": float64(1),
				},
			}},
		}

		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			try.GetRequest(t, fmt.Sprintf("http://127.0.0.1:%v/api/services/kafka/A%%20sample%%20AsyncApi%%20Kafka%%20streaming%%20api", suite.cfg.Api.Port),
				nil,
				try.HasStatusCode(http.StatusOK),
				try.BodyContainsData(expected),
			)
			if !t.Failed() {
				break
			}
			if time.Now().After(deadline) {
				t.FailNow()
			}
			// reset test failure state before retrying
			t.Cleanup(func() {})
			time.Sleep(100 * time.Millisecond)
		}
	})
}

func (suite *KafkaSuite) TestKafka_TopicConfig() {
	c := kafkatest.NewClient("127.0.0.1:19092", "test")
	defer c.Close()

	r, err := c.Metadata(0, &metaData.Request{})
	require.NoError(suite.T(), err)
	require.Len(suite.T(), r.Topics, 1)
	require.Equal(suite.T(), "petstore.order-event", r.Topics[0].Name)
	require.Len(suite.T(), r.Topics[0].Partitions, 2)

	require.Equal(suite.T(), 2, suite.cmd.App.Kafka.Len())
}

func (suite *KafkaSuite) TestKafka_Produce_InvalidFormat() {
	c := kafkatest.NewClient("127.0.0.1:19092", "test")
	defer c.Close()

	r, err := c.Produce(0, &produce.Request{Topics: []produce.RequestTopic{
		{Name: "petstore.order-event", Partitions: []produce.RequestPartition{
			{
				Index: 0,
				Record: kafka.RecordBatch{
					Records: []*kafka.Record{
						{
							Offset:  0,
							Time:    time.Now(),
							Key:     kafka.NewBytes([]byte(`foo`)),
							Value:   kafka.NewBytes([]byte(`{}`)),
							Headers: nil,
						},
					},
				},
			},
		},
		}},
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "petstore.order-event", r.Topics[0].Name)
	require.Equal(suite.T(), kafka.InvalidRecord, r.Topics[0].Partitions[0].ErrorCode)
	require.Equal(suite.T(), int64(0), r.Topics[0].Partitions[0].BaseOffset)
}

func (suite *KafkaSuite) TestKafkaProduce() {
	c := kafkatest.NewClient("127.0.0.1:19092", "test")
	defer c.Close()
	r, err := c.Produce(0, &produce.Request{Topics: []produce.RequestTopic{
		{Name: "petstore.order-event", Partitions: []produce.RequestPartition{
			{
				Index: 0,
				Record: kafka.RecordBatch{
					Records: []*kafka.Record{
						{
							Offset:  0,
							Time:    time.Now(),
							Key:     kafka.NewBytes([]byte(`foo`)),
							Value:   kafka.NewBytes([]byte(`{"id": 12345}`)),
							Headers: nil,
						},
					},
				},
			},
		},
		}},
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "petstore.order-event", r.Topics[0].Name)
	require.Equal(suite.T(), kafka.None, r.Topics[0].Partitions[0].ErrorCode)
}

func (suite *KafkaSuite) TestKafkaEventAndMetrics() {
	// ensure scripts are executed
	time.Sleep(3 * time.Second)

	// test kafka metrics
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%d/api/metrics/kafka", suite.cfg.Api.Port), nil,
		try.BodyContains(`kafka_messages_total{service=\"A sample AsyncApi Kafka streaming api\",topic=\"petstore.order-event\"}","value":1}`),
	)

	// test kafka events, header added by JavaScript event handler
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%d/api/events?namespace=kafka", suite.cfg.Api.Port), nil,
		try.BodyContains(`"headers":{"foo":{"value":"bar","binary":"YmFy"}`),
		try.BodyContains(`"messageId":"order"`),
	)
}

func (suite *KafkaSuite) TestKafka3_Consume() {
	// ensure scripts are executed
	time.Sleep(3 * time.Second)

	c := kafkatest.NewClient("localhost:19093", "test")
	defer c.Close()

	r, err := c.Fetch(12, &fetch.Request{
		MaxBytes:  1000,
		MinBytes:  1,
		MaxWaitMs: 5000,
		Topics: []fetch.Topic{
			{
				Name: "petstore.order-event",
				Partitions: []fetch.RequestPartition{{
					Index:    0,
					MaxBytes: 1000,
				}},
			},
		},
	})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	require.Len(suite.T(), r.Topics[0].Partitions[0].RecordSet.Records, 1)
}
