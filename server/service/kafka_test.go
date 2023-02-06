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
	t.Parallel()
	port, err := try.GetFreePort()
	require.NoError(t, err)
	addr := fmt.Sprintf("127.0.0.1:%v", port)
	called := false
	handler := kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {
		called = true
		rw.Write(&apiVersion.Response{})
	})
	b := NewKafkaBroker(fmt.Sprintf("%v", port), handler)
	b.Start()
	defer b.Stop()

	client := kafkatest.NewClient(addr, "test")
	r, err := client.ApiVersion(3, &apiVersion.Request{})
	require.Equal(t, kafka.None, r.ErrorCode)
	require.True(t, called, "handler should be called")
}
