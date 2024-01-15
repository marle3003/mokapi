package runtime

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"mokapi/kafka/kafkatest"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"net/url"
	"testing"
)

func TestApp_AddKafka(t *testing.T) {
	defer events.Reset()

	app := New()
	c := asyncapitest.NewConfig(asyncapitest.WithInfo("foo", "", ""))
	app.AddKafka(getConfig(c), enginetest.NewEngine())

	require.Contains(t, app.Kafka, "foo")
	err := events.Push("bar", events.NewTraits().WithNamespace("kafka").WithName("foo"))
	require.NoError(t, err, "event store should be available")
}

func TestApp_AddKafka_Topic(t *testing.T) {
	defer events.Reset()

	app := New()
	c := asyncapitest.NewConfig(asyncapitest.WithInfo("foo", "", ""), asyncapitest.WithChannel("bar"))
	app.AddKafka(getConfig(c), enginetest.NewEngine())

	require.Contains(t, app.Kafka, "foo")
	err := events.Push("bar", events.NewTraits().WithNamespace("kafka").WithName("foo").With("path", "bar"))
	require.NoError(t, err, "event store should be available")
}

func TestKafkaHandler(t *testing.T) {
	hf := kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {
		v, ok := monitor.KafkaFromContext(req.Context)
		require.True(t, ok)
		require.NotNil(t, v)
	})
	h := &KafkaHandler{
		kafka: New().Monitor.Kafka,
		next:  hf,
	}

	h.ServeMessage(kafkatest.NewRecorder(), kafkatest.NewRequest("", 1.0, &apiVersion.Request{}))
}

func getConfig(c *asyncApi.Config) *dynamic.Config {
	u, _ := url.Parse("foo.bar")
	cfg := &dynamic.Config{Data: c}
	cfg.Info.Url = u
	return cfg
}
