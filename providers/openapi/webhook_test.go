package openapi_test

import (
	"fmt"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandler_Webhook(t *testing.T) {
	addScript := func(e *engine.Engine, src string, args ...any) error {
		src = fmt.Sprintf(src, args...)
		return e.AddScript(newScript(fmt.Sprintf("%s.ts", t.Name()), src))
	}

	setup := func(opts ...runtimetest.Options) (*runtime.App, *engine.Engine) {
		app := runtimetest.NewApp(opts...)
		e := enginetest.NewEngine(engine.WithApp(app))
		app.Engine = e
		return app, e
	}

	type Arg struct {
		Result any
	}

	testcases := []struct {
		name    string
		handler http.Handler
		test    func(t *testing.T, url string)
	}{
		{
			name: "webhook not found",
			test: func(t *testing.T, url string) {
				_, e := setup()
				err := addScript(e, `
					import { webhook } from 'mokapi'
					export default function() {
						webhook('foo', '%s')
					}
				`, url)
				require.EqualError(t, err, "webhook not found: foo")
			},
		},
		{
			name: "no operation specified",
			test: func(t *testing.T, url string) {
				_, e := setup(
					runtimetest.WithHttp(
						openapitest.NewConfig("3.1.0",
							openapitest.WithWebhook("foo"),
						),
					),
				)

				err := addScript(e, `
					import { webhook } from 'mokapi'
					export default function() {
						webhook('foo', '%s')
					}
				`, url)
				require.EqualError(t, err, "webhook foo failed: no operations specified")
			},
		},
		{
			name: "ambiguous method",
			test: func(t *testing.T, url string) {
				_, e := setup(
					runtimetest.WithHttp(
						openapitest.NewConfig("3.1.0",
							openapitest.WithWebhook("foo",
								openapitest.WithOperation(http.MethodGet),
								openapitest.WithOperation(http.MethodPost),
							),
						),
					),
				)

				err := addScript(e, `
					import { webhook } from 'mokapi'
					export default function() {
						webhook('foo', '%s')
					}
				`, url)
				require.EqualError(t, err, "webhook foo failed: multiple operations specified: use args.method to refine")
			},
		},
		{
			name: "set method but request body is required",
			test: func(t *testing.T, url string) {
				_, e := setup(
					runtimetest.WithHttp(
						openapitest.NewConfig("3.1.0",
							openapitest.WithWebhook("foo",
								openapitest.WithOperation(http.MethodGet),
								openapitest.WithOperation(http.MethodPost, openapitest.WithRequestBody("", true)),
							),
						),
					),
				)

				err := addScript(e, `
					import { webhook } from 'mokapi'
					export default function() {
						webhook('foo', '%s', { method: 'post' })
					}
				`, url)
				require.EqualError(t, err, "webhook foo failed: request body is required")
			},
		},
		{
			name: "invoke webhook with request body",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			test: func(t *testing.T, url string) {
				_, e := setup(
					runtimetest.WithHttp(
						openapitest.NewConfig("3.1.0",
							openapitest.WithWebhook("foo",
								openapitest.WithOperation(http.MethodPost,
									openapitest.WithRequestBody("", true, openapitest.WithRequestContent("application/json",
										openapitest.WithSchema(schematest.New("object",
											schematest.WithProperty("foo", schematest.New("string")),
										)),
									)),
									openapitest.WithResponse(http.StatusOK),
								),
							),
						),
					),
				)

				err := addScript(e, `
					import { on, webhook } from 'mokapi'
					export default function() {
						const res = webhook('foo', '%s', { data: { foo: 'bar' } })
						on('debug', (arg) => arg.result = res)
					}
				`, url)
				arg := Arg{}
				e.Run("debug", &arg)
				require.NoError(t, err)
				require.Equal(t, &common.WebhookResponse{
					StatusCode: http.StatusOK,
					Headers:    map[string]any{},
				}, arg.Result)
			},
		},
		{
			name: "request body validation error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			test: func(t *testing.T, url string) {
				_, e := setup(
					runtimetest.WithHttp(
						openapitest.NewConfig("3.1.0",
							openapitest.WithWebhook("foo",
								openapitest.WithOperation(http.MethodPost,
									openapitest.WithRequestBody("", true, openapitest.WithRequestContent("application/json",
										openapitest.WithSchema(schematest.New("object",
											schematest.WithProperty("foo", schematest.New("string")),
										)),
									)),
									openapitest.WithResponse(http.StatusOK),
								),
							),
						),
					),
				)

				err := addScript(e, `
					import { on, webhook } from 'mokapi'
					export default function() {
						webhook('foo', '%s', { data: { foo: 123 } })
					}
				`, url)
				arg := Arg{}
				e.Run("debug", &arg)
				require.EqualError(t, err, "webhook foo failed: Validation error count 1:\n\t- #/foo/type: invalid type, expected string but got integer")
			},
		},
		{
			name: "response with body",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{ "foo": "bar" }`))
			}),
			test: func(t *testing.T, url string) {
				_, e := setup(
					runtimetest.WithHttp(
						openapitest.NewConfig("3.1.0",
							openapitest.WithWebhook("foo",
								openapitest.WithOperation(http.MethodPost,
									openapitest.WithRequestBody("", true, openapitest.WithRequestContent("application/json",
										openapitest.WithSchema(schematest.New("object",
											schematest.WithProperty("foo", schematest.New("string")),
										)),
									)),
									openapitest.WithResponse(http.StatusOK,
										openapitest.WithContent("application/json",
											openapitest.WithSchema(
												schematest.New("object",
													schematest.WithProperty("foo", schematest.New("string")),
												),
											),
										),
									),
								),
							),
						),
					),
				)

				err := addScript(e, `
					import { on, webhook } from 'mokapi'
					export default function() {
						const res = webhook('foo', '%s', { data: { foo: 'bar' } })
						on('debug', (arg) => arg.result = res)
					}
				`, url)
				arg := Arg{}
				e.Run("debug", &arg)
				require.NoError(t, err)
				require.Equal(t, &common.WebhookResponse{
					StatusCode: http.StatusOK,
					Data:       map[string]interface{}{"foo": "bar"},
					Headers:    map[string]any{"Content-Type": "application/json"},
				}, arg.Result)
			},
		},
		{
			name: "request header missing",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			test: func(t *testing.T, url string) {
				_, e := setup(
					runtimetest.WithHttp(
						openapitest.NewConfig("3.1.0",
							openapitest.WithWebhook("foo",
								openapitest.WithOperation(http.MethodPost,
									openapitest.WithHeaderParam("foo", true),
									openapitest.WithResponse(http.StatusOK),
								),
							),
						),
					),
				)

				err := addScript(e, `
					import { on, webhook } from 'mokapi'
					export default function() {
						webhook('foo', '%s', {  })
					}
				`, url)
				arg := Arg{}
				e.Run("debug", &arg)
				require.EqualError(t, err, "webhook foo failed: required header parameter foo not found")
			},
		},
		{
			name: "request header validation error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			test: func(t *testing.T, url string) {
				_, e := setup(
					runtimetest.WithHttp(
						openapitest.NewConfig("3.1.0",
							openapitest.WithWebhook("foo",
								openapitest.WithOperation(http.MethodPost,
									openapitest.WithHeaderParam("foo", true, openapitest.WithParamSchema(schematest.New("integer"))),
									openapitest.WithResponse(http.StatusOK),
								),
							),
						),
					),
				)

				err := addScript(e, `
					import { on, webhook } from 'mokapi'
					export default function() {
						webhook('foo', '%s', { headers: { foo: 'bar' } })
					}
				`, url)
				arg := Arg{}
				e.Run("debug", &arg)
				require.EqualError(t, err, "webhook foo failed: failed to parse header parameter foo: Validation error count 1:\n\t- #/type: invalid type, expected integer but got string")
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(tc.handler)

			tc.test(t, s.URL)
		})
	}
}
