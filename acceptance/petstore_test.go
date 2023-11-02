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
	"mokapi/runtime/events"
	"mokapi/try"
	"net/http"
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
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.HasHeader("Access-Control-Allow-Origin", "*"))
}

func (suite *PetStoreSuite) TestJsFile() {
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

	e := events.GetEvents(events.NewTraits().WithNamespace("http"))
	require.Len(suite.T(), e, 3)
}

func (suite *PetStoreSuite) TestLuaFile() {
	// ensure scripts are executed
	time.Sleep(2 * time.Second)
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/findByStatus?status=available&status=pending",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody("[{\"name\":\"Gidget\",\"photoUrls\":[\"http://www.pets.com/gidget.png\"],\"status\":\"pending\"},{\"name\":\"Max\",\"photoUrls\":[\"http://www.pets.com/max.png\"],\"status\":\"available\"}]"))
}

func (suite *PetStoreSuite) TestGetPetById() {
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/1",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"id":-8379641344161477543,"category":{"id":7424164296119123376,"name":"id1"},"name":"doggie","photoUrls":["OwQ;ezYvmtLRfv","ZL","evUwYR5rljgmr z"],"tags":[{"id":1502793126295339460,"name":"Wgsfi6SFflnzb"},{"id":-7391388163417809074,"name":"m"},{"id":1750077968446365139,"name":"DHXGaQPSE"},{"id":-4201786370340656298,"name":"KwoQEfHPR99"}],"status":"pending"}`))

	try.GetRequest(suite.T(), "https://localhost:18443/pet/5",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"id":2780694049194110144,"category":{"id":358698270060065978,"name":" qcXEuR"},"name":"doggie","photoUrls":["euX","cJQIn"],"tags":[{"id":-7925257062691635148,"name":"XLc.OqnSYDNeJn"},{"id":-8530491859977087890,"name":"Q2sq4zDyvB0Q"},{"id":-3767257451481315098,"name":"M9J7SZsU"},{"id":2343312407715696586,"name":"e"},{"id":-8004558519261086467,"name":"V"}],"status":"available"}`))
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
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/1",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK))

	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/api/events?namespace=http&name=Swagger%%20Petstore", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.BodyContains("Swagger Petstore"))
}
