package js_test

import (
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestScript_Mqtt_Publish(t *testing.T) {
	host := &enginetest.Host{
		MqttClientTest: &enginetest.MqttClient{
			PublishFunc: func(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
				return &common.MqttPublishResult{Api: "Broker", Topic: "Topic", Value: "foo"}, nil
			},
		},
	}

	s, err := jstest.New(jstest.WithSource(
		`import { publish } from 'mokapi/mqtt'
			 export default function() {
			 	const res = publish()
			 	return { api: res.api, topic: res.topic, value: res.value, cluster: res.cluster }
			 }`),
		js.WithHost(host))
	r.NoError(t, err)
	v, err := s.RunDefault()
	r.NoError(t, err)
	r.Equal(t, map[string]any{
		"api":     "Broker",
		"topic":   "Topic",
		"value":   "foo",
		"cluster": nil,
	}, v.Export())
}
