package kafka_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"mokapi/kafka/kafkatest"
	"regexp"
	"testing"
)

func TestServer(t *testing.T) {
	testcases := []struct {
		name    string
		handler func(rw kafka.ResponseWriter, req *kafka.Request)
		fn      func(client *kafkatest.Client)
	}{
		{"version 1",
			func(rw kafka.ResponseWriter, req *kafka.Request) {
				require.Equal(t, kafka.ApiVersions, req.Header.ApiKey)
				require.Equal(t, int16(1), req.Header.ApiVersion)
				require.Regexp(t, regexp.MustCompile(`^127\.0\.0\.1:[0-9]+$`), req.Host)

				msg, ok := req.Message.(*apiVersion.Request)
				require.True(t, ok)
				require.Equal(t, "", msg.ClientSwName)
				require.Equal(t, "", msg.ClientSwName)

				rw.WriteHeader(kafka.ApiVersions, 1, 0)
				rw.Write(&apiVersion.Response{})
			},
			func(client *kafkatest.Client) {
				_, err := client.Send(kafkatest.NewRequest("kafkatest", 1,
					&apiVersion.Request{ClientSwName: "foo", ClientSwVersion: "bar"}))
				require.NoError(t, err)
			}},
		{"version 3",
			func(rw kafka.ResponseWriter, req *kafka.Request) {
				require.Equal(t, kafka.ApiVersions, req.Header.ApiKey)
				require.Equal(t, int16(3), req.Header.ApiVersion)
				require.Regexp(t, regexp.MustCompile(`^127\.0\.0\.1:[0-9]+$`), req.Host)

				msg, ok := req.Message.(*apiVersion.Request)
				require.True(t, ok)
				require.Equal(t, "foo", msg.ClientSwName)
				require.Equal(t, "bar", msg.ClientSwVersion)

				rw.WriteHeader(kafka.ApiVersions, 1, 0)
				rw.Write(&apiVersion.Response{})
			}, func(client *kafkatest.Client) {
				_, err := client.Send(kafkatest.NewRequest("kafkatest", 3,
					&apiVersion.Request{ClientSwName: "foo", ClientSwVersion: "bar"}))
				require.NoError(t, err)

			}},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ts := kafkatest.NewServer(tc.handler)
			defer ts.Close()
			ts.Start()

			client := kafkatest.NewClient(ts.Listener.Addr().String(), "kafkatest")
			defer client.Close()

			tc.fn(client)
		})
	}
}
