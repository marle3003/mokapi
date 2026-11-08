package acceptance

import (
	"encoding/json"
	"mokapi/config/static"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/try"
	"net/http"
	"os"
	"path"
	"strings"
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

func (suite *MqttSuite) TestHandlerCanPublishToAnotherTopic() {
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
			ClientId:     "client-forward",
		},
	})
	require.NoError(t, err)

	reading := `{"sensorId":"s-forward","value":42}`
	res, err := c.Send(&mqtt.Message{
		Header: &mqtt.Header{
			Type: mqtt.PUBLISH,
			QoS:  1,
		},
		Payload: &mqtt.PublishRequest{
			Topic:     "sensors/temperature",
			MessageId: 11,
			Data:      []byte(reading),
		},
	})

	require.NoError(t, err)
	require.Equal(t, mqtt.PUBACK, res.Header.Type)
	msg := res.Payload.(*mqtt.PublishResponse)
	require.Equal(t, uint16(11), msg.MessageId)
	require.Equal(t, mqtt.PublishSuccess, msg.ReasonCode)

	m := suite.cmd.App.Mqtt.Get("Sensor Service")
	topic, ok := m.Topic("sensors/alerts")
	require.True(t, ok)
	require.NotNil(t, topic.Retained)
	require.Equal(t, reading, string(topic.Retained.Data))
}

// simulatePublish asks the fixture script to publish a message to the topic and returns the
// retained payload of that topic.
func (suite *MqttSuite) simulatePublish(topic, message string, conditions ...try.ResponseCondition) []byte {
	waitForScript(suite.T(), "http://localhost:18095")

	body := `{"topic":"` + topic + `","message":` + message + `}`
	try.Request(suite.T(), http.MethodPost, "http://localhost:18095/sim/mqtt",
		map[string]string{"Content-Type": "application/json"}, strings.NewReader(body), conditions...)

	m := suite.cmd.App.Mqtt.Get("Sensor Service")
	t, ok := m.Topic(topic)
	require.True(suite.T(), ok)
	if t.Retained == nil {
		return nil
	}
	return t.Retained.Data
}

// TestPublishGeneratesPayloadOfSelectedMessage verifies that messageId alone publishes a payload
// generated from the selected message of the topic.
func (suite *MqttSuite) TestPublishGeneratesPayloadOfSelectedMessage() {
	data := suite.simulatePublish("sensors/events", `{"messageId":"alarm"}`, try.HasStatusCode(http.StatusNoContent))

	var payload map[string]any
	require.NoError(suite.T(), json.Unmarshal(data, &payload), string(data))
	require.Contains(suite.T(), []any{"warning", "critical"}, payload["level"])
	require.NotEmpty(suite.T(), payload["sensorId"])
}

// TestPublishDataIsEncodedWithSelectedMessage verifies that data valid for the selected message
// is published as its encoded payload.
func (suite *MqttSuite) TestPublishDataIsEncodedWithSelectedMessage() {
	data := suite.simulatePublish("sensors/events",
		`{"messageId":"reading","data":{"sensorId":"s1","value":21.5}}`, try.HasStatusCode(http.StatusNoContent))

	require.JSONEq(suite.T(), `{"sensorId":"s1","value":21.5}`, string(data))
}

// TestPublishInvalidDataForSelectedMessageFails verifies that data is validated against the
// selected message, even if another message of the topic would accept it.
func (suite *MqttSuite) TestPublishInvalidDataForSelectedMessageFails() {
	suite.simulatePublish("sensors/events",
		`{"messageId":"alarm","data":{"sensorId":"s1","value":21.5}}`,
		try.HasStatusCode(http.StatusUnprocessableEntity),
		try.BodyContains("required properties are missing: level"))
}

// TestPublishGeneratesPayloadWithoutMessageId verifies that a message without messageId, data and
// value gets a generated payload of one of the topic's messages.
func (suite *MqttSuite) TestPublishGeneratesPayloadWithoutMessageId() {
	data := suite.simulatePublish("sensors/events", `{}`, try.HasStatusCode(http.StatusNoContent))

	var payload map[string]any
	require.NoError(suite.T(), json.Unmarshal(data, &payload), string(data))
	require.NotEmpty(suite.T(), payload["sensorId"])
}

// TestPublishUnknownMessageIdFails verifies that a typo in messageId is reported to the script.
func (suite *MqttSuite) TestPublishUnknownMessageIdFails() {
	suite.simulatePublish("sensors/events", `{"messageId":"alrm"}`,
		try.HasStatusCode(http.StatusUnprocessableEntity),
		try.BodyContains("message 'alrm' not found in channel 'sensors/events'"))
}

// TestPublishDataNullIsValidatedAgainstSelectedMessage verifies that data null is data like any
// other: it is validated against the selected message instead of generating a payload.
func (suite *MqttSuite) TestPublishDataNullIsValidatedAgainstSelectedMessage() {
	suite.simulatePublish("sensors/events", `{"messageId":"alarm","data":null}`,
		try.HasStatusCode(http.StatusUnprocessableEntity))
}
