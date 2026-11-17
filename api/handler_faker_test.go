package api

import (
	"encoding/json"
	"errors"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema/schematest"
	"mokapi/try"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func addFakerNode(t *testing.T, n *generator.Node) {
	t.Helper()
	root := generator.FindByName(generator.RootName)
	root.Append(n)
	t.Cleanup(func() {
		require.NoError(t, root.Remove(n.Name))
	})
}

// addFakerBranch adds a grouping node with a child "code", so that the same name exists in several branches.
func addFakerBranch(t *testing.T, name string, code string) {
	t.Helper()
	branch := generator.NewNode(name)
	branch.Append(&generator.Node{
		Name: "code",
		Fake: func(r *generator.Request) (any, error) { return code, nil },
	})
	addFakerNode(t, branch)
}

func TestHandler_Faker(t *testing.T) {
	testcases := []struct {
		name          string
		app           func() *runtime.App
		requestUrl    string
		requestMethod string
		requestBody   string
		setup         func(t *testing.T)
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
			requestUrl: "http://foo.api/api/faker/node/fake?path=/person/firstname",
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
			requestUrl:    "http://foo.api/api/faker/node/fake?path=/zip",
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
			requestUrl:    "http://foo.api/api/faker/node/fake?path=/zip",
			requestMethod: "POST",
			requestBody:   `{"type": 123}`,
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusBadRequest),
				try.HasBody("schema error at field 'type': expected type string or array, got number"),
			},
		},
		{
			name: "run fake on node with the same name in two branches",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			setup: func(t *testing.T) {
				addFakerBranch(t, "billing", "BILL-1")
				addFakerBranch(t, "shipping", "SHIP-1")
			},
			requestUrl: "http://foo.api/api/faker/node/fake?path=/shipping/code",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.HasBody(`"SHIP-1"`),
			},
		},
		{
			name: "run fake on unknown node",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl: "http://foo.api/api/faker/node/fake?path=/person/unknown",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusNotFound),
			},
		},
		{
			name: "run fake on node without path",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl: "http://foo.api/api/faker/node/fake",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusNotFound),
			},
		},
		{
			name: "run fake on root",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl: "http://foo.api/api/faker/node/fake?path=/",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusBadRequest),
				try.HasBody("node / has no fake function"),
			},
		},
		{
			name: "run fake on grouping node",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			setup: func(t *testing.T) {
				addFakerBranch(t, "billing", "BILL-1")
			},
			requestUrl: "http://foo.api/api/faker/node/fake?path=/billing",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusBadRequest),
				try.HasBody("node /billing has no fake function"),
			},
		},
		{
			name: "run fake on node with failing fake function",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			setup: func(t *testing.T) {
				addFakerNode(t, &generator.Node{
					Name: "failing",
					Fake: func(r *generator.Request) (any, error) { return nil, errors.New("boom") },
				})
			},
			requestUrl: "http://foo.api/api/faker/node/fake?path=/failing",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusBadRequest),
				try.HasBody("boom"),
			},
		},
		{
			name: "explain property path",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl:    "http://foo.api/api/faker/fake",
			requestMethod: "POST",
			requestBody:   `{"path": ["customerEmail"]}`,
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.HasHeader("Content-Type", "application/json"),
				try.AssertBody(func(t *testing.T, body string) {
					var result generator.Result
					require.NoError(t, json.Unmarshal([]byte(body), &result))
					require.Len(t, result.Matches, 1)
					m := result.Matches[0]
					require.Equal(t, "customerEmail", m.PropertyPath)
					require.Equal(t, "/email", m.Node)
					require.Equal(t, []string{"email"}, m.Tokens)
					require.Equal(t, "node", m.Source)
					require.Equal(t, result.Value, m.Value)
					require.Contains(t, body, `"propertyPath":"customerEmail","node":"/email","tokens":["email"],"source":"node","value":`)
				}),
			},
		},
		{
			name: "explain schema with nested property",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl:    "http://foo.api/api/faker/fake",
			requestMethod: "POST",
			requestBody:   `{"schema": {"type": "object", "required": ["address"], "properties": {"address": {"type": "object", "required": ["city"], "properties": {"city": {"type": "string"}}}}}}`,
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.AssertBody(func(t *testing.T, body string) {
					var result generator.Result
					require.NoError(t, json.Unmarshal([]byte(body), &result))
					require.Len(t, result.Matches, 1)
					require.Equal(t, "address.city", result.Matches[0].PropertyPath)
				}),
			},
		},
		{
			name: "explain without body generates a value",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl:    "http://foo.api/api/faker/fake",
			requestMethod: "POST",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.BodyContains(`"matches":[`),
			},
		},
		{
			name: "explain value without match rows has an empty list",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl:    "http://foo.api/api/faker/fake",
			requestMethod: "POST",
			requestBody:   `{"schema": {"type": "array", "items": {"type": "string"}, "maxItems": 0}}`,
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.HasBody(`{"value":[],"matches":[]}`),
			},
		},
		{
			name: "explain with invalid json",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl:    "http://foo.api/api/faker/fake",
			requestMethod: "POST",
			requestBody:   `{"path": `,
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusBadRequest),
				try.BodyContains("unexpected end of JSON input"),
			},
		},
		{
			name: "explain with invalid schema",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl:    "http://foo.api/api/faker/fake",
			requestMethod: "POST",
			requestBody:   `{"schema": {"type": 123}}`,
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusBadRequest),
				try.HasBody("schema error at field 'type': expected type string or array, got number"),
			},
		},
		{
			name: "explain reports generator error",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl:    "http://foo.api/api/faker/fake",
			requestMethod: "POST",
			requestBody:   `{"schema": {"type": "array", "minItems": 3, "maxItems": 1}}`,
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusBadRequest),
				try.BodyContains("minItems must be less than maxItems"),
			},
		},
		{
			name: "old route with node name is removed",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl: "http://foo.api/api/faker/node/firstname/fake",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusNotFound),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(12345)
			if tc.setup != nil {
				tc.setup(t)
			}

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
