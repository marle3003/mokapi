package mcp_test

import (
	"context"
	"encoding/json"
	"mokapi/mcp"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/generator"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_Run(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "run JavaScript code",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0"),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `1+1`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, int64(2), r.Result)
			},
		},
		{
			name: "List APIs skip empty name",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0"),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApis()`,
					},
				)
				require.NoError(t, err)
				require.Len(t, r.Result, 0)
			},
		},
		{
			name: "get HTTP APIs",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
				),
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("bar", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApis()`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, []mcp.ApiSummary{
					{Name: "bar", Type: "http"},
					{Name: "foo", Type: "http"},
				}, r.Result)
			},
		},
		{
			name: "get specific HTTP API",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
				),
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("bar", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo')`,
					},
				)
				require.NoError(t, err)
				b, err := json.Marshal(r.Result)
				require.NoError(t, err)
				require.Equal(t, `{"name":"foo","type":"http"}`, string(b))
			},
		},
		{
			name: "get API's operation list",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithOperationSummary("GET summary"),
						),
						openapitest.WithOperation(http.MethodPut,
							openapitest.WithOperationSummary("PUT summary"),
						),
					),
					openapitest.WithPath("/users",
						openapitest.WithPathInfo("path summary", ""),
						openapitest.WithOperation(http.MethodPost),
					),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo').getOperations()`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, []mcp.OperationSummary{
					{Method: "GET", Path: "/pets", Summary: "GET summary"},
					{Method: "PUT", Path: "/pets", Summary: "PUT summary"},
					{Method: "POST", Path: "/users", Summary: "path summary"}},
					r.Result)
			},
		},
		{
			name: "get API's operation details",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithOperationSummary("GET summary"),
							openapitest.WithHeaderParam("foo", false, openapitest.WithParamSchema(schematest.New("string"))),
							openapitest.WithRequestBody(
								"request body description",
								true,
								openapitest.WithRequestContent(
									"application/json",
									&openapi.MediaType{
										Schema: schematest.New("string"),
									},
								),
							),
							openapitest.WithResponse(200,
								openapitest.WithContent("application/json", openapitest.WithSchema(
									schematest.New("string"),
								)),
							),
						),
						openapitest.WithOperation(http.MethodPut,
							openapitest.WithOperationSummary("PUT summary"),
						),
					),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo').getOperationDetails('/pets', 'GET')`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, &mcp.OperationDetails{}, r.Result)
				op := r.Result.(*mcp.OperationDetails)
				require.Equal(t, "", op.OperationId)
				require.Equal(t, "GET", op.Method)
				require.Equal(t, "/pets", op.Path)
				require.Equal(t, "GET summary", op.Summary)
				require.Equal(t, "", op.Description)
				require.Equal(t, "foo", op.Parameters[0].Name)
				require.Equal(t, "header", op.Parameters[0].In)
				require.Equal(t, false, op.Parameters[0].Required)
				require.Equal(t, "string", op.Parameters[0].Schema.Type[0])
				require.Equal(t, "", op.Parameters[0].Description)
				require.Equal(t, "request body description", op.RequestBody.Description)
				require.Equal(t, true, op.RequestBody.Required)
				require.Equal(t, "application/json", op.RequestBody.Contents[0].ContentType)
				require.Equal(t, "string", op.RequestBody.Contents[0].Schema.Type[0])
			},
		},
		{
			name: "getResponseSchema",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithResponse(200,
								openapitest.WithResponseDescription("response description"),
								openapitest.WithContent("application/json", openapitest.WithSchema(
									schematest.New("string"),
								)),
							),
						),
					),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const op = mokapi.getApi('foo').getOperationDetails('/pets', 'GET')
op.getResponseSchema(200)`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, &mcp.Response{}, r.Result)
				res := r.Result.(*mcp.Response)
				require.Equal(t, 200, res.StatusCode)
				require.Equal(t, "response description", res.Description)
				require.Equal(t, "application/json", res.Content[0].ContentType)
				require.Equal(t, "string", res.Content[0].Schema.Type[0])
			},
		},
		{
			name: "getResponseSchema not exists",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithResponse(200,
								openapitest.WithResponseDescription("response description"),
								openapitest.WithContent("application/json", openapitest.WithSchema(
									schematest.New("string"),
								)),
							),
						),
					),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const op = mokapi.getApi('foo').getOperationDetails('/pets', 'GET')
op.getResponseSchema(404)`,
					},
				)
				require.NoError(t, err)
				require.Nil(t, r.Result)
			},
		},
		{
			name: "invoke request",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithResponse(200,
								openapitest.WithResponseDescription("response description"),
								openapitest.WithContent("application/json", openapitest.WithSchema(
									schematest.New("string"),
								)),
							),
						),
					),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const op = mokapi.getApi('foo').getOperationDetails('/pets', 'GET');
