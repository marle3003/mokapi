package acceptance

import (
	"encoding/json"
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type PetStoreSuite struct{ BaseSuite }

func (suite *PetStoreSuite) SetupSuite() {
	cfg := static.NewConfig()
	port := try.GetFreePort()
	cfg.Api.Port = fmt.Sprintf("%v", port)
	cfg.Providers.File.Directories = []string{"./petstore"}
	cfg.Api.Search.Enabled = true
	suite.initCmd(cfg)
}

func (suite *PetStoreSuite) SetupTest() {
	generator.Seed(11)
}

func (suite *PetStoreSuite) TestApi() {
	suite.T().Run("CORS", func(t *testing.T) {
		try.GetRequest(t, fmt.Sprintf("http://127.0.0.1:%v", suite.cfg.Api.Port),
			nil,
			try.HasStatusCode(http.StatusOK),
			try.HasHeader("Access-Control-Allow-Origin", "*"))
	})

	suite.T().Run("get Swagger HTTP service", func(t *testing.T) {
		try.GetRequest(t, fmt.Sprintf("http://127.0.0.1:%v/api/services/http/Swagger%%20Petstore", suite.cfg.Api.Port),
			nil,
			try.HasStatusCode(http.StatusOK),
			try.BodyContains(`{"name":"Swagger Petstore","description":"This is a sample server Petstore server.  You can find out more about `),
			try.BodyMatch(`"configs":\[{"id":".*","url":".*\/acceptance\/petstore\/openapi\.yml","provider":"file","time":".*"}\]`),
		)
	})

	suite.T().Run("get AsyncAPI service", func(t *testing.T) {
		expected := map[string]interface{}{
			"version":     "1.0.0",
			"name":        "A sample AsyncApi Kafka streaming api",
			"description": "",
			"servers": []interface{}{
				map[string]interface{}{
					"description": "",
					"host":        "127.0.0.1:19092",
					"name":        "broker",
					"protocol":    "kafka",
				},
			},

			"topics": []interface{}{map[string]interface{}{
				"bindings":    map[string]interface{}{"partitions": float64(2), "segmentMs": float64(30000), "valueSchemaValidation": true},
				"description": "",
				"messages": map[string]interface{}{
					"order": map[string]interface{}{
						"name":        "order",
						"contentType": "application/json",
						"header": map[string]interface{}{
							"schema": map[string]interface{}{
								"properties": map[string]interface{}{
									"number": map[string]interface{}{
										"type": "number",
									}, "test": map[string]interface{}{
										"type": "string",
									},
								}, "type": "object",
							},
						},
						"key": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "string",
							},
						},
						"payload": map[string]interface{}{
							"schema": map[string]interface{}{
								"properties": map[string]interface{}{
									"accepted": map[string]interface{}{
										"properties": map[string]interface{}{
											"timestamp": map[string]interface{}{"format": "date-time", "type": "string"},
										},
										"type": "object",
									},
									"completed": map[string]interface{}{
										"properties": map[string]interface{}{
											"timestamp": map[string]interface{}{"format": "date-time", "type": "string"},
										},
										"type": "object",
									},
									"id": map[string]interface{}{"type": "integer"},
									"placed": map[string]interface{}{
										"properties": map[string]interface{}{
											"petid":     map[string]interface{}{"type": "integer"},
											"quantity":  map[string]interface{}{"format": "int32", "type": "integer"},
											"ship-date": map[string]interface{}{"format": "date-time", "type": "string"},
										},
										"type": "object",
									},
								},
								"required": []interface{}{"id"},
								"type":     "object",
							},
						},
					},
				},
				"name": "petstore.order-event",
				"partitions": []interface{}{
					map[string]interface{}{"id": float64(0), "leader": map[string]interface{}{"addr": "127.0.0.1:19092", "name": "broker"}, "offset": float64(1), "segments": float64(1), "startOffset": float64(0)},
					map[string]interface{}{"id": float64(1), "leader": map[string]interface{}{"addr": "127.0.0.1:19092", "name": "broker"}, "offset": float64(0), "segments": float64(0), "startOffset": float64(0)},
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

func (suite *PetStoreSuite) TestJsHttpHandler() {
	// ensure scripts are executed
	time.Sleep(4 * time.Second)
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/2",
		map[string]string{"Accept": "application/json", "api_key": "123"},
		try.HasStatusCode(http.StatusNotFound),
		try.HasBody(""))

	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/3",
		map[string]string{"Accept": "application/json", "api_key": "123"},
		try.HasStatusCode(http.StatusNotFound),
		try.HasBody(""))

	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/4",
		map[string]string{"Accept": "application/json", "api_key": "123"},
		try.HasStatusCode(http.StatusInternalServerError),
		try.HasBody("encoding data to 'application/json' failed: error count 1:\n\t- #/required: required properties are missing: name, photoUrls\n"))

	// use generated data but change pet's name
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/5",
		map[string]string{"Accept": "application/json", "api_key": "123"},
		try.HasStatusCode(http.StatusOK),
		try.BodyContains(`"name":"Zoe"`))

	// test http metrics
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%s/api/metrics/http?path=/pet/{petId}", suite.cfg.Api.Port), nil,
		try.BodyContains(`http_requests_total{service=\"Swagger Petstore\",endpoint=\"/pet/{petId}\"}","value":4}`),
	)
}

func (suite *PetStoreSuite) TestLuaFile() {
	// ensure scripts are executed
	time.Sleep(2 * time.Second)
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/findByStatus?status=available&status=pending",
		map[string]string{"Accept": "application/json", "Authorization": "foo"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody("[{\"name\":\"Gidget\",\"photoUrls\":[\"http://www.pets.com/gidget.png\"],\"status\":\"pending\"},{\"name\":\"Max\",\"photoUrls\":[\"http://www.pets.com/max.png\"],\"status\":\"available\"}]"))
}

func (suite *PetStoreSuite) TestGetOrderById() {
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/store/order/1",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"petId":23377,"quantity":92,"shipDate":"2012-01-30T07:58:01Z","complete":false}`))

	try.GetRequest(suite.T(), "https://localhost:18443/store/order/10",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		// properties like id or petId are optional
		try.HasBody(`{"id":93761,"petId":83318,"quantity":27,"shipDate":"2014-02-04T10:00:17Z","status":"placed","complete":true}`))
}

func (suite *PetStoreSuite) TestTls() {
	try.GetRequest(suite.T(), "https://localhost:18443/store/order/10",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.IsTls("localhost"),
	)
}

func (suite *PetStoreSuite) TestKafka_TopicConfig() {
	c := kafkatest.NewClient("127.0.0.1:19092", "test")
	defer c.Close()

	r, err := c.Metadata(0, &metaData.Request{})
	require.NoError(suite.T(), err)
	require.Len(suite.T(), r.Topics, 1)
	require.Equal(suite.T(), "petstore.order-event", r.Topics[0].Name)
	require.Len(suite.T(), r.Topics[0].Partitions, 2)

	require.Len(suite.T(), suite.cmd.App.ListHttp(), 1)
}

func (suite *PetStoreSuite) TestKafka_Produce_InvalidFormat() {
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

func (suite *PetStoreSuite) TestKafkaProduce() {
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

func (suite *PetStoreSuite) TestEvents() {
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/user/bob",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK))

	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/api/events?namespace=http&name=Swagger%%20Petstore&path=/user/{username}", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.BodyContains(`"url":"http://127.0.0.1:18080/user/bob"`))

	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/api/search/query?q=type:event%%20event.traits.namespace=http", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var data map[string]any
			err := json.Unmarshal([]byte(body), &data)
			require.NoError(t, err)
			results := data["results"].([]any)
			evt := results[0].(map[string]any)
			require.Equal(t, "Event", evt["type"])
			require.Equal(t, "GET http://127.0.0.1:18080/user/bob", evt["title"])
			require.Equal(t, "Swagger Petstore", evt["domain"])
		}),
	)
}

