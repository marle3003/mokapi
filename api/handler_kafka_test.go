package api

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	kafka2 "mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/providers/openapi/openapitest"
	schematest2 "mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/schema/schematest"
	"mokapi/try"
	"net/http"
	"testing"
	"time"
)

func TestHandler_Kafka(t *testing.T) {
	mustTime := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return t
	}

	testcases := []struct {
		name         string
		app          func() *runtime.App
		requestUrl   string
		responseBody string
	}{
		{
			name: "get services",
			app: func() *runtime.App {
				return runtimetest.NewKafkaApp(
					asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
					),
				)
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"kafka"}]`,
		},
		{
			name: "get kafka services",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				_, _ = app.Kafka.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithContact("mokapi", "https://mokapi.io", "info@mokapi.io"),
					),
				}, enginetest.NewEngine())
				app.AddHttp(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("http.yaml")},
					Data: openapitest.NewConfig("3.0",
						openapitest.WithInfo("foo", "bar", "1.0"),
					),
				})
				return app
			},
			requestUrl:   "http://foo.api/api/services/kafka",
			responseBody: `[{"name":"foo","description":"bar","contact":{"name":"mokapi","url":"https://mokapi.io","email":"info@mokapi.io"},"version":"1.0"}]`,
		},
		{
			name: "get kafka services with contact",
			app: func() *runtime.App {
				return runtimetest.NewKafkaApp(
					asyncapi3test.NewConfig(
						asyncapi3test.WithContact("foo", "https://foo.bar", "foo@bar.com"),
					),
				)
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"test","contact":{"name":"foo","url":"https://foo.bar","email":"foo@bar.com"},"version":"1.0","type":"kafka"}]`,
		},
		{
			name: "get specific",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				cfg := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(),
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
					),
				}
				cfg.Info.Time = mustTime("2023-12-27T13:01:30+00:00")

				_, _ = app.Kafka.Add(cfg, enginetest.NewEngine())
				return app
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","configs":[{"id":"64613435-3062-6462-3033-316532633233","url":"file://foo.yml","provider":"test","time":"2023-12-27T13:01:30Z"}]}`,
		},
		{
			name: "get specific with contact",
			app: func() *runtime.App {
				return runtimetest.NewApp(runtimetest.WithKafkaInfo("foo", &runtime.KafkaInfo{
					Config: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithContact("foo", "https://foo.bar", "foo@bar.com"),
					),
					Store: &store.Store{},
				}))
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","contact":{"name":"foo","url":"https://foo.bar","email":"foo@bar.com"}}`,
		},
		{
			name: "get specific with server",
			app: func() *runtime.App {
				return runtimetest.NewApp(runtimetest.WithKafkaInfo("foo", &runtime.KafkaInfo{
					Config: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("bar")),
					),
					Store: &store.Store{},
				}))
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","servers":[{"name":"foo","host":"foo.bar","protocol":"kafka","description":"bar"}]}`,
		},
		{
			name: "server with tags",
			app: func() *runtime.App {
				return runtimetest.NewApp(runtimetest.WithKafkaInfo("foo", &runtime.KafkaInfo{
					Config: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithServer("foo", "kafka", "foo.bar",
							asyncapi3test.WithServerDescription("bar"),
							asyncapi3test.WithServerTags(asyncapi3.Tag{
								Name:        "env:test",
								Description: "This environment is for running internal tests",
							}),
						)),
					Store: &store.Store{},
				}))
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","servers":[{"name":"foo","host":"foo.bar","protocol":"kafka","description":"bar","tags":[{"name":"env:test","description":"This environment is for running internal tests"}]}]}`,
		},
		{
			name: "get specific with topic",
			app: func() *runtime.App {
				c := asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "bar", "1.0"),
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithChannelDescription("bar"),
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithPayload(schematest.New("string")),
							asyncapi3test.WithContentType("application/json"),
						),
					),
				)
				s := store.New(c, enginetest.NewEngine(), &eventstest.Handler{})

				return runtimetest.NewApp(runtimetest.WithKafkaInfo("foo", &runtime.KafkaInfo{
					Config: c,
					Store:  s,
				}))
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","topics":[{"name":"foo","description":"bar","partitions":[{"id":0,"startOffset":0,"offset":0,"leader":{"name":"","addr":""},"segments":0}],"messages":{"foo":{"name":"foo","payload":{"schema":{"type":"string"}},"contentType":"application/json"}},"bindings":{"partitions":1,"valueSchemaValidation":true}}]}`,
		},
		{
			name: "get specific with topic and multi schema format",
			app: func() *runtime.App {
				c := asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "bar", "1.0"),
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithChannelDescription("bar"),
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithPayloadMulti("foo", schematest.New("string")),
							asyncapi3test.WithContentType("application/json"),
						),
					),
				)
				s := store.New(c, enginetest.NewEngine(), &eventstest.Handler{})

				return runtimetest.NewApp(runtimetest.WithKafkaInfo("foo", &runtime.KafkaInfo{
					Config: c,
					Store:  s,
				}))
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","topics":[{"name":"foo","description":"bar","partitions":[{"id":0,"startOffset":0,"offset":0,"leader":{"name":"","addr":""},"segments":0}],"messages":{"foo":{"name":"foo","payload":{"format":"foo","schema":{"type":"string"}},"contentType":"application/json"}},"bindings":{"partitions":1,"valueSchemaValidation":true}}]}`,
		},
		{
			name: "get specific with group",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				app.Kafka.Set("foo", getKafkaInfoWithGroup(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "bar", "1.0"),
					asyncapi3test.WithServer("foo", "kafka", "foo.bar"),
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
				))
				return app
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","servers":[{"name":"foo","host":"foo.bar","protocol":"kafka","description":""}],"groups":[{"name":"foo","members":null,"coordinator":"foo.bar:9092","leader":"","state":"PreparingRebalance","protocol":"range","topics":null}]}`,
		},
		{
			name: "get specific with group containing members",
			app: func() *runtime.App {
				mustTime := func(s string) time.Time {
					t1, err := time.Parse(time.RFC3339, s)
					if err != nil {
						panic(err)
					}
					return t1
				}

				app := runtime.New(&static.Config{})
				app.Kafka.Set("foo", getKafkaInfoWithGroup(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "bar", "1.0"),
					asyncapi3test.WithServer("foo", "kafka", "foo.bar"),
				),
					&store.Group{
						Name:  "foo",
						State: store.PreparingRebalance,
						Generation: &store.Generation{
							Id:                 3,
							Protocol:           "range",
							LeaderId:           "m1",
							RebalanceTimeoutMs: 0,
							Members: map[string]*store.Member{
								"m1": {
									Client: &kafka2.ClientContext{
										Addr:                  "192.168.0.100",
										ClientId:              "client1",
										ClientSoftwareName:    "mokapi",
										ClientSoftwareVersion: "1.0",
										Heartbeat:             mustTime("2024-04-22T15:04:05+07:00"),
									},
									Partitions: map[string][]int{"topic": {1, 2, 5}},
								},
								"m2": {
									Client: &kafka2.ClientContext{
										Addr:                  "192.168.0.200",
										ClientId:              "client2",
										ClientSoftwareName:    "mokapi",
										ClientSoftwareVersion: "1.0",
										Heartbeat:             mustTime("2024-04-22T15:04:10+07:00"),
									},
									Partitions: map[string][]int{"topic": {3, 4, 6}},
								},
							},
						},
						Commits: nil,
					},
				))

				return app
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","servers":[{"name":"foo","host":"foo.bar","protocol":"kafka","description":""}],"groups":[{"name":"foo","members":[{"name":"m1","addr":"192.168.0.100","clientSoftwareName":"mokapi","clientSoftwareVersion":"1.0","heartbeat":"2024-04-22T15:04:05+07:00","partitions":{"topic":[1,2,5]}},{"name":"m2","addr":"192.168.0.200","clientSoftwareName":"mokapi","clientSoftwareVersion":"1.0","heartbeat":"2024-04-22T15:04:10+07:00","partitions":{"topic":[3,4,6]}}],"coordinator":"foo.bar:9092","leader":"m1","state":"PreparingRebalance","protocol":"range","topics":null}]}`,
		},
		{
			name: "get specific with topic and openapi schema",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				app.Kafka.Set("foo", getKafkaInfo(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "bar", "1.0"),
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithChannelDescription("bar"),
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithPayloadMulti("foo", schematest2.New("string")),
							asyncapi3test.WithContentType("application/json"),
						),
					),
				)))
				return app
			},
			requestUrl:   "http://foo.api/api/services/kafka/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","topics":[{"name":"foo","description":"bar","partitions":[{"id":0,"startOffset":0,"offset":0,"leader":{"name":"","addr":""},"segments":0}],"messages":{"foo":{"name":"foo","payload":{"format":"foo","schema":{"type":"string"}},"contentType":"application/json"}},"bindings":{"partitions":1,"valueSchemaValidation":true}}]}`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app(), static.Api{})

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

func TestHandler_KafkaAPI(t *testing.T) {
	type request struct {
		method  string
		url     string
		headers map[string]string
		body    string
	}

	testcases := []struct {
		name    string
		app     func() *runtime.App
		request request
		expect  []try.ResponseCondition
	}{
		{
			name: "get kafka topics but empty",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				_, _ = app.Kafka.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
					),
				}, enginetest.NewEngine())
				return app
			},
			request: request{url: "http://foo.api/api/services/kafka/foo/topics"},
			expect: []try.ResponseCondition{
				try.HasBody(`[]`),
			},
		},
		{
			name: "get kafka topics with one topic",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				_, _ = app.Kafka.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithServer("broker-1", "kafka", "localhost:9092"),
						asyncapi3test.WithChannel("topic-1",
							asyncapi3test.WithChannelDescription("foobar"),
							asyncapi3test.WithMessage("foo"),
						),
					),
				}, enginetest.NewEngine())
				return app
			},
			request: request{url: "http://foo.api/api/services/kafka/foo/topics"},
			expect: []try.ResponseCondition{
				try.HasBody(`[{"name":"topic-1","description":"foobar","partitions":[{"id":0,"startOffset":0,"offset":0,"leader":{"name":"broker-1","addr":"localhost:9092"},"segments":0}],"messages":{"foo":{"name":"foo","payload":null,"contentType":"application/json"}},"bindings":{"partitions":1,"valueSchemaValidation":true}}]`),
			},
		},
		{
			name: "get specific kafka topic",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				_, _ = app.Kafka.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithServer("broker-1", "kafka", "localhost:9092"),
						asyncapi3test.WithChannel("topic-1",
							asyncapi3test.WithChannelDescription("foobar"),
							asyncapi3test.WithMessage("foo"),
						),
					),
				}, enginetest.NewEngine())
				return app
			},
			request: request{url: "http://foo.api/api/services/kafka/foo/topics/topic-1"},
			expect: []try.ResponseCondition{
				try.HasBody(`{"name":"topic-1","description":"foobar","partitions":[{"id":0,"startOffset":0,"offset":0,"leader":{"name":"broker-1","addr":"localhost:9092"},"segments":0}],"messages":{"foo":{"name":"foo","payload":null,"contentType":"application/json"}},"bindings":{"partitions":1,"valueSchemaValidation":true}}`),
			},
		},
		{
			name: "get specific kafka topic but not found",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				_, _ = app.Kafka.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithServer("broker-1", "kafka", "localhost:9092"),
						asyncapi3test.WithChannel("topic-1",
							asyncapi3test.WithChannelDescription("foobar"),
							asyncapi3test.WithMessage("foo"),
						),
					),
				}, enginetest.NewEngine())
				return app
			},
			request: request{url: "http://foo.api/api/services/kafka/foo/topics/foo"},
			expect: []try.ResponseCondition{
				try.HasStatusCode(404),
			},
		},
		{
			name: "produce kafka message into topic",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				_, _ = app.Kafka.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithServer("broker-1", "kafka", "localhost:9092"),
						asyncapi3test.WithChannel("topic-1",
							asyncapi3test.WithChannelDescription("foobar"),
							asyncapi3test.WithMessage("foo"),
						),
					),
				}, enginetest.NewEngine())
				return app
			},
			request: request{
				method:  "POST",
				url:     "http://foo.api/api/services/kafka/foo/topics/topic-1",
				headers: map[string]string{"Content-Type": "application/json"},
				body: `{
"records": [{"key": "foo", "value": "bar"}]
}`,
			},
			expect: []try.ResponseCondition{
				try.HasBody(`{"offsets":[{"partition":0,"offset":0,"error":""}]}`),
			},
		},
		{
			name: "get kafka partitions",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				_, _ = app.Kafka.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithServer("broker-1", "kafka", "localhost:9092"),
						asyncapi3test.WithChannel("topic-1",
							asyncapi3test.WithChannelDescription("foobar"),
							asyncapi3test.WithMessage("foo"),
						),
					),
				}, enginetest.NewEngine())
				return app
			},
			request: request{url: "http://foo.api/api/services/kafka/foo/topics/topic-1/partitions"},
			expect: []try.ResponseCondition{
				try.HasBody(`[{"id":0,"startOffset":0,"offset":0,"leader":{"name":"broker-1","addr":"localhost:9092"},"segments":0}]`),
			},
		},
		{
			name: "get specific kafka partition",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				_, _ = app.Kafka.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithServer("broker-1", "kafka", "localhost:9092"),
						asyncapi3test.WithChannel("topic-1",
							asyncapi3test.WithChannelDescription("foobar"),
							asyncapi3test.WithMessage("foo"),
						),
					),
				}, enginetest.NewEngine())
				return app
			},
			request: request{url: "http://foo.api/api/services/kafka/foo/topics/topic-1/partitions/0"},
			expect: []try.ResponseCondition{
				try.HasBody(`{"id":0,"startOffset":0,"offset":0,"leader":{"name":"broker-1","addr":"localhost:9092"},"segments":0}`),
			},
		},
		{
			name: "produce kafka message into specific partition",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				_, _ = app.Kafka.Add(&dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("kafka.yaml")},
					Data: asyncapi3test.NewConfig(
						asyncapi3test.WithInfo("foo", "bar", "1.0"),
						asyncapi3test.WithServer("broker-1", "kafka", "localhost:9092"),
						asyncapi3test.WithChannel("topic-1",
							asyncapi3test.WithChannelDescription("foobar"),
							asyncapi3test.WithMessage("foo"),
						),
					),
				}, enginetest.NewEngine())
				return app
			},
			request: request{
				method:  "POST",
				url:     "http://foo.api/api/services/kafka/foo/topics/topic-1/partitions/0",
				headers: map[string]string{"Content-Type": "application/json"},
				body: `{
"records": [{"key": "foo", "value": "bar"}]
}`,
			},
			expect: []try.ResponseCondition{
				try.HasBody(`{"offsets":[{"partition":0,"offset":0,"error":""}]}`),
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app(), static.Api{})

			method := tc.request.method
			if method == "" {
				method = http.MethodGet
			}

			try.Handler(t,
				method,
				tc.request.url,
				tc.request.headers,
				tc.request.body,
				h,
				tc.expect...)
		})
	}
}

func TestHandler_Kafka_NotFound(t *testing.T) {
	h := New(runtime.New(&static.Config{}), static.Api{})

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
			name:         "service list with metric",
			app:          runtimetest.NewApp(runtimetest.WithKafkaInfo("foo", getKafkaInfo(asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo"))))),
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","version":"1.0","type":"kafka","metrics":[{"name":"kafka_messages_total{service=\"foo\",topic=\"topic\"}","value":1}]}]`,
			addMetrics: func(monitor *monitor.Monitor) {
				monitor.Kafka.Messages.WithLabel("foo", "topic").Add(1)
			},
		},
		{
			name:         "specific with metric",
			app:          runtimetest.NewApp(runtimetest.WithKafkaInfo("foo", getKafkaInfo(asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo"))))),
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

func getKafkaInfo(config *asyncapi3.Config) *runtime.KafkaInfo {
	return &runtime.KafkaInfo{
		Config: config,
		Store:  store.New(config, enginetest.NewEngine(), &eventstest.Handler{}),
	}
}

func getKafkaInfoWithGroup(config *asyncapi3.Config, group *store.Group) *runtime.KafkaInfo {
	s := store.New(config, enginetest.NewEngine(), &eventstest.Handler{})
	g := s.GetOrCreateGroup(group.Name, 0)
	group.Coordinator, _ = s.Broker(0)
	*g = *group
	return &runtime.KafkaInfo{
		Config: config,
		Store:  s,
	}
}