op.invoke()`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, mcp.InvokeResponse{}, r.Result)
				res := r.Result.(mcp.InvokeResponse)
				require.Equal(t, 200, res.StatusCode)
				require.Equal(t, `"P8"`, res.Body)
				require.Equal(t, map[string][]string{
					"Content-Type": {"application/json"},
				}, res.Headers)
			},
		},
		{
			name: "invoke request with path parameter",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/{foo}/{bar}/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithOperationParam("foo", true, openapitest.WithParamSchema(schematest.New("string"))),
							openapitest.WithOperationParam("bar", true, openapitest.WithParamSchema(schematest.New("string"))),
							openapitest.WithResponse(200,
								openapitest.WithResponseDescription("response description"),
								openapitest.WithContent("application/json", openapitest.WithSchema(
									schematest.New("string"),
								)),
							),
						),
					),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const op = mokapi.getApi('foo').getOperationDetails('/{foo}/{bar}/pets', 'GET');
op.invoke({ path: { foo: 'val1', 'bar': 'val2' }})`,
					},
				)
				require.NoError(t, err)
				res := r.Result.(mcp.InvokeResponse)
				require.Equal(t, http.StatusOK, res.StatusCode, res.Body)
			},
		},
		{
			name: "invoke request with path parameter but not specified",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/{foo}/{bar}/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithOperationParam("foo", true, openapitest.WithParamSchema(schematest.New("string"))),
							openapitest.WithOperationParam("bar", true, openapitest.WithParamSchema(schematest.New("string"))),
							openapitest.WithResponse(200,
								openapitest.WithResponseDescription("response description"),
								openapitest.WithContent("application/json", openapitest.WithSchema(
									schematest.New("string"),
								)),
							),
						),
					),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				_, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const op = mokapi.getApi('foo').getOperationDetails('/{foo}/{bar}/pets', 'GET');
op.invoke()`,
					},
				)
				require.EqualError(t, err, "GoError: invoke request GET /{foo}/{bar}/pets failed: missing path parameter 'foo' at reflect.methodValueCall (native)")
			},
		},
		{
			name: "invoke request with query parameter",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithQueryParam("foo", true, openapitest.WithParamSchema(schematest.New("string"))),
							openapitest.WithResponse(200,
								openapitest.WithResponseDescription("response description"),
								openapitest.WithContent("application/json", openapitest.WithSchema(
									schematest.New("string"),
								)),
							),
						),
					),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const op = mokapi.getApi('foo').getOperationDetails('/pets', 'GET');
op.invoke({ query: { foo: 'val1' }})`,
					},
				)
				require.NoError(t, err)
				res := r.Result.(mcp.InvokeResponse)
				require.Equal(t, http.StatusOK, res.StatusCode, res.Body)
			},
		},
		{
			name: "invoke request with header parameter",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithHeaderParam("foo", true, openapitest.WithParamSchema(schematest.New("string"))),
							openapitest.WithResponse(200,
								openapitest.WithResponseDescription("response description"),
								openapitest.WithContent("application/json", openapitest.WithSchema(
									schematest.New("string"),
								)),
							),
						),
					),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const op = mokapi.getApi('foo').getOperationDetails('/pets', 'GET');
op.invoke({ header: { foo: ['val1'] }})`,
					},
				)
				require.NoError(t, err)
				res := r.Result.(mcp.InvokeResponse)
				require.Equal(t, http.StatusOK, res.StatusCode, res.Body)
			},
		},
		{
			name: "fake",
			app:  runtimetest.NewApp(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.fake({ type: 'string', format: 'email' })`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, "ivyjones@ziemann.com", r.Result)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(123456)

			s := mcp.NewService(tc.app)
			tc.test(t, s)
		})
	}
}
