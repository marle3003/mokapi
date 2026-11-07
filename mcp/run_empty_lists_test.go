package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/directory"
	"mokapi/providers/mail"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// A list function without results must return an empty JavaScript array
// and not null, so that scripts can call length, map or find on it.
func TestService_Run_EmptyLists(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		code string
	}{
		{
			name: "mokapi.getApis",
			app:  runtimetest.NewApp(),
			code: `mokapi.getApis()`,
		},
		{
			name: "mokapi.getApis skip empty name",
			app:  runtimetest.NewHttpApp(openapitest.NewConfig("3.1.0")),
			code: `mokapi.getApis()`,
		},
		{
			name: "mokapi.getEvents",
			app:  runtimetest.NewApp(),
			code: `mokapi.getEvents()`,
		},
		{
			name: "Kafka getTopics",
			app: runtimetest.NewKafkaApp(
				asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", "")),
			),
			code: `mokapi.getApi('foo').getTopics()`,
		},
		{
			name: "Kafka topic consume",
			app: runtimetest.NewKafkaApp(
				asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("foo", "", ""),
					asyncapi3test.WithChannel("channel-1"),
				),
			),
			code: `mokapi.getApi('foo').getTopic('channel-1').consume(0, 0, 10)`,
		},
		{
			name: "LDAP getEntries without entries",
			app: runtimetest.NewApp(
				runtimetest.WithLdap(&directory.Config{Info: directory.Info{Name: "foo"}}),
			),
			code: `mokapi.getApi('foo').getEntries()`,
		},
		{
			name: "Mail getMailboxes",
			app: runtimetest.NewApp(
				runtimetest.WithMail(&mail.Config{Version: "1.0", Info: mail.Info{Name: "foo"}}),
			),
			code: `mokapi.getApi('foo').getMailboxes()`,
		},
		{
			name: "HTTP getOperations without paths",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0", openapitest.WithInfo("foo", "", "")),
			),
			code: `mokapi.getApi('foo').getOperations()`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := mcp.NewService(tc.app)

			r, err := s.GetRunResponse(context.Background(), mcp.RunInput{Code: tc.code})
			require.NoError(t, err)
			require.NotNil(t, r.Result, "expected empty array but got null")
			require.Equal(t, reflect.Slice, reflect.TypeOf(r.Result).Kind())
			require.Equal(t, 0, reflect.ValueOf(r.Result).Len())

			r, err = s.GetRunResponse(context.Background(), mcp.RunInput{Code: `(function() { var r = ` + tc.code + `; return r !== null && r.length === 0; })()`})
			require.NoError(t, err)
			require.Equal(t, true, r.Result)
		})
	}
}