func (suite *PetStoreSuite) TestKafkaEventAndMetrics() {
	// ensure scripts are executed
	time.Sleep(3 * time.Second)

	// test kafka metrics
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%s/api/metrics/kafka", suite.cfg.Api.Port), nil,
		try.BodyContains(`kafka_messages_total{service=\"A sample AsyncApi Kafka streaming api\",topic=\"petstore.order-event\"}","value":1}`),
	)

	// test kafka events, header added by JavaScript event handler
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%s/api/events?namespace=kafka", suite.cfg.Api.Port), nil,
		try.BodyContains(`"headers":{"foo":{"value":"bar","binary":"YmFy"}`),
		try.BodyContains(`"messageId":"order"`),
	)
}

func (suite *PetStoreSuite) TestKafka3_Consume() {
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

func (suite *PetStoreSuite) TestSearch_Paging() {
	time.Sleep(3 * time.Second)

	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/api/search/query?q=api:%%22Swagger%%20Petstore%%22%%20type:http", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var data map[string]any
			err := json.Unmarshal([]byte(body), &data)
			assert.NoError(t, err)
			assert.NotNil(t, data)

			assert.Equal(t, float64(35), data["total"])

			items := data["results"].([]any)
			assert.Len(t, items, 10)
			evt := items[0].(map[string]interface{})
			assert.Equal(t, "HTTP", evt["type"])
			assert.Equal(t, "Swagger Petstore", evt["title"])
			assert.NotContains(t, evt, "domain")
		}),
	)

	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/api/search/query?q=api:%%22Swagger%%20Petstore%%22%%20type:http&index=1", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var data map[string]any
			err := json.Unmarshal([]byte(body), &data)
			assert.NoError(t, err)
			assert.NotNil(t, data)

			assert.Equal(t, float64(35), data["total"])

			items := data["results"].([]any)
			assert.Len(t, items, 10)
			evt := items[0].(map[string]interface{})
			assert.Equal(t, "HTTP", evt["type"])
			assert.Equal(t, "GET /pet/{petId}", evt["title"])
			assert.Equal(t, "Swagger Petstore", evt["domain"])
		}),
	)
}
