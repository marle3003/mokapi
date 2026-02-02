package api

import (
	"context"
	"encoding/json"
	"fmt"
	"mokapi/config/static"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/try"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandler_Events(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, h http.Handler, sm *events.StoreManager)
	}{
		{
			name: "empty http events",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?namespace=http",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[]`))
			},
		},
		{
			name: "with http events",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("http"))
				err := sm.Push(&eventstest.Event{Name: "foo"}, events.NewTraits().WithNamespace("http"))
				event := sm.GetEvents(events.NewTraits())[0]
				require.NoError(t, err)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?namespace=http",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`[{"id":"%v","traits":{"namespace":"http"},"data":{"Name":"foo","api":""},"time":"%v"}]`,
						event.Id,
						event.Time.Format(time.RFC3339Nano))))
			},
		},
		{
			name: "get specific event",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("http"))
				err := sm.Push(&eventstest.Event{Name: "foo"}, events.NewTraits().WithNamespace("http"))
				event := sm.GetEvents(events.NewTraits())[0]
				require.NoError(t, err)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/"+event.Id,
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`{"id":"%v","traits":{"namespace":"http"},"data":{"Name":"foo","api":""},"time":"%v"}`,
						event.Id,
						event.Time.Format(time.RFC3339Nano))))
			},
		},
		{
			name: "get http event with header parameter as string",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("http"))

				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("foo", "bar")

				params := &openapi.RequestParameters{Header: map[string]openapi.RequestParameterValue{}}
				v := "bar"
				params.Header["Foo"] = openapi.RequestParameterValue{
					Value: "bar",
					Raw:   &v,
				}
				r = r.WithContext(openapi.NewContext(context.Background(), params))

				_, err := openapi.NewLogEventContext(r, false, sm, events.NewTraits().WithNamespace("http"))
				require.NoError(t, err)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?namespace=http",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.AssertBody(func(t *testing.T, body string) {
						var m []map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &m))
						require.Equal(t, map[string]interface{}{
							"actions":    interface{}(nil),
							"api":        "",
							"deprecated": false,
							"duration":   float64(0),
							"path":       "",
							"request": map[string]interface{}{
								"method": "get",
								"parameters": []interface{}{
									map[string]interface{}{
										"name":  "Foo",
										"raw":   "bar",
										"type":  "header",
										"value": "\"bar\"",
									},
								},
								"url": "http://localhost/foo",
							},
							"response": map[string]interface{}{"body": "", "size": float64(0), "statusCode": float64(0)},
							"clientIP": "192.0.2.1",
						},
							m[0]["data"])
					}))
			},
		},
		{
			name: "get http event with header parameter as object",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("http"))

				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("foo", "role,admin,firstName,Alex")

				params := &openapi.RequestParameters{Header: map[string]openapi.RequestParameterValue{}}
				v := "bar"
				params.Header["Foo"] = openapi.RequestParameterValue{
					Value: map[string]any{"role": "admin", "firstName": "Alex"},
					Raw:   &v,
				}
				r = r.WithContext(openapi.NewContext(context.Background(), params))

				_, err := openapi.NewLogEventContext(r, false, sm, events.NewTraits().WithNamespace("http"))
				require.NoError(t, err)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?namespace=http",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.AssertBody(func(t *testing.T, body string) {
						var m []map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &m))
						require.Equal(t, map[string]interface{}{
							"actions":    interface{}(nil),
							"api":        "",
							"deprecated": false,
							"duration":   float64(0),
							"path":       "",
							"request": map[string]interface{}{
								"method": "get",
								"parameters": []interface{}{
									map[string]interface{}{
										"name":  "Foo",
										"raw":   "bar",
										"type":  "header",
										"value": "{\"firstName\":\"Alex\",\"role\":\"admin\"}",
									},
								},
								"url": "http://localhost/foo",
							},
							"response": map[string]interface{}{"body": "", "size": float64(0), "statusCode": float64(0)},
							"clientIP": "192.0.2.1",
						},
							m[0]["data"])
					}))
			},
		},
		{
			name: "http with request parameter",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/1234",
					nil,
					"",
					h,
					try.HasStatusCode(404))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg)

			h := New(app, static.Api{})
			tc.fn(t, h, app.Events)
		})
	}
}

func TestHandler_KafkaEvents(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, h http.Handler, sm *events.StoreManager)
	}{
		{
			name: "empty kafka events",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?namespace=kafka",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[]`))
			},
		},
		{
			name: "with kafka events",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("kafka"))
				err := sm.Push(&eventstest.Event{Name: "foo"}, events.NewTraits().WithNamespace("kafka"))
				event := sm.GetEvents(events.NewTraits())[0]
				require.NoError(t, err)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?namespace=kafka",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`[{"id":"%v","traits":{"namespace":"kafka"},"data":{"Name":"foo","api":""},"time":"%v"}]`,
						event.Id,
						event.Time.Format(time.RFC3339Nano))))
			},
		},
		{
			name: "get specific event",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("kafka"))
				err := sm.Push(&eventstest.Event{Name: "foo"}, events.NewTraits().WithNamespace("kafka"))
				event := sm.GetEvents(events.NewTraits())[0]
				require.NoError(t, err)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/"+event.Id,
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`{"id":"%v","traits":{"namespace":"kafka"},"data":{"Name":"foo","api":""},"time":"%v"}`,
						event.Id,
						event.Time.Format(time.RFC3339Nano))))
			},
		},
		{
			name: "get kafka with producerId",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("kafka"))

				err := sm.Push(&store.KafkaMessageLog{
					Offset:         123,
					Key:            store.LogValue{},
					Message:        store.LogValue{},
					MessageId:      "foo-1",
					Partition:      1,
					ProducerId:     3,
					ProducerEpoch:  1,
					SequenceNumber: 2,
					ClientId:       "foo",
					ScriptFile:     "123-123",
				}, events.NewTraits().WithNamespace("kafka"))
				require.NoError(t, err)

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?namespace=kafka",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.AssertBody(func(t *testing.T, body string) {
						var m []map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &m))
						require.Equal(t, map[string]any{
							"api":     "",
							"deleted": false,
							"headers": interface{}(nil),
							"key": map[string]interface{}{
								"binary": interface{}(nil),
								"value":  "",
							},
							"message": map[string]interface{}{
								"binary": interface{}(nil),
								"value":  "",
							},
							"messageId":      "foo-1",
							"offset":         float64(123),
							"partition":      float64(1),
							"producerEpoch":  float64(1),
							"producerId":     float64(3),
							"schemaId":       float64(0),
							"sequenceNumber": float64(2),
							"clientId":       "foo",
							"script":         "123-123",
						},
							m[0]["data"])
					}))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg)

			h := New(app, static.Api{})
			tc.fn(t, h, app.Events)
		})
	}
}
