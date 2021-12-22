package protocol_test

import (
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/apiVersion"
	"mokapi/test"
	"testing"
)

func TestApiVersion(t *testing.T) {
	testdata := []struct {
		name    string
		handler func(rw protocol.ResponseWriter, req *protocol.Request)
		fn      func(client *kafkatest.Client)
	}{
		{"version 1",
			func(rw protocol.ResponseWriter, req *protocol.Request) {
				test.Equals(t, protocol.ApiVersions, req.Header.ApiKey)
				test.Equals(t, int16(1), req.Header.ApiVersion)

				msg, ok := req.Message.(*apiVersion.Request)
				test.Equals(t, true, ok)
				test.Equals(t, "", msg.ClientSwName)
				test.Equals(t, "", msg.ClientSwName)

				rw.WriteHeader(protocol.ApiVersions, 1, 0)
				rw.Write(&apiVersion.Response{})
			},
			func(client *kafkatest.Client) {
				_, err := client.Send(kafkatest.NewRequest("kafkatest", 1,
					&apiVersion.Request{ClientSwName: "foo", ClientSwVersion: "bar"}))
				test.Ok(t, err)
			}},
		{"version 3",
			func(rw protocol.ResponseWriter, req *protocol.Request) {
				test.Equals(t, protocol.ApiVersions, req.Header.ApiKey)
				test.Equals(t, int16(3), req.Header.ApiVersion)

				msg, ok := req.Message.(*apiVersion.Request)
				test.Equals(t, true, ok)
				test.Equals(t, "foo", msg.ClientSwName)
				test.Equals(t, "bar", msg.ClientSwVersion)

				rw.WriteHeader(protocol.ApiVersions, 1, 0)
				rw.Write(&apiVersion.Response{})
			}, func(client *kafkatest.Client) {
				_, err := client.Send(kafkatest.NewRequest("kafkatest", 3,
					&apiVersion.Request{ClientSwName: "foo", ClientSwVersion: "bar"}))
				test.Ok(t, err)

			}},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			ts := kafkatest.NewServer(data.handler)
			ts.Start()

			client := kafkatest.NewClient(ts.Listener.Addr().String(), "kafkatest")

			data.fn(client)

			ts.Close()
			client.Close()
		})
	}
}
