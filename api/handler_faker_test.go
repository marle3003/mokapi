package api

import (
	"encoding/json"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema/schematest"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Faker(t *testing.T) {
	testcases := []struct {
		name          string
		app           func() *runtime.App
		requestUrl    string
		requestMethod string
		requestBody   string
		test          []try.ResponseCondition
	}{
		{
			name: "get tree",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl: "http://foo.api/api/faker/tree",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.BodyContains(`{"name":"root","path":"/","children":[`),
			},
		},
		{
			name: "run fake on node",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl: "http://foo.api/api/faker/node/firstname/fake",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.HasBody(`"Zoe"`),
			},
		},
		{
			name: "run fake on node with schema in body",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl:    "http://foo.api/api/faker/node/zip/fake",
			requestMethod: "POST",
			requestBody: func() string {
				b, err := json.Marshal(
					schematest.New("number",
						schematest.WithMinimum(1000),
						schematest.WithMaximum(9999),
					),
				)
				if err != nil {
					panic(err)
				}
				return string(b)
			}(),
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.HasBody(`8703`),
			},
		},
		{
			name: "run fake on node with invalid schema",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl:    "http://foo.api/api/faker/node/zip/fake",
			requestMethod: "POST",
			requestBody:   `{"type": 123}`,
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusBadRequest),
				try.HasBody("schema error at field 'type': expected type string or array, got number"),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(12345)

			h := New(tc.app(), static.Api{})
			method := tc.requestMethod
			if method == "" {
				method = http.MethodGet
			}

			try.Handler(t,
				method,
				tc.requestUrl,
				nil,
				tc.requestBody,
				h,
				tc.test...)
		})
	}
}
