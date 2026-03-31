package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/directory"
	"mokapi/providers/mail"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_GetApiSpec(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "List APIs skip empty name",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0"),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{})
				require.NoError(t, err)
				require.Len(t, r.Apis, 0)
			},
		},
		{
			name: "List APIs with HTTP",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{})
				require.NoError(t, err)
				require.Len(t, r.Apis, 1)
				require.Equal(t, "foo", r.Apis[0].Name)
				require.Equal(t, "http", r.Apis[0].Type)
			},
		},
		{
			name: "List APIs with Kafka",
			app: runtimetest.NewApp(
				runtimetest.WithKafka(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "", ""),
				)),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{})
				require.NoError(t, err)
				require.Len(t, r.Apis, 1)
				require.Equal(t, "foo", r.Apis[0].Name)
				require.Equal(t, "kafka", r.Apis[0].Type)
			},
		},
		{
			name: "List APIs with Kafka and HTTP",
			app: runtimetest.NewApp(
				runtimetest.WithHttp(openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
				)),
				runtimetest.WithKafka(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("bar", "", ""),
				)),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{})
				require.NoError(t, err)
				require.Len(t, r.Apis, 2)
				require.Equal(t, "foo", r.Apis[0].Name)
				require.Equal(t, "http", r.Apis[0].Type)
				require.Equal(t, "bar", r.Apis[1].Name)
				require.Equal(t, "kafka", r.Apis[1].Type)
			},
		},
		{
			name: "List APIs with Kafka and HTTP using filter",
			app: runtimetest.NewApp(
				runtimetest.WithHttp(openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
				)),
				runtimetest.WithKafka(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("bar", "", ""),
				)),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{Type: "kafka"})
				require.NoError(t, err)
				require.Len(t, r.Apis, 1)
				require.Equal(t, "bar", r.Apis[0].Name)
				require.Equal(t, "kafka", r.Apis[0].Type)
			},
		},
		{
			name: "List APIs with LDAP",
			app: runtimetest.NewApp(
				runtimetest.WithLdap(&directory.Config{
					Info: directory.Info{Name: "foo"},
				}),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{})
				require.NoError(t, err)
				require.Len(t, r.Apis, 1)
				require.Equal(t, "foo", r.Apis[0].Name)
				require.Equal(t, "ldap", r.Apis[0].Type)
			},
		},
		{
			name: "List APIs with Mail",
			app: runtimetest.NewApp(
				runtimetest.WithMail(&mail.Config{
					Info: mail.Info{Name: "foo"},
				}),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{})
				require.NoError(t, err)
				require.Len(t, r.Apis, 1)
				require.Equal(t, "foo", r.Apis[0].Name)
				require.Equal(t, "mail", r.Apis[0].Type)
			},
		},
		{
			name: "Get HTTP API spec",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{Name: "foo", Type: "http"})
				require.NoError(t, err)
				require.IsType(t, &openapi.Config{}, r.Apis[0].Spec)
				require.Equal(t, "foo", r.Apis[0].Spec.(*openapi.Config).Info.Name)
			},
		},
		{
			name: "Get HTTP API spec name does not exist",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{Name: "bar", Type: "http"})
				require.NoError(t, err)
				require.Len(t, r.Apis, 0)
			},
		},
		{
			name: "Get Kafka API spec",
			app: runtimetest.NewApp(
				runtimetest.WithKafka(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "", ""),
				)),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{Name: "foo", Type: "kafka"})
				require.NoError(t, err)
				require.IsType(t, &asyncapi3.Config{}, r.Apis[0].Spec)
				require.Equal(t, "foo", r.Apis[0].Spec.(*asyncapi3.Config).Info.Name)
			},
		},
		{
			name: "Get Mail API spec",
			app: runtimetest.NewApp(
				runtimetest.WithLdap(&directory.Config{Info: directory.Info{Name: "foo"}}),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{Name: "foo", Type: "ldap"})
				require.NoError(t, err)
				require.IsType(t, &directory.Config{}, r.Apis[0].Spec)
				require.Equal(t, "foo", r.Apis[0].Spec.(*directory.Config).Info.Name)
			},
		},
		{
			name: "Get Mail API spec",
			app: runtimetest.NewApp(
				runtimetest.WithMail(&mail.Config{Info: mail.Info{Name: "foo"}}),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{Name: "foo", Type: "mail"})
				require.NoError(t, err)
				require.IsType(t, &mail.Config{}, r.Apis[0].Spec)
				require.Equal(t, "foo", r.Apis[0].Spec.(*mail.Config).Info.Name)
			},
		},
		{
			name: "Get API spec unknown type",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				_, err := s.GetApiSpec(context.Background(), mcp.GetApiSpecInput{Name: "bar", Type: "unknown"})
				require.EqualError(t, err, "unknown type: unknown")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := mcp.NewService(tc.app)
			tc.test(t, s)
		})
	}
}
