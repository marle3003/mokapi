package acceptance

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/metaData"
	"mokapi/kafka/produce"
	"mokapi/try"
	"net/http"
	"testing"
	"time"
)

type PetStoreSuite struct{ BaseSuite }

func (suite *PetStoreSuite) SetupSuite() {
	cfg := static.NewConfig()
	port := try.GetFreePort()
	cfg.Api.Port = fmt.Sprintf("%v", port)
	cfg.Providers.File.Directories = []string{"./petstore"}
	suite.initCmd(cfg)
}

func (suite *PetStoreSuite) SetupTest() {
	gofakeit.Seed(11)
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
		// ensure scripts are executed
		time.Sleep(5 * time.Second)

		expected := map[string]interface{}{
			"version":     "1.0.0",
			"name":        "A sample AsyncApi Kafka streaming api",
			"description": "",
			"servers": []interface{}{
				map[string]interface{}{
					"description": "",
					"host":        "127.0.0.1:19092",
					"name":        "broker",
				},
			},

			"topics": []interface{}{map[string]interface{}{
				"description": "",
				"messages": map[string]interface{}{
					"#/components/messages/order": map[string]interface{}{
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

		try.GetRequest(t, fmt.Sprintf("http://127.0.0.1:%v/api/services/kafka/A%%20sample%%20AsyncApi%%20Kafka%%20streaming%%20api", suite.cfg.Api.Port),
			nil,
			try.HasStatusCode(http.StatusOK),
			try.BodyContainsData(expected),
		)
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
		try.HasBody("encoding data to 'application/json' failed: found 1 error:\nrequired properties are missing: name, photoUrls\nschema path #/required\n"))

	// use generated data but change pet's name
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/5",
		map[string]string{"Accept": "application/json", "api_key": "123"},
		try.HasStatusCode(http.StatusOK),
		try.BodyContains(`},"name":"Zoe","photoUrls":`))

	// test http metrics
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%s/api/metrics/http?path=/pet/{petId}", suite.cfg.Api.Port), nil,
		try.BodyContains(`http_requests_total{service=\"Swagger Petstore\",endpoint=\"/pet/{petId}\"}","value":4}`),
	)
}

func (suite *PetStoreSuite) TestLuaFile() {
	// ensure scripts are executed
	time.Sleep(2 * time.Second)
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/findByStatus?status=available&status=pending",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody("[{\"name\":\"Gidget\",\"photoUrls\":[\"http://www.pets.com/gidget.png\"],\"status\":\"pending\"},{\"name\":\"Max\",\"photoUrls\":[\"http://www.pets.com/max.png\"],\"status\":\"available\"}]"))
}

func (suite *PetStoreSuite) TestGetOrderById() {
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/store/order/1",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"id":98266,"petId":23377,"quantity":92,"shipDate":"2012-01-30T07:58:01Z","status":"approved","complete":false}`))

	try.GetRequest(suite.T(), "https://localhost:18443/store/order/10",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"id":12545,"petId":20895,"quantity":16,"shipDate":"2027-11-26T16:57:16Z","status":"approved","complete":true}`))
}

func (suite *PetStoreSuite) TestKafka_TopicConfig() {
	c := kafkatest.NewClient("127.0.0.1:19092", "test")
	defer c.Close()

	r, err := c.Metadata(0, &metaData.Request{})
	require.NoError(suite.T(), err)
	require.Len(suite.T(), r.Topics, 1)
	require.Equal(suite.T(), "petstore.order-event", r.Topics[0].Name)
	require.Len(suite.T(), r.Topics[0].Partitions, 2)

	require.Len(suite.T(), suite.cmd.App.Http.List(), 1)
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
		try.BodyContains(`"messageId":"#/components/messages/order"`),
	)
}
