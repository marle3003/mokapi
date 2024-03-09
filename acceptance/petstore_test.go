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
	cfg.Providers.File.Directory = "./petstore"
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

	suite.T().Run("get service", func(t *testing.T) {
		try.GetRequest(t, fmt.Sprintf("http://127.0.0.1:%v/api/services/http/Swagger%%20Petstore", suite.cfg.Api.Port),
			nil,
			try.HasStatusCode(http.StatusOK),
			try.BodyContains(`{"name":"Swagger Petstore","description":"This is a sample server Petstore server.  You can find out more about `),
			try.BodyMatch(`"configs":\[{"id":".*","url":".*\/acceptance\/petstore\/openapi\.yml","provider":"file","time":".*"}\]`),
		)
	})
}

func (suite *PetStoreSuite) TestJsHttpHandler() {
	// ensure scripts are executed
	time.Sleep(2 * time.Second)
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/2",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusNotFound),
		try.HasBody(""))

	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/3",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusNotFound),
		try.HasBody(""))

	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/4",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusInternalServerError),
		try.HasBody("marshal data to 'application/json' failed: does not match schema type=object properties=[id, category, name, photoUrls, tags, status] required=[name photoUrls]: missing required field 'name'\n"))

	// use generated data but change pet's name
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/5",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"id":4365710250675650087,"category":{"id":3349113537306089850,"name":"water buffalo"},"name":"Zoe","photoUrls":[],"tags":[{"id":8485680975708586437,"name":"yKMEz"},{"id":5271418922154914286,"name":"JQInuLnfGNcRsE"},{"id":4318152911919099031,"name":"BZBBc Q"}],"status":"pending"}`))

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
		try.HasBody(`{"id":843730692693298266,"petId":7424164296119123377,"quantity":-652938557,"shipDate":"1989-01-30T07:58:01Z","status":"approved","complete":false}`))

	try.GetRequest(suite.T(), "https://localhost:18443/store/order/10",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"id":6118637534854712545,"petId":5549848391338120895,"quantity":-924758850,"shipDate":"1989-11-19T16:57:16Z","status":"approved","complete":true}`))
}

func (suite *PetStoreSuite) TestKafka_TopicConfig() {
	c := kafkatest.NewClient("127.0.0.1:19092", "test")
	defer c.Close()

	r, err := c.Metadata(0, &metaData.Request{})
	require.NoError(suite.T(), err)
	require.Len(suite.T(), r.Topics, 1)
	require.Equal(suite.T(), "petstore.order-event", r.Topics[0].Name)
	require.Len(suite.T(), r.Topics[0].Partitions, 2)

	require.Len(suite.T(), suite.cmd.App.Http, 1)
}

func (suite *PetStoreSuite) TestKafka_Produce_InvalidFormat() {
	c := kafkatest.NewClient("127.0.0.1:19092", "test")
	defer c.Close()

	r, err := c.Produce(0, &produce.Request{Topics: []produce.RequestTopic{
		{Name: "petstore.order-event", Partitions: []produce.RequestPartition{
			{
				Index: 0,
				Record: kafka.RecordBatch{
					Records: []kafka.Record{
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
	require.Equal(suite.T(), kafka.CorruptMessage, r.Topics[0].Partitions[0].ErrorCode)
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
					Records: []kafka.Record{
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
	// test kafka metrics
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%s/api/metrics/kafka", suite.cfg.Api.Port), nil,
		try.BodyContains(`kafka_messages_total{service=\"A sample AsyncApi Kafka streaming api\",topic=\"petstore.order-event\"}","value":1}`),
	)

	// test kafka events, header added by JavaScript event handler
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%s/api/events?namespace=kafka", suite.cfg.Api.Port), nil,
		try.BodyContains(`"headers":{"foo":"bar"}`),
	)
}
