package service

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"mokapi/kafka/kafkatest"
	"mokapi/try"
	"testing"
)

func TestKafkaBroker(t *testing.T) {
	//t.Parallel()
	port := try.GetFreePort()
	addr := fmt.Sprintf("127.0.0.1:%v", port)
	called := false
	handler := kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {
		called = true
		err := rw.Write(&apiVersion.Response{ApiKeys: []apiVersion.ApiKeyResponse{{ApiKey: kafka.ApiVersions, MinVersion: 1, MaxVersion: 2}}})
		require.NoError(t, err)
	})
	b := NewKafkaBroker(fmt.Sprintf("%v", port), handler)
	b.Start()
	defer b.Stop()

	client := kafkatest.NewClient(addr, "test")
	defer client.Close()
	r, err := client.ApiVersion(3, &apiVersion.Request{})
	require.NoError(t, err)
	require.Equal(t, kafka.None, r.ErrorCode)
	require.True(t, called, "handler should be called")

}
