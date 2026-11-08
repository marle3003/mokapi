package api_test

import (
	"fmt"
	"mokapi/api"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/kafka"
	"mokapi/kafka/heartbeat"
	"mokapi/kafka/kafkatest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"sync"
	"testing"
)

// The dashboard shows a client while the client joins consumer groups. Run with -race.
func TestHandler_Kafka_ClientIsReadWhileClientJoinsGroups(t *testing.T) {
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	info := getKafkaInfo(asyncapi3test.NewConfig(
		asyncapi3test.WithInfo("foo", "bar", "1.0"),
		asyncapi3test.WithServer("foo", "kafka", "127.0.0.1:9092"),
	))
	app.Kafka.Set("foo", info)
	h := api.New(app, static.Api{})

	req := kafkatest.NewRequest("client1", 4, &heartbeat.Request{GroupId: "group-0"})
	client := kafka.ClientFromContext(req.Context)
	info.Store.ServeMessage(kafkatest.NewRecorder(), req)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			client.AddGroup(fmt.Sprintf("group-%d", i), "member")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			try.Handler(t,
				http.MethodGet,
				"http://foo.api/api/services/kafka/foo/clients/client1",
				nil, "", h,
				try.HasStatusCode(http.StatusOK),
			)
		}
	}()
	wg.Wait()
}
