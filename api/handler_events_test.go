package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"mokapi/api"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	kafka "mokapi/providers/asyncapi3/kafka/store"
	mqtt "mokapi/providers/asyncapi3/mqtt/store"
	"mokapi/providers/asyncapi3/websocket"
	"mokapi/providers/directory"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/try"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandler_Events(t *testing.T) {

	newHttpEvent := func(t *testing.T, r *http.Request, traits events.Traits, sm *events.StoreManager) events.Event {
		ctx, err := openapi.NewLogEventContext(r, false, traits)
		require.NoError(t, err)
		l, _ := openapi.LogEventFromContext(ctx)
		err = sm.Push(l, traits)
		require.NoError(t, err)
		evts := sm.GetEvents(traits)
		// events are LIFO
		return evts[0]
	}

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
					try.HasBody(`{"items":[],"page":1,"size":25,"total":0}`))
			},
		},
		{
			name: "paging",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(5, events.NewTraits().WithNamespace("http"))
				var ids []string
				for i := 1; i <= 5; i++ {
					r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost/foo%v", i), nil)
					evt := newHttpEvent(t, r, events.NewTraits().WithNamespace("http"), sm)
					ids = append(ids, evt.Id)
				}

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?traits[namespace]=http&page=2&size=2",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.AssertBody(func(t *testing.T, body string) {
						var data api.EventPage
						require.NoError(t, json.Unmarshal([]byte(body), &data))
						require.Equal(t, 2, data.Page)
						require.Equal(t, 2, data.Size)
						require.Equal(t, 5, data.Total)
						require.Len(t, data.Items, 2)
						// ids are FIFO, events are LIFO
						require.Equal(t, ids[2], data.Items[0].Id)
						require.Equal(t, ids[1], data.Items[1].Id)
						require.Equal(t, map[string]any{"deprecated": false, "duration": float64(0), "statusCode": float64(0), "url": "http://localhost/foo3"}, data.Items[0].Data)
						require.Equal(t, map[string]any{"deprecated": false, "duration": float64(0), "statusCode": float64(0), "url": "http://localhost/foo2"}, data.Items[1].Data)
					}),
				)
			},
		},
		{
			name: "filter http events",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("http"))
				sm.SetStore(1, events.NewTraits().WithNamespace("kafka"))

				err := sm.Push(&eventstest.Event{Name: "foo"}, events.NewTraits().WithNamespace("http"))
				require.NoError(t, err)
				err = sm.Push(&eventstest.Event{Name: "foo"}, events.NewTraits().WithNamespace("kafka"))
				require.NoError(t, err)

				evt := sm.GetEvents(events.NewTraits().WithNamespace("http"))[0]
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?traits[namespace]=http",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.AssertBody(func(t *testing.T, body string) {
						var data map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &data))
						require.Len(t, data["items"], 1)
						require.Equal(t, []any{
							map[string]any{
								"id":     evt.Id,
								"traits": map[string]any{"namespace": "http"},
								"time":   evt.Time.Format(time.RFC3339Nano),
								"data": map[string]any{
									"name": "foo",
									"api":  "",
								},
							},
						}, data["items"])
					}),
				)
			},
		},
		{
			name: "filter http method",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(2, events.NewTraits().WithNamespace("http"))

				r1 := httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil)
				r2 := httptest.NewRequest(http.MethodPost, "http://localhost/foo", nil)
				evt1 := newHttpEvent(t, r1, events.NewTraits().WithNamespace("http"), sm)
				newHttpEvent(t, r2, events.NewTraits().WithNamespace("http"), sm)

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?traits[namespace]=http&http[method]=GET",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.AssertBody(func(t *testing.T, body string) {
						var data map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &data))
						require.Len(t, data["items"], 1)
						require.Equal(t, []any{
							map[string]any{
								"id":     evt1.Id,
								"traits": map[string]any{"namespace": "http"},
								"time":   evt1.Time.Format(time.RFC3339Nano),
								"data":   map[string]any{"deprecated": false, "duration": float64(0), "statusCode": float64(0), "url": "http://localhost/foo"},
							},
						}, data["items"])
					}),
				)
			},
		},
		{
			name: "filter http methods",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(2, events.NewTraits().WithNamespace("http"))

				r1 := httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil)
				r2 := httptest.NewRequest(http.MethodPost, "http://localhost/foo", nil)
				evt1 := newHttpEvent(t, r1, events.NewTraits().WithNamespace("http"), sm)
				evt2 := newHttpEvent(t, r2, events.NewTraits().WithNamespace("http"), sm)

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?traits[namespace]=http&http[method]=GET,POST",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.AssertBody(func(t *testing.T, body string) {
						var data map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &data))
						require.Len(t, data["items"], 2)
						require.Equal(t, []any{
							map[string]any{
								"id":     evt2.Id,
								"traits": map[string]any{"namespace": "http"},
								"time":   evt2.Time.Format(time.RFC3339Nano),
								"data":   map[string]any{"deprecated": false, "duration": float64(0), "statusCode": float64(0), "url": "http://localhost/foo"},
							},
							map[string]any{
								"id":     evt1.Id,
								"traits": map[string]any{"namespace": "http"},
								"time":   evt1.Time.Format(time.RFC3339Nano),
								"data":   map[string]any{"deprecated": false, "duration": float64(0), "statusCode": float64(0), "url": "http://localhost/foo"},
							},
						}, data["items"])
					}),
				)
			},
		},
		{
			name: "filter http url",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(2, events.NewTraits().WithNamespace("http"))

				r1 := httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil)
				r2 := httptest.NewRequest(http.MethodPost, "http://localhost/bar", nil)
				newHttpEvent(t, r1, events.NewTraits().WithNamespace("http"), sm)
				evt2 := newHttpEvent(t, r2, events.NewTraits().WithNamespace("http"), sm)

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?traits[namespace]=http&http[url]=bar",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.AssertBody(func(t *testing.T, body string) {
						var data map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &data))
						require.Len(t, data["items"], 1)
						require.Equal(t, []any{
							map[string]any{
								"id":     evt2.Id,
								"traits": map[string]any{"namespace": "http"},
								"time":   evt2.Time.Format(time.RFC3339Nano),
								"data":   map[string]any{"deprecated": false, "duration": float64(0), "statusCode": float64(0), "url": "http://localhost/bar"},
							},
						}, data["items"])
					}),
				)
			},
		},
		{
			name: "filter http request header",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(2, events.NewTraits().WithNamespace("http"))

				r1 := httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil)
				r2 := httptest.NewRequest(http.MethodPost, "http://localhost/bar", nil)
				newHttpEvent(t, r1, events.NewTraits().WithNamespace("http"), sm)
				evt2 := newHttpEvent(t, r2, events.NewTraits().WithNamespace("http"), sm)
				log := evt2.Data.(*openapi.HttpLog)
				log.Request.Parameters = append(log.Request.Parameters, openapi.HttpParameter{
					Name:  "role",
					Type:  "header",
					Value: "",
					Raw:   new("admin"),
				})

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?traits[namespace]=http&http[requestHeader]=role:admin",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.AssertBody(func(t *testing.T, body string) {
						var data map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &data))
						require.Len(t, data["items"], 1)
						require.Equal(t, []any{
							map[string]any{
								"id":     evt2.Id,
								"traits": map[string]any{"namespace": "http"},
								"time":   evt2.Time.Format(time.RFC3339Nano),
								"data":   map[string]any{"deprecated": false, "duration": float64(0), "statusCode": float64(0), "url": "http://localhost/bar"},
							},
						}, data["items"])
					}),
				)
			},
		},
		{
			name: "filter http status code",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(2, events.NewTraits().WithNamespace("http"))

				data := []struct {
					url    string
					method string
					status int
				}{
					{
						url:    "http://localhost/foo",
						method: http.MethodGet,
						status: 200,
					},
					{
						url:    "http://localhost/foo",
						method: http.MethodGet,
						status: 201,
					},
					{
						url:    "http://localhost/foo",
						method: http.MethodGet,
						status: 301,
					},
				}

				for _, d := range data {
					r := httptest.NewRequest(d.method, d.url, nil)
					evt := newHttpEvent(t, r, events.NewTraits().WithNamespace("http"), sm)
					log := evt.Data.(*openapi.HttpLog)
					log.Response.StatusCode = d.status
				}

				getStatusCode := func(index int, items []any) int {
					if index >= len(items) {
						return -1
					}
					m := items[index].(map[string]any)
					d := m["data"].(map[string]any)
					return int(d["statusCode"].(float64))
				}

				tests := []struct {
					url   string
					check try.ResponseCondition
				}{
					{
						url: "http://foo.api/api/events?traits[namespace]=http&http[statusCode]=200",
						check: try.AssertBody(func(t *testing.T, body string) {
							var data map[string]any
							require.NoError(t, json.Unmarshal([]byte(body), &data))
							items := data["items"].([]any)
							require.Len(t, items, 1)
							require.Equal(t, 200, getStatusCode(0, items))
						}),
					},
					{
						url: "http://foo.api/api/events?traits[namespace]=http&http[statusCode]=200-201",
						check: try.AssertBody(func(t *testing.T, body string) {
							var data map[string]any
							require.NoError(t, json.Unmarshal([]byte(body), &data))
							items := data["items"].([]any)
							require.Len(t, items, 2)
							require.Equal(t, 201, getStatusCode(0, items))
							require.Equal(t, 200, getStatusCode(1, items))
						}),
					},
					{
						url: "http://foo.api/api/events?traits[namespace]=http&http[statusCode]=2xx",
						check: try.AssertBody(func(t *testing.T, body string) {
							var data map[string]any
							require.NoError(t, json.Unmarshal([]byte(body), &data))
							items := data["items"].([]any)
							require.Len(t, items, 2)
							require.Equal(t, 201, getStatusCode(0, items))
							require.Equal(t, 200, getStatusCode(1, items))
						}),
					},
					{
						url: "http://foo.api/api/events?traits[namespace]=http&http[statusCode]=!200-201",
						check: try.AssertBody(func(t *testing.T, body string) {
							var data map[string]any
							require.NoError(t, json.Unmarshal([]byte(body), &data))
							items := data["items"].([]any)
							require.Len(t, items, 1)
							require.Equal(t, 301, getStatusCode(0, items))
						}),
					},
				}

				for _, d := range tests {
					try.Handler(t,
						http.MethodGet,
						d.url,
						nil,
						"",
						h,
						try.HasStatusCode(200),
						try.HasHeader("Content-Type", "application/json"),
						d.check,
					)
				}

			},
		},
		{
			name: "filter http response header",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(2, events.NewTraits().WithNamespace("http"))

				r1 := httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil)
				r2 := httptest.NewRequest(http.MethodPost, "http://localhost/bar", nil)
				newHttpEvent(t, r1, events.NewTraits().WithNamespace("http"), sm)
				evt2 := newHttpEvent(t, r2, events.NewTraits().WithNamespace("http"), sm)
				log := evt2.Data.(*openapi.HttpLog)
				log.Response.Headers = map[string]string{"role": "admin"}

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events?traits[namespace]=http&http[responseHeader]=role:admin",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.AssertBody(func(t *testing.T, body string) {
						var data map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &data))
						require.Len(t, data["items"], 1)
						require.Equal(t, []any{
							map[string]any{
								"id":     evt2.Id,
								"traits": map[string]any{"namespace": "http"},
								"time":   evt2.Time.Format(time.RFC3339Nano),
								"data":   map[string]any{"deprecated": false, "duration": float64(0), "statusCode": float64(0), "url": "http://localhost/bar"},
							},
						}, data["items"])
					}),
				)
			},
		},
		{
			name: "filter clientIP",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(2, events.NewTraits().WithNamespace("http"))

				data := []struct {
					url      string
					method   string
					clientIP string
				}{
					{
						url:      "http://localhost/foo1",
						method:   http.MethodGet,
						clientIP: "127.0.0.1",
					},
					{
						url:      "http://localhost/foo2",
						method:   http.MethodGet,
						clientIP: "127.0.0.1",
					},
					{
						url:      "http://localhost/foo3",
						method:   http.MethodGet,
						clientIP: "192.0.1.24",
					},
				}

				for _, d := range data {
					r := httptest.NewRequest(d.method, d.url, nil)
					evt := newHttpEvent(t, r, events.NewTraits().WithNamespace("http"), sm)
					log := evt.Data.(*openapi.HttpLog)
					log.ClientIP = d.clientIP
				}

				getUrl := func(index int, items []any) string {
					if index >= len(items) {
						return ""
					}
					m := items[index].(map[string]any)
					d := m["data"].(map[string]any)
					return d["url"].(string)
				}

				tests := []struct {
					url   string
					check try.ResponseCondition
				}{
					{
						url: "http://foo.api/api/events?traits[namespace]=http&http[clientIP]=127.0.0.1",
						check: try.AssertBody(func(t *testing.T, body string) {
							var data map[string]any
							require.NoError(t, json.Unmarshal([]byte(body), &data))
							items := data["items"].([]any)
							require.Len(t, items, 2)
							require.Equal(t, "http://localhost/foo2", getUrl(0, items))
							require.Equal(t, "http://localhost/foo1", getUrl(1, items))
						}),
					},
					{
						url: "http://foo.api/api/events?traits[namespace]=http&http[clientIP]=192.0.1.1/24",
						check: try.AssertBody(func(t *testing.T, body string) {
							var data map[string]any
							require.NoError(t, json.Unmarshal([]byte(body), &data))
							items := data["items"].([]any)
							require.Len(t, items, 1)
							require.Equal(t, "http://localhost/foo3", getUrl(0, items))
						}),
					},
				}

				for _, d := range tests {
					try.Handler(t,
						http.MethodGet,
						d.url,
						nil,
						"",
						h,
						try.HasStatusCode(200),
						try.HasHeader("Content-Type", "application/json"),
						d.check,
					)
				}

			},
		},
		{
			name: "get specific event",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("http"))
				err := sm.Push(&eventstest.Event{Name: "foo"}, events.NewTraits().WithNamespace("http"))
				require.NoError(t, err)

				evt := sm.GetEvents(events.NewTraits())[0]
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/"+evt.Id,
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`{"id":"%v","traits":{"namespace":"http"},"data":{"name":"foo","api":""},"time":"%v"}`,
						evt.Id,
						evt.Time.Format(time.RFC3339Nano))))
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

				traits := events.NewTraits().WithNamespace("http")
				ctx, err := openapi.NewLogEventContext(r, false, traits)
				require.NoError(t, err)
				l, _ := openapi.LogEventFromContext(ctx)
				err = sm.Push(l, traits)
				require.NoError(t, err)

				evt := sm.GetEvents(events.NewTraits())[0]
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/"+evt.Id,
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.AssertBody(func(t *testing.T, body string) {
						var m map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &m))
						require.Equal(t, map[string]any{
							"actions":    any(nil),
							"api":        "",
							"deprecated": false,
							"duration":   float64(0),
							"path":       "",
							"request": map[string]any{
								"method": "get",
								"parameters": []any{
									map[string]any{
										"name":  "Foo",
										"raw":   "bar",
										"type":  "header",
										"value": "\"bar\"",
									},
								},
								"url": "http://localhost/foo",
							},
							"response": map[string]any{"size": float64(0), "statusCode": float64(0)},
							"clientIP": "192.0.2.1",
						},
							m["data"])
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

				traits := events.NewTraits().WithNamespace("http")
				ctx, err := openapi.NewLogEventContext(r, false, traits)
				require.NoError(t, err)
				l, _ := openapi.LogEventFromContext(ctx)
				err = sm.Push(l, traits)
				require.NoError(t, err)

				evt := sm.GetEvents(events.NewTraits())[0]
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/"+evt.Id,
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.AssertBody(func(t *testing.T, body string) {
						var m map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &m))
						require.Equal(t, map[string]any{
							"actions":    any(nil),
							"api":        "",
							"deprecated": false,
							"duration":   float64(0),
							"path":       "",
							"request": map[string]any{
								"method": "get",
								"parameters": []any{
									map[string]any{
										"name":  "Foo",
										"raw":   "bar",
										"type":  "header",
										"value": "{\"firstName\":\"Alex\",\"role\":\"admin\"}",
									},
								},
								"url": "http://localhost/foo",
							},
							"response": map[string]any{"size": float64(0), "statusCode": float64(0)},
							"clientIP": "192.0.2.1",
						},
							m["data"])
					}))
			},
		},
		{
			name: "event not found",
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
		{
			name: "test all events",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(2, events.NewTraits())

				newHttpEvent(t, httptest.NewRequest("get", "http://localhost/foo", nil), events.NewTraits().WithNamespace("http"), sm)
				err := sm.Push(&kafka.KafkaMessageLog{
					Topic:   "topic",
					Key:     &kafka.LogValue{Value: "foo"},
					Message: &kafka.LogValue{Value: "bar"},
				}, events.NewTraits().WithNamespace("kafka"))
				require.NoError(t, err)
				err = sm.Push(&mqtt.LogMessage{Topic: "topic", Message: &mqtt.LogValue{Value: "foo"}}, events.NewTraits().WithNamespace("kafka"))
				require.NoError(t, err)
				err = sm.Push(&websocket.MessageLog{Channel: "channel", Message: &websocket.LogValue{Value: "foo"}}, events.NewTraits().WithNamespace("kafka"))
				require.NoError(t, err)
				err = sm.Push(&directory.SearchLog{
					Request:  &directory.SearchRequest{Operation: "Search", BaseDN: "foo", Filter: "bar"},
					Response: &directory.SearchResponse{Status: "ok"},
					Duration: int64(10),
				}, events.NewTraits().WithNamespace("kafka"))
				require.NoError(t, err)

				getItemData := func(items []any, index int) map[string]any {
					if index >= len(items) {
						t.FailNow()
					}
					m := items[index].(map[string]any)
					d, ok := m["data"]
					require.True(t, ok, "data field not found for index %d", index)
					return d.(map[string]any)
				}

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.AssertBody(func(t *testing.T, body string) {
						var data map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &data))
						require.Len(t, data["items"], 5)
						items := data["items"].([]any)
						// get the inserted order not LIFO
						slices.Reverse(items)
						require.Equal(t, map[string]any{"deprecated": false, "duration": float64(0), "statusCode": float64(0), "url": "http://localhost/foo"}, getItemData(items, 0))
						require.Equal(t, map[string]any{"topic": "topic", "deleted": false, "key": map[string]any{"binary": nil, "value": "foo"}, "message": map[string]any{"binary": nil, "value": "bar"}}, getItemData(items, 1))
						require.Equal(t, map[string]any{"topic": "topic", "message": map[string]any{"binary": nil, "value": "foo"}}, getItemData(items, 2))
						require.Equal(t, map[string]any{"channel": "channel", "client": map[string]any{"address": "", "header": nil, "id": "", "query": nil, "server": ""}, "direction": "", "message": map[string]any{"binary": nil, "value": "foo"}}, getItemData(items, 3))
						require.Equal(t, map[string]any{"criteria": "bar", "dn": "foo", "duration": float64(10), "operation": "Search", "status": "ok"}, getItemData(items, 4))
					}),
				)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg, &dynamictest.Reader{})

			h := api.New(app, static.Api{})
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
					try.HasBody(`{"items":[],"page":1,"size":25,"total":0}`))
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
					try.HasBody(fmt.Sprintf(`{"items":[{"id":"%v","traits":{"namespace":"kafka"},"time":"%v","data":{"name":"foo","api":""}}],"page":1,"size":25,"total":1}`,
						event.Id,
						event.Time.Format(time.RFC3339Nano))))
			},
		},
		{
			name: "get specific event",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("kafka"))
				err := sm.Push(&eventstest.Event{Name: "foo"}, events.NewTraits().WithNamespace("kafka"))
				require.NoError(t, err)

				event := sm.GetEvents(events.NewTraits())[0]
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/"+event.Id,
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`{"id":"%v","traits":{"namespace":"kafka"},"data":{"name":"foo","api":""},"time":"%v"}`,
						event.Id,
						event.Time.Format(time.RFC3339Nano))))
			},
		},
		{
			name: "get kafka with producerId",
			fn: func(t *testing.T, h http.Handler, sm *events.StoreManager) {
				sm.SetStore(1, events.NewTraits().WithNamespace("kafka"))

				err := sm.Push(&kafka.KafkaMessageLog{
					Topic:          "foo",
					Offset:         123,
					Key:            &kafka.LogValue{},
					Message:        &kafka.LogValue{},
					MessageId:      "foo-1",
					Partition:      1,
					ProducerId:     3,
					ProducerEpoch:  1,
					SequenceNumber: 2,
					ClientId:       "foo",
					ScriptFile:     "123-123",
				}, events.NewTraits().WithNamespace("kafka"))
				require.NoError(t, err)

				event := sm.GetEvents(events.NewTraits())[0]
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/events/"+event.Id,
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.AssertBody(func(t *testing.T, body string) {
						var m map[string]any
						require.NoError(t, json.Unmarshal([]byte(body), &m))
						require.Equal(t, map[string]any{
							"topic":   "foo",
							"api":     "",
							"deleted": false,
							"headers": any(nil),
							"key": map[string]any{
								"binary": any(nil),
								"value":  "",
							},
							"message": map[string]any{
								"binary": any(nil),
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
							"actions":        nil,
						},
							m["data"])
					}))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg, &dynamictest.Reader{})

			h := api.New(app, static.Api{})
			tc.fn(t, h, app.Events)
		})
	}
}
