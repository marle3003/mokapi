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
	port, err := try.GetFreePort()
	require.NoError(suite.T(), err)
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

	e := events.GetEvents(events.NewTraits().WithNamespace("kafka"))
	require.Len(suite.T(), e, 1)
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
		try.HasBody(`{"id":-8379641344161477543,"category":{"id":-1799207740735652432,"name":"RMaRxHkiJBPtapW"},"name":"doggie","photoUrls":[],"tags":[{"id":-3430133205295092491,"name":"nSMKgtlxwnqhqcl"},{"id":-4360704630090834069,"name":"YkWwfoRLOPxLIok"},{"id":-9084870506124948944,"name":"qanPAKaXSMQFpZy"}],"status":"pending"}`))

	try.GetRequest(suite.T(), "https://localhost:18443/pet/1",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"id":-5233707484353581840,"category":{"id":-7922211254674255348,"name":"HGyyvqqdHueUxcv"},"name":"doggie","photoUrls":[],"tags":[{"id":-885632864726843768,"name":"eDjRRGUnsAxdBXG"}],"status":"pending"}`))
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

func (suite *PetStoreSuite) KafkaProduce() {
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
	require.Equal(suite.T(), int64(0), r.Topics[0].Partitions[0].BaseOffset)
}
