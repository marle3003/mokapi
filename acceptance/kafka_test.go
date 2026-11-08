package acceptance

import (
	"encoding/json"
	"fmt"
	"mokapi/api"
	"mokapi/config/static"
	"mokapi/kafka"
	"mokapi/kafka/fetch"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/metaData"
	"mokapi/kafka/produce"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/schema/json/generator"
	"mokapi/try"
	"net/http"
	"strings"
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
					"messages_total": float64(1),
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

	var eventId string
	// test kafka events, header added by JavaScript event handler
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%d/api/events?namespace=kafka", suite.cfg.Api.Port), nil,
		try.AssertBody(func(t *testing.T, body string) {
			var data api.EventPage
			err := json.Unmarshal([]byte(body), &data)
			require.NoError(t, err)
			require.Len(t, data.Items, 2)
			// first event does not define a header specification therefor header only contains binary
			event := data.Items[1]
			eventId = event.Id
		}),
	)

	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%d/api/events/"+eventId, suite.cfg.Api.Port), nil,
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

// KafkaMessageIdSuite verifies that a script selects the AsyncAPI message of a channel by messageId.
type KafkaMessageIdSuite struct{ BaseSuite }

func (suite *KafkaMessageIdSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./kafka-messageid"}}
	suite.initCmd(cfg)
	waitForScript(suite.T(), "http://localhost:18094")
}

func (suite *KafkaMessageIdSuite) partition() *store.Partition {
	return suite.cmd.App.Kafka.Get("Order Service").Store.Topic("orders").Partition(0)
}

func (suite *KafkaMessageIdSuite) post(body string, conditions ...try.ResponseCondition) {
	try.Request(suite.T(), http.MethodPost, "http://localhost:18094/sim/kafka",
		map[string]string{"Content-Type": "application/json"}, strings.NewReader(body), conditions...)
}

// simulate asks the fixture script to produce messages and returns the records written by it.
func (suite *KafkaMessageIdSuite) simulate(body string, conditions ...try.ResponseCondition) []*kafka.Record {
	start := suite.partition().Offset()
	suite.post(body, conditions...)

	c := kafkatest.NewClient("localhost:19094", "test")
	defer c.Close()
	r, err := c.Fetch(12, &fetch.Request{
		MaxBytes:  100000,
		MinBytes:  1,
		MaxWaitMs: 5000,
		Topics: []fetch.Topic{{
			Name:       "orders",
			Partitions: []fetch.RequestPartition{{Index: 0, FetchOffset: start, MaxBytes: 100000}},
		}},
	})
	require.NoError(suite.T(), err)
	return r.Topics[0].Partitions[0].RecordSet.Records
}

// TestGeneratesPayloadAndKeyFromSelectedMessage verifies that messageId alone produces a record
// whose value and key are generated from the selected message.
func (suite *KafkaMessageIdSuite) TestGeneratesPayloadAndKeyFromSelectedMessage() {
	records := suite.simulate(`{"messages":[{"messageId":"orderDeleted"}]}`, try.HasStatusCode(http.StatusNoContent))

	require.Len(suite.T(), records, 1)
	require.Regexp(suite.T(), `^deleted-[0-9]{3}$`, kafka.BytesToString(records[0].Key))
	var value map[string]any
	require.NoError(suite.T(), json.Unmarshal(kafka.Read(records[0].Value), &value))
	require.Equal(suite.T(), "deleted", value["event"])
	require.Contains(suite.T(), value, "reason")
}

// TestMixedMessagesInOneCall verifies that every message of one produce call is generated
// from its own messageId.
func (suite *KafkaMessageIdSuite) TestMixedMessagesInOneCall() {
	records := suite.simulate(`{"messages":[{"messageId":"orderCreated"},{"messageId":"orderDeleted"}]}`,
		try.HasStatusCode(http.StatusNoContent))

	require.Len(suite.T(), records, 2)
	require.Regexp(suite.T(), `^created-[0-9]{3}$`, kafka.BytesToString(records[0].Key))
	require.Contains(suite.T(), kafka.BytesToString(records[0].Value), `"event":"created"`)
	require.Regexp(suite.T(), `^deleted-[0-9]{3}$`, kafka.BytesToString(records[1].Key))
	require.Contains(suite.T(), kafka.BytesToString(records[1].Value), `"event":"deleted"`)
}

// TestTombstoneWithMessageId verifies that value null still produces a record without value
// when a messageId is given.
func (suite *KafkaMessageIdSuite) TestTombstoneWithMessageId() {
	records := suite.simulate(`{"messages":[{"messageId":"orderDeleted","key":"order-1","value":null}]}`,
		try.HasStatusCode(http.StatusNoContent))

	require.Len(suite.T(), records, 1)
	require.Equal(suite.T(), "order-1", kafka.BytesToString(records[0].Key))
	require.Nil(suite.T(), records[0].Value)
}

// TestTombstoneWithoutMessageId verifies that value null still produces a record without value
// when no messageId is given.
func (suite *KafkaMessageIdSuite) TestTombstoneWithoutMessageId() {
	records := suite.simulate(`{"messages":[{"key":"order-1","value":null}]}`, try.HasStatusCode(http.StatusNoContent))

	require.Len(suite.T(), records, 1)
	require.Equal(suite.T(), "order-1", kafka.BytesToString(records[0].Key))
	require.Nil(suite.T(), records[0].Value)
}

// TestInvalidDataForSelectedMessageFails verifies that data is validated against the selected
// message only, even if another message of the topic would accept it.
func (suite *KafkaMessageIdSuite) TestInvalidDataForSelectedMessageFails() {
	start := suite.partition().Offset()

	suite.post(`{"messages":[{"messageId":"orderDeleted","data":{"event":"created","orderId":1}}]}`,
		try.HasStatusCode(http.StatusUnprocessableEntity),
		try.BodyContains("failed to produce message to Kafka topic 'orders'"))

	require.Equal(suite.T(), start, suite.partition().Offset())
}

// TestUnknownMessageIdFails verifies that a typo in messageId is reported to the script
// instead of falling back to a random message.
func (suite *KafkaMessageIdSuite) TestUnknownMessageIdFails() {
	start := suite.partition().Offset()

	suite.post(`{"messages":[{"messageId":"orderCreatd"}]}`,
		try.HasStatusCode(http.StatusUnprocessableEntity),
		try.BodyContains("message 'orderCreatd' not found in channel 'orders'"))

	require.Equal(suite.T(), start, suite.partition().Offset())
}
