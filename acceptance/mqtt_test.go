package acceptance

import (
	"mokapi/config/static"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/try"
	"os"
	"path"
	"time"

	"github.com/stretchr/testify/require"
)

type MqttSuite struct{ BaseSuite }

func (suite *MqttSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	wd, err := os.Getwd()
	require.NoError(suite.T(), err)
	cfg.ConfigFile = path.Join(wd, "mokapi.yaml")
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./mqtt"}}
	cfg.Api.Search.Enabled = true
	suite.initCmd(cfg)
}

func (suite *MqttSuite) TestScriptPublish() {
	time.Sleep(2 * time.Second)
	t := suite.T()

	m := suite.cmd.App.Mqtt.Get("Sensor Service")
	topic, _ := m.Topic("sensors/temperature")
	require.NotNil(t, topic.Retained)
	require.Equal(t, `{"sensorId":"s123","value":12.3}`, string(topic.Retained.Data))
}

func (suite *MqttSuite) TestPublish() {
	time.Sleep(2 * time.Second)
	t := suite.T()

	c := mqtttest.NewClient("localhost:8883")
	defer c.Close()

	_, err := c.Send(&mqtt.Message{
		Header: &mqtt.Header{
			Type: mqtt.CONNECT,
		},
		Payload: &mqtt.ConnectRequest{
			Protocol:     "MQTT",
			Version:      5,
			CleanSession: true,
			KeepAlive:    60,
			ClientId:     "client-foo",
		},
	})
	require.NoError(suite.T(), err)

	p := &mqtt.Message{
		Header: &mqtt.Header{
			Type: mqtt.PUBLISH,
			QoS:  1,
		},
		Payload: &mqtt.PublishRequest{
			Topic:     "sensors/temperature",
			MessageId: 10,
			Data:      []byte(`{ "sensorId":"s123", "value":10 }`),
		},
	}
	res, err := c.Send(p)

	require.NoError(t, err)
	require.Equal(t, mqtt.PUBACK, res.Header.Type)
	msg := res.Payload.(*mqtt.PublishResponse)
	require.Equal(t, uint16(10), msg.MessageId)
	require.Equal(t, mqtt.PublishSuccess, msg.ReasonCode)
}
