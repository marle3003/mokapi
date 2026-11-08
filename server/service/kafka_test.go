package service_test

import (
	"fmt"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"mokapi/kafka/kafkatest"
	"mokapi/server/service"
	"mokapi/try"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKafkaBroker(t *testing.T) {
	t.Parallel()
	port := try.GetFreePort()
	addr := fmt.Sprintf("127.0.0.1:%v", port)
	called := false
	handler := kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {
		called = true
		err := rw.Write(&apiVersion.Response{ApiKeys: []apiVersion.ApiKeyResponse{{ApiKey: kafka.ApiVersions, MinVersion: 1, MaxVersion: 2}}})
		require.NoError(t, err)
	})
	b := service.NewKafkaBroker(fmt.Sprintf("%v", port), handler)
	require.NoError(t, b.Start())
	defer b.Stop()

	client := kafkatest.NewClient(addr, "test")
	defer client.Close()
	r, err := client.ApiVersion(3, &apiVersion.Request{})
	require.NoError(t, err)
	require.Equal(t, kafka.None, r.ErrorCode)
	require.True(t, called, "handler should be called")
	require.Equal(t, fmt.Sprintf(":%v", port), b.Addr())

}

func newKafkaBroker(port string) *service.KafkaBroker {
	return service.NewKafkaBroker(port, kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {}))
}

func TestKafkaBroker_Start_PortIsBoundWhenStartReturns(t *testing.T) {
	requirePortIsBoundWhenStartReturns(t, newKafkaBroker)
}

func TestKafkaBroker_Start_ReturnsErrorWhenPortIsInUse(t *testing.T) {
	requireStartReturnsErrorWhenPortIsInUse(t, newKafkaBroker)
}

func TestKafkaBroker_Stop_FreesPort(t *testing.T) {
	requireStopFreesPort(t, newKafkaBroker)
}
