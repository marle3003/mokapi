package swagger

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/schema"
	"net/http"
	"testing"
)

func TestConvert(t *testing.T) {
	testcases := []struct {
		name   string
		config string
		f      func(t *testing.T, config *openapi.Config)
	}{
		{
			"only header",
			`{"swagger": "2.0"}`,
			func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "3.0.1", config.OpenApi)
			},
		},
		{
			"info",
			`{"swagger": "2.0", "info": {"title": "Foo", "description": "bar", "contact": {"name": "foobar"}, "version": "1.0"}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "Foo", config.Info.Name)
				require.Equal(t, "bar", config.Info.Description)
				require.Equal(t, "foobar", config.Info.Contact.Name)
				require.Equal(t, "1.0", config.Info.Version)
			},
		},
		{
			"server",
			`{"swagger": "2.0", "host": "server:8080", "basePath": "/api"}`,
			func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "http://server:8080/api", config.Servers[0].Url)
			},
		},
		{
			"path request",
			`{"swagger": "2.0", "paths": {"/foo": {"get": {"summary": "foo path", "description": "foo description", "tags": ["foo"], "parameters": [{"in": "query", "name": "bar", "type": "string", "description":"bar description"}]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				p := config.Paths.Value["/foo"]
				require.NotNil(t, p.Value.Get)
				get := p.Value.Get
				require.Equal(t, "foo path", get.Summary)
				require.Equal(t, "foo description", get.Description)
				require.Equal(t, []string{"foo"}, get.Tags)
				require.Equal(t, parameter.Query, get.Parameters[0].Value.Type)
				require.Equal(t, "bar", get.Parameters[0].Value.Name)
				require.Equal(t, "string", get.Parameters[0].Value.Schema.Value.Type)
				require.Equal(t, "bar description", get.Parameters[0].Value.Description)
			},
		},
		{
			"path response",
			`{"swagger": "2.0", "paths": {"/foo": {"get": {"produces": ["application/json"], "responses": {"200": {"description": "response description", "schema": {"$ref": "#/definitions/foo"}}}}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				p := config.Paths.Value["/foo"]
				require.NotNil(t, p.Value.Get)
				get := p.Value.Get
				ok := get.Responses.GetResponse(http.StatusOK)
				require.NotNil(t, ok)
				require.Equal(t, "response description", ok.Description)
				require.Equal(t, "#/components/schemas/foo", ok.Content["application/json"].Schema.Ref)
			},
		},
		{
			"definitions",
			`{"swagger": "2.0", "definitions": {"Foo": {"type": "string"}, "Bar": {"type": "object","properties": {"title":{"type": "string"}}}}}`,
			func(t *testing.T, config *openapi.Config) {
				foo := config.Components.Schemas.Value.Get("Foo").(*schema.Ref)
				require.Equal(t, "string", foo.Value.Type)
				bar := config.Components.Schemas.Value.Get("Bar").(*schema.Ref)
				require.Equal(t, "object", bar.Value.Type)
				title := bar.Value.Properties.Value.Get("title").(*schema.Ref)
				require.Equal(t, "string", title.Value.Type)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Config{}
			err := json.Unmarshal([]byte(tc.config), &c)
			require.NoError(t, err)
			converted, err := Convert(c)
			require.NoError(t, err)
			tc.f(t, converted)
		})
	}
}
