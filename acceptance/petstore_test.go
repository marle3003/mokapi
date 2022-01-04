package acceptance

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/metaData"
	"mokapi/kafka/protocol/produce"
	"mokapi/server/web/webtest"
	"time"
)

type PetStoreSuite struct{ BaseSuite }

func (suite *PetStoreSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Providers.File.Directory = "./petstore"
	suite.initCmd(cfg)
}

func (suite *PetStoreSuite) SetupTest() {
	gofakeit.Seed(11)
}

func (suite *PetStoreSuite) TestJsFile() {
	// ensure scripts are executed
	time.Sleep(2 * time.Second)
	err := webtest.GetRequest("http://127.0.0.1:8080/pet/2",
		map[string]string{"Accept": "application/json"},
		webtest.HasStatusCode(suite.T(), 404),
		webtest.HasBody(suite.T(),
			`{"id":-8379641344161477543,"category":{"id":-1799207740735652432,"name":"RMaRxHkiJBPtapW"},"name":"doggie","photoUrls":[],"tags":[{"id":-3430133205295092491,"name":"nSMKgtlxwnqhqcl"},{"id":-4360704630090834069,"name":"YkWwfoRLOPxLIok"},{"id":-9084870506124948944,"name":"qanPAKaXSMQFpZy"}],"status":"pending"}`,
		))
	require.NoError(suite.T(), err)

	err = webtest.GetRequest("http://127.0.0.1:8080/pet/3",
		map[string]string{"Accept": "application/json"},
		webtest.HasStatusCode(suite.T(), 404),
		webtest.HasBody(suite.T(), ""))
	require.NoError(suite.T(), err)
}

func (suite *PetStoreSuite) TestLuaFile() {
	// ensure scripts are executed
	time.Sleep(time.Second)
	err := webtest.GetRequest("http://127.0.0.1:8080/pet/findByStatus?status=available&status=pending",
		map[string]string{"Accept": "application/json"},
		webtest.HasStatusCode(suite.T(), 200),
		webtest.HasBody(suite.T(),
			"[{\"name\":\"Gidget\",\"photoUrls\":[\"http://www.pets.com/gidget.png\"],\"status\":\"pending\"},{\"name\":\"Max\",\"photoUrls\":[\"http://www.pets.com/max.png\"],\"status\":\"available\"}]"))
	require.NoError(suite.T(), err)
}

func (suite *PetStoreSuite) TestGetPetById() {
	err := webtest.GetRequest("http://127.0.0.1:8080/pet/1",
		map[string]string{"Accept": "application/json"},
		webtest.HasStatusCode(suite.T(), 200),
		webtest.HasBody(suite.T(),
			`{"id":-8379641344161477543,"category":{"id":-1799207740735652432,"name":"RMaRxHkiJBPtapW"},"name":"doggie","photoUrls":[],"tags":[{"id":-3430133205295092491,"name":"nSMKgtlxwnqhqcl"},{"id":-4360704630090834069,"name":"YkWwfoRLOPxLIok"},{"id":-9084870506124948944,"name":"qanPAKaXSMQFpZy"}],"status":"pending"}`,
		))
	require.NoError(suite.T(), err)

	err = webtest.GetRequest("https://localhost:8443/pet/1",
		map[string]string{"Accept": "application/json"},
		webtest.HasStatusCode(suite.T(), 200),
		webtest.HasBody(suite.T(),
			`{"id":-5233707484353581840,"category":{"id":-7922211254674255348,"name":"HGyyvqqdHueUxcv"},"name":"doggie","photoUrls":[],"tags":[{"id":-885632864726843768,"name":"eDjRRGUnsAxdBXG"}],"status":"pending"}`,
		))
	require.NoError(suite.T(), err)
}

func (suite *PetStoreSuite) TestKafka_TopicConfig() {
	c := kafkatest.NewClient("127.0.0.1:9092", "test")
	defer c.Close()

	r, err := c.Metadata(0, &metaData.Request{})
	require.NoError(suite.T(), err)
	require.Len(suite.T(), r.Topics, 1)
	require.Equal(suite.T(), "petstore.order-event", r.Topics[0].Name)
	require.Len(suite.T(), r.Topics[0].Partitions, 2)
}

func (suite *PetStoreSuite) TestKafka_Produce_InvalidFormat() {
	c := kafkatest.NewClient("127.0.0.1:9092", "test")
	defer c.Close()

	r, err := c.Produce(0, &produce.Request{Topics: []produce.RequestTopic{
		{Name: "petstore.order-event", Partitions: []produce.RequestPartition{
			{
				Index: 0,
				Record: protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Offset:  0,
							Time:    time.Now(),
							Key:     protocol.NewBytes([]byte(`foo`)),
							Value:   protocol.NewBytes([]byte(`{}`)),
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
	require.Equal(suite.T(), protocol.CorruptMessage, r.Topics[0].Partitions[0].ErrorCode)
	require.Equal(suite.T(), int64(0), r.Topics[0].Partitions[0].BaseOffset)
}

func (suite *PetStoreSuite) KafkaProduce() {
	c := kafkatest.NewClient("127.0.0.1:9092", "test")
	defer c.Close()
	r, err := c.Produce(0, &produce.Request{Topics: []produce.RequestTopic{
		{Name: "petstore.order-event", Partitions: []produce.RequestPartition{
			{
				Index: 0,
				Record: protocol.RecordBatch{
					Records: []protocol.Record{
						{
							Offset:  0,
							Time:    time.Now(),
							Key:     protocol.NewBytes([]byte(`foo`)),
							Value:   protocol.NewBytes([]byte(`{"id": 12345}`)),
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
	require.Equal(suite.T(), protocol.None, r.Topics[0].Partitions[0].ErrorCode)
	require.Equal(suite.T(), int64(0), r.Topics[0].Partitions[0].BaseOffset)
}
