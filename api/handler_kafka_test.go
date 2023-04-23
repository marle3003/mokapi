package api

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Kafka(t *testing.T) {
	testcases := []struct {
		name         string
		app          *runtime.App
		requestUrl   string
		responseBody string
	}{
		{
			name: "get kafka services",
			app: &runtime.App{
				Monitor: monitor.New(),
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": getKafkaInfo(asyncapitest.NewConfig(
						asyncapitest.WithInfo("foo", "bar", "1.0"),
					)),
				},
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"kafka"}]`,
		},
		{
			name: "get kafka services with contact",
			app: &runtime.App{
				Monitor: monitor.New(),
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": getKafkaInfo(asyncapitest.NewConfig(
						asyncapitest.WithContact("foo", "https://foo.bar", "foo@bar.com"),
					)),
				},
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"test","contact":{"name":"foo","url":"https://foo.bar","email":"foo@bar.com"},"version":"1.0","type":"kafka"}]`,
		},
		{
			name: "get specific",
			app: &runtime.App{
				Monitor: monitor.New(),
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": getKafkaInfo(asyncapitest.NewConfig(
						asyncapitest.WithInfo("foo", "bar", "1.0"),
					)),
				},
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0"}`,
		},
		{
			name: "get specific with contact",
			app: &runtime.App{
				Monitor: monitor.New(),
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": getKafkaInfo(asyncapitest.NewConfig(
						asyncapitest.WithInfo("foo", "bar", "1.0"),
						asyncapitest.WithContact("foo", "https://foo.bar", "foo@bar.com"),
					)),
				},
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","contact":{"name":"foo","url":"https://foo.bar","email":"foo@bar.com"}}`,
		},
		{
			name: "get specific with server",
			app: &runtime.App{
				Monitor: monitor.New(),
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": getKafkaInfo(asyncapitest.NewConfig(
						asyncapitest.WithInfo("foo", "bar", "1.0"),
						asyncapitest.WithServer("foo", "kafka", "foo.bar", asyncapitest.WithServerDescription("bar")),
					)),
				},
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","servers":[{"name":"foo","url":"foo.bar","description":"bar"}]}`,
		},
		{
			name: "get specific with topic",
			app: &runtime.App{
				Monitor: monitor.New(),
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": getKafkaInfo(asyncapitest.NewConfig(
						asyncapitest.WithInfo("foo", "bar", "1.0"),
						asyncapitest.WithChannel("foo",
							asyncapitest.WithChannelDescription("bar"),
							asyncapitest.WithSubscribeAndPublish(
								asyncapitest.WithMessage(
									asyncapitest.WithPayload(schematest.New("string")),
									asyncapitest.WithContentType("application/json"),
								),
							)),
					)),
				},
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","topics":[{"name":"foo","description":"bar","partitions":[{"id":0,"startOffset":0,"offset":0,"leader":{"name":"","addr":""},"segments":0}],"configs":{"message":{"type":"string"},"messageType":"application/json"}}]}`,
		},
		{
			name: "get specific with group",
			app: &runtime.App{
				Monitor: monitor.New(),
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": getKafkaInfoWithGroup(asyncapitest.NewConfig(
						asyncapitest.WithInfo("foo", "bar", "1.0"),
						asyncapitest.WithServer("foo", "kafka", "foo.bar"),
					),
						&store.Group{
							Name:  "foo",
							State: store.PreparingRebalance,
							Generation: &store.Generation{
								Id:                 3,
								Protocol:           "range",
								LeaderId:           "",
								RebalanceTimeoutMs: 0,
							},
							Commits: nil,
						},
					),
				},
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","servers":[{"name":"foo","url":"foo.bar","description":""}],"groups":[{"name":"foo","members":null,"coordinator":"foo.bar:9092","leader":"","state":"PreparingRebalance","protocol":"range","topics":null}]}`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app, static.Api{})

			try.Handler(t,
				http.MethodGet,
				tc.requestUrl,
				nil,
				"",
				h,
				try.HasStatusCode(200),
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(tc.responseBody))
		})
	}
}

func TestHandler_Kafka_NotFound(t *testing.T) {
	h := New(runtime.New(), static.Api{})

	try.Handler(t,
		http.MethodGet,
		"http://foo.api/api/services/kafka/foo",
		nil,
		"",
		h,
		try.HasStatusCode(404))
}

func TestHandler_Kafka_Metrics(t *testing.T) {
	testcases := []struct {
		name         string
		app          *runtime.App
		requestUrl   string
		responseBody string
		addMetrics   func(monitor *monitor.Monitor)
	}{
		{
			name: "service list with metric",
			app: &runtime.App{
				Monitor: monitor.New(),
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": getKafkaInfo(asyncapitest.NewConfig(asyncapitest.WithTitle("foo"))),
				},
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","version":"1.0","type":"kafka","metrics":[{"name":"kafka_messages_total{service=\"foo\",topic=\"topic\"}","value":1}]}]`,
			addMetrics: func(monitor *monitor.Monitor) {
				monitor.Kafka.Messages.WithLabel("foo", "topic").Add(1)
			},
		},
		{
			name: "specific with metric",
			app: &runtime.App{
				Monitor: monitor.New(),
				Kafka: map[string]*runtime.KafkaInfo{
					"foo": getKafkaInfo(asyncapitest.NewConfig(asyncapitest.WithTitle("foo"))),
				},
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"","version":"1.0","metrics":[{"name":"kafka_messages_total{service=\"foo\",topic=\"topic\"}","value":1}]}`,
			addMetrics: func(monitor *monitor.Monitor) {
				monitor.Kafka.Messages.WithLabel("foo", "topic").Add(1)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app, static.Api{})
			tc.addMetrics(tc.app.Monitor)

			try.Handler(t,
				http.MethodGet,
				tc.requestUrl,
				nil,
				"",
				h,
				try.HasStatusCode(200),
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(tc.responseBody))
		})
	}
}

func getKafkaInfo(config *asyncApi.Config) *runtime.KafkaInfo {
	return &runtime.KafkaInfo{
		Config: config,
		Store:  store.New(config, enginetest.NewEngine()),
	}
}

func getKafkaInfoWithGroup(config *asyncApi.Config, group *store.Group) *runtime.KafkaInfo {
	s := store.New(config, enginetest.NewEngine())
	g := s.GetOrCreateGroup(group.Name, 0)
	group.Coordinator, _ = s.Broker(0)
	*g = *group
	return &runtime.KafkaInfo{
		Config: config,
		Store:  s,
	}
}
