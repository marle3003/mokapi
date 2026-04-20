package mcp_test

import (
	"context"
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

func TestService_Run_Http(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
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
					openapitest.WithServer("http://localhost/foo", "server description"),
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
				require.IsType(t, &mcp.OpenAPI{}, r.Result)
				api := r.Result.(*mcp.OpenAPI)
				require.Equal(t, "foo", api.Name)
				require.Equal(t, "http", api.Type)
				require.Equal(t, []mcp.OpenAPIServer{{Url: "http://localhost/foo", Description: "server description"}}, api.Servers)
			},
		},
		{
			name: "get API's operation list",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithOperationId("pets"),
							openapitest.WithOperationSummary("GET summary"),
						),
						openapitest.WithOperation(http.MethodPut,
							openapitest.WithOperationSummary("PUT summary"),
						),
					),
					openapitest.WithPath("/users",
						openapitest.WithPathInfo("path summary", ""),
						openapitest.WithPathParam("foo"),
						openapitest.WithOperation(http.MethodPost, openapitest.WithOperationParam("bar", false)),
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
					{Id: "pets", Method: "GET", Path: "/pets", Summary: "GET summary"},
					{Id: "put-/pets", Method: "PUT", Path: "/pets", Summary: "PUT summary"},
					{Id: "post-/users", Method: "POST", Path: "/users", Summary: "path summary", Parameters: []string{"bar", "foo"}}},
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
								openapitest.WithResponseDescription("response description"),
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
						Code: `mokapi.getApi('foo').getOperation('get-/pets')`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, &mcp.Operation{}, r.Result)
				op := r.Result.(*mcp.Operation)
				require.Equal(t, "get-/pets", op.OperationId)
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

				require.Len(t, op.Responses, 1)
				require.Equal(t, http.StatusOK, op.Responses[0].StatusCode)
				require.Equal(t, "response description", op.Responses[0].Description)
				require.Len(t, op.Responses[0].Content, 1)
				require.Equal(t, "application/json", op.Responses[0].Content[0].ContentType)
				require.Equal(t, "string", op.Responses[0].Content[0].Schema.Type[0])
			},
		},
		{
			name: "get API's operation using projection",
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
						Code: `const op = mokapi.getApi('foo').getOperation('get-/pets')
const result = { path: op.path, method: op.method }; result`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, map[string]any{"method": "GET", "path": "/pets"}, r.Result)
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
						Code: `const op = mokapi.getApi('foo').getOperation('get-/pets')
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
									schematest.New("object",
										schematest.WithProperty("foo", schematest.New("string")),
										schematest.WithProperty("bar", schematest.New("integer")),
										schematest.WithRequired("foo", "bar"),
									),
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
						Code: `const op = mokapi.getApi('foo').getOperation('get-/pets');
op.invoke()`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, mcp.InvokeResponse{}, r.Result)
				res := r.Result.(mcp.InvokeResponse)
				require.Equal(t, 200, res.StatusCode)
				require.Equal(t, `{"foo":"P8","bar":-804702}`, res.Body)
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
						Code: `const op = mokapi.getApi('foo').getOperation('get-/{foo}/{bar}/pets');
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
						Code: `const op = mokapi.getApi('foo').getOperation('get-/{foo}/{bar}/pets');
op.invoke()`,
					},
				)
				require.EqualError(t, err, "invoke request GET /{foo}/{bar}/pets failed: missing path parameter 'foo'")
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
						Code: `const op = mokapi.getApi('foo').getOperation('get-/pets');
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
						Code: `const op = mokapi.getApi('foo').getOperation('get-/pets');
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
