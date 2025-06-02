package service

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/try"
	"testing"
)

func TestMqttBroker(t *testing.T) {
	t.Parallel()
	port := try.GetFreePort()
	addr := fmt.Sprintf("127.0.0.1:%v", port)
	called := false
	handler := mqtt.HandlerFunc(func(rw mqtt.ResponseWriter, req *mqtt.Request) {
		called = true
		rw.Write(mqtt.CONNACK, &mqtt.ConnectResponse{
			SessionPresent: false,
			ReturnCode:     mqtt.Accepted,
		})
	})
	b := NewMqttBroker(fmt.Sprintf("%v", port), handler)
	b.Start()
	defer b.Stop()

	client := mqtttest.Client{Addr: addr}
	defer client.Close()
	_, err := client.Send(&mqtt.Request{
		// The DUP, QoS, and RETAIN flags are not used in the CONNECT message.
		Header: &mqtt.Header{
			Type: mqtt.CONNECT,
		},
		Message: &mqtt.ConnectRequest{
			Header: mqtt.ConnectHeader{
				Protocol:     "MQTT",
				Version:      4,
				CleanSession: true,
				KeepAlive:    60,
			},
			ClientId: "client-foo",
		},
		Context: nil,
	})
	require.NoError(t, err)
	require.True(t, called, "handler should be called")
}
