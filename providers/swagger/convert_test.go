package swagger

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/parameter"
	"mokapi/version"
	"net/http"
	"testing"
)

func TestConvert(t *testing.T) {
	testcases := []struct {
		name   string
		config string
		test   func(t *testing.T, config *openapi.Config)
	}{
		{
			name:   "header",
			config: `{"swagger": "2.0"}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Equal(t, version.New("3.0.3"), config.OpenApi)
			},
		},
		{
			name:   "info title",
			config: `{"swagger": "2.0", "info": {"title": "foo"}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "foo", config.Info.Name)
			},
		},
		{
			name:   "info description",
			config: `{"swagger": "2.0", "info": {"description": "foo"}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "foo", config.Info.Description)
			},
		},
		{
			name:   "contact",
			config: `{"swagger": "2.0", "info": {"contact": {"name": "foo","url":"http://foo.bar","email":"foo@bar.com"}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "foo", config.Info.Contact.Name)
				require.Equal(t, "http://foo.bar", config.Info.Contact.Url)
				require.Equal(t, "foo@bar.com", config.Info.Contact.Email)
			},
		},
		{
			name:   "version",
			config: `{"swagger": "2.0", "info": {"version": "1.0"}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "1.0", config.Info.Version)
			},
		},
		{
			name:   "host",
			config: `{"swagger": "2.0", "host": "server:8080"}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "http://server:8080", config.Servers[0].Url)
			},
		},
		{
			name:   "basePath",
			config: `{"swagger": "2.0", "basePath": "/foo"}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "/foo", config.Servers[0].Url)
			},
		},
		{
			name:   "host",
			config: `{"swagger": "2.0", "host": "foo"}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "http://foo", config.Servers[0].Url)
			},
		},
		{
			name:   "host with port",
			config: `{"swagger": "2.0", "host": "foo:8080"}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "http://foo:8080", config.Servers[0].Url)
			},
		},
		{
			name:   "host with port and basePath",
			config: `{"swagger": "2.0", "host": "foo:8080", "basePath": "/bar"}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "http://foo:8080/bar", config.Servers[0].Url)
			},
		},
		{
			name:   "scheme with basePath",
			config: `{"swagger": "2.0","schemes":["https"],"basePath": "/bar"}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "https:///bar", config.Servers[0].Url)
			},
		},
		{
			name:   "scheme with host and basePath",
			config: `{"swagger": "2.0","schemes":["https"],"host":"foo","basePath": "/bar"}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "https://foo/bar", config.Servers[0].Url)
			},
		},
		{
			name:   "path ref",
			config: `{"swagger": "2.0","paths":{"/foo":{"$ref":"./foo.json"}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "./foo.json", config.Paths["/foo"].Ref)
			},
		},
		{
			name:   "GET /foo",
			config: `{"swagger": "2.0","paths":{"/foo":{"get":{}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths["/foo"].Value.Get)
			},
		},
		{
			"PUT /foo",
			`{"swagger": "2.0","paths":{"/foo":{"put":{}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths["/foo"].Value.Put)
			},
		},
		{
			name:   "POST /foo",
			config: `{"swagger": "2.0","paths":{"/foo":{"post":{}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths["/foo"].Value.Post)
			},
		},
		{
			name:   "DELETE /foo",
			config: `{"swagger": "2.0","paths":{"/foo":{"delete":{}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths["/foo"].Value.Delete)
			},
		},
		{
			name:   "OPTIONS /foo",
			config: `{"swagger": "2.0","paths":{"/foo":{"options":{}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths["/foo"].Value.Options)
			},
		},
		{
			name:   "HEAD /foo",
			config: `{"swagger": "2.0","paths":{"/foo":{"head":{}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths["/foo"].Value.Head)
			},
		},
		{
			name:   "PATCH /foo",
			config: `{"swagger": "2.0","paths":{"/foo":{"patch":{}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths["/foo"].Value.Patch)
			},
		},
		{
			name:   "path parameter",
			config: `{"swagger": "2.0", "paths": {"/foo/{id}":{"parameters": [{"name": "id","in":"path","required":true,"description":"id parameter","type":"integer","format":"int64"}]}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo/{id}")
				p := config.Paths["/foo/{id}"].Value
				require.Equal(t, parameter.Path, p.Parameters[0].Value.Type)
				require.Equal(t, "id", p.Parameters[0].Value.Name)
				require.True(t, p.Parameters[0].Value.Required)
				require.Equal(t, "integer", p.Parameters[0].Value.Schema.Type.String())
				require.Equal(t, "int64", p.Parameters[0].Value.Schema.Format)
				require.Equal(t, "id parameter", p.Parameters[0].Value.Description)
			},
		},
		{
			name:   "operation tags",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"tags": ["foo","bar"]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.Equal(t, []string{"foo", "bar"}, get.Tags)
			},
		},
		{
			name:   "operation summary",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"summary": "foo"}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.Equal(t, "foo", get.Summary)
			},
		},
		{
			name:   "operation summary",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"summary": "foo"}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.Equal(t, "foo", get.Summary)
			},
		},
		{
			name:   "operation description",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"description": "foo"}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.Equal(t, "foo", get.Description)
			},
		},
		{
			name:   "operation operationId",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"operationId": "foo"}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.Equal(t, "foo", get.OperationId)
			},
		},
		{
			name:   "operation consumes without parameter body",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"consumes": ["application/json"]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.Nil(t, get.RequestBody)
			},
		},
		{
			name:   "operation parameter body and empty consumes",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"parameters": [{"in":"body","name":"body","schema":{"type":["string"]}}]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.Contains(t, get.RequestBody.Value.Content, "application/json")
				require.Equal(t, "string", get.RequestBody.Value.Content["application/json"].Schema.Type.String())
			},
		},
		{
			name:   "operation parameter body required",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"parameters": [{"in":"body","name":"body","required":true,"schema":{"type":["string"]}}]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.True(t, get.RequestBody.Value.Required)
			},
		},
		{
			name:   "operation parameter body and consumes",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"consumes":["application/json"],"parameters": [{"in":"body","name":"body","schema":{"type":["string"]}}]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.Contains(t, get.RequestBody.Value.Content, "application/json")
				require.Equal(t, "string", get.RequestBody.Value.Content["application/json"].Schema.Type.String())
			},
		},
		{
			name:   "operation parameter body and global consumes",
			config: `{"swagger": "2.0","consumes":["application/json"],"paths": {"/foo": {"get": {"parameters": [{"in":"body","name":"body","schema":{"type":["string"]}}]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.Contains(t, get.RequestBody.Value.Content, "application/json")
				require.Equal(t, "string", get.RequestBody.Value.Content["application/json"].Schema.Type.String())
			},
		},
		{
			name:   "operation parameter body empty consumes and global consumes",
			config: `{"swagger": "2.0","consumes":["application/json"],"paths": {"/foo": {"get": {"consumes":[],"parameters": [{"in":"body","name":"body","schema":{"type":["string"]}}]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.Contains(t, get.RequestBody.Value.Content, "application/json")
				require.Equal(t, "string", get.RequestBody.Value.Content["application/json"].Schema.Type.String())
			},
		},
		{
			name:   "operation parameter body empty consumes and global consumes",
			config: `{"swagger": "2.0","consumes":["text/plain"],"paths": {"/foo": {"get": {"consumes":["application/json"],"parameters": [{"in":"body","name":"body","schema":{"type":["string"]}}]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.Contains(t, get.RequestBody.Value.Content, "application/json")
				require.Equal(t, "string", get.RequestBody.Value.Content["application/json"].Schema.Type.String())
			},
		},
		{
			name:   "operation parameter path",
			config: `{"swagger": "2.0","paths": {"/foo": {"get": {"parameters": [{"in":"path","name":"id"}]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.Equal(t, "path", get.Parameters[0].Value.Type.String())
			},
		},
		{
			name:   "operation parameter query",
			config: `{"swagger": "2.0","paths": {"/foo": {"get": {"parameters": [{"in":"query","name":"id"}]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.Equal(t, "query", get.Parameters[0].Value.Type.String())
			},
		},
		{
			name:   "operation parameter header",
			config: `{"swagger": "2.0","paths": {"/foo": {"get": {"parameters": [{"in":"header","name":"id"}]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				get := config.Paths["/foo"].Value.Get
				require.Equal(t, "header", get.Parameters[0].Value.Type.String())
			},
		},
		{
			name:   "operation default response",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"responses": {   "default": { "description": "default" }  }}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				p := config.Paths["/foo"]
				require.NotNil(t, p.Value.Get)
				get := p.Value.Get
				res, ok := get.Responses.Get("default")
				require.True(t, ok)
				require.NotNil(t, res)
				require.Equal(t, "default", res.Value.Description)
			},
		},
		{
			name:   "operation responses order",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"responses": {   "200": { "description": "200" }, "204": { "description": "200" }, "202": { "description": "200" }, "301": { "description": "301" }, "404": { "description": "404" }  }}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				p := config.Paths["/foo"]
				require.NotNil(t, p.Value.Get)
				get := p.Value.Get
				res := get.Responses
				require.Equal(t, 5, res.Len())
				require.Equal(t, []string{"200", "204", "202", "301", "404"}, res.Keys())
			},
		},
		{
			name:   "path parameter body",
			config: `{"swagger": "2.0", "paths": {"/foo/{id}":{"parameters": [{"name": "id","in":"body","required":true,"description":"id parameter","schema":{"type": "string"}}],"get":{"consumes":["application/json"]}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo/{id}")
				p := config.Paths["/foo/{id}"].Value
				require.NotNil(t, p.Get.RequestBody)
				require.NotNil(t, p.Get.RequestBody.Value)
				require.Equal(t, "id parameter", p.Get.RequestBody.Value.Description)
				require.Contains(t, p.Get.RequestBody.Value.Content, "application/json")
				content := p.Get.RequestBody.Value.Content["application/json"]
				require.Equal(t, "string", content.Schema.Type.String())
			},
		},
		{
			name:   "path response",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"produces": ["application/json"], "responses": {"200": {"description": "response description", "schema": {"$ref": "#/definitions/foo"}}}}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				p := config.Paths["/foo"]
				require.NotNil(t, p.Value.Get)
				get := p.Value.Get
				ok := get.Responses.GetResponse(http.StatusOK)
				require.NotNil(t, ok)
				require.Equal(t, "response description", ok.Description)
				require.Equal(t, "#/components/schemas/foo", ok.Content["application/json"].Schema.Ref)
			},
		},
		{
			name:   "path response root produces",
			config: `{"swagger": "2.0", "produces": ["application/json"], "paths": {"/foo": {"get": {"responses": {"200": {"description": "response description", "schema": {"$ref": "#/definitions/foo"}}}}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				p := config.Paths["/foo"]
				require.NotNil(t, p.Value.Get)
				get := p.Value.Get
				ok := get.Responses.GetResponse(http.StatusOK)
				require.NotNil(t, ok)
				require.Equal(t, "response description", ok.Description)
				require.Equal(t, "#/components/schemas/foo", ok.Content["application/json"].Schema.Ref)
			},
		},
		{
			name:   "path response with default MIME type",
			config: `{"swagger": "2.0", "paths": {"/foo": {"get": {"responses": {"200": {"description": "response description", "schema": {"$ref": "#/definitions/foo"}}}}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths, "/foo")
				p := config.Paths["/foo"]
				require.NotNil(t, p.Value.Get)
				get := p.Value.Get
				ok := get.Responses.GetResponse(http.StatusOK)
				require.NotNil(t, ok)
				require.Equal(t, "response description", ok.Description)
				require.Equal(t, "#/components/schemas/foo", ok.Content["application/json"].Schema.Ref)
			},
		},
		{
			name:   "definitions",
			config: `{"swagger": "2.0", "definitions": {"Foo": {"type": "string"}, "Bar": {"type": "object","properties": {"title":{"type": ["string"]}}}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				foo := config.Components.Schemas.Get("Foo")
				require.Equal(t, "string", foo.Type.String())
				bar := config.Components.Schemas.Get("Bar")
				require.Equal(t, "object", bar.Type.String())
				title := bar.Properties.Get("title")
				require.Equal(t, "string", title.Type.String())
			},
		},
		{
			name:   "integer with empty format needs to be int32",
			config: `{"swagger": "2.0", "definitions": {"Foo": {"type": ["integer"]}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				foo := config.Components.Schemas.Get("Foo")
				require.Equal(t, "int32", foo.Format)
			},
		},
		{
			name:   "integer with format needs to be int64",
			config: `{"swagger": "2.0", "definitions": {"Foo": {"type": ["integer"], "format": "int64"}}}`,
			test: func(t *testing.T, config *openapi.Config) {
				foo := config.Components.Schemas.Get("Foo")
				require.Equal(t, "int64", foo.Format)
			},
		},
		{
			name:   "security apiKey",
			config: `{"swagger": "2.0", "securityDefinitions": { "api_key": { "type": "apiKey", "name": "api_key", "in": "header" } } }`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Components.SecuritySchemes, 1)
				apiKey := config.Components.SecuritySchemes["api_key"].(*openapi.ApiKeySecurityScheme)
				require.Equal(t, "apiKey", apiKey.Type)
				require.Equal(t, "api_key", apiKey.Name)
				require.Equal(t, "header", apiKey.In)
			},
		},
		{
			name:   "security oauth2",
			config: `{"swagger": "2.0", "securityDefinitions": { "oauth": { "type": "oauth2", "description": "foo", "authorizationUrl": "https://swagger.io/api/oauth/dialog", "flow": "implicit", "scopes": { "write:pets": "modify pets in your account", "read:pets": "read your pets" } } } }`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Components.SecuritySchemes, 1)
				oauth2 := config.Components.SecuritySchemes["oauth"].(*openapi.OAuth2SecurityScheme)
				require.Equal(t, "oauth2", oauth2.Type)
				require.Equal(t, "foo", oauth2.Description)
				require.Equal(t, "https://swagger.io/api/oauth/dialog", oauth2.Flows["implicit"].AuthorizationUrl)
				require.Equal(t, "modify pets in your account", oauth2.Flows["implicit"].Scopes["write:pets"])
				require.Equal(t, "read your pets", oauth2.Flows["implicit"].Scopes["read:pets"])
			},
		},
		{
			name:   "security basic",
			config: `{"swagger": "2.0", "securityDefinitions": { "foo": { "type": "basic", "description": "foo" } } }`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Components.SecuritySchemes, 1)
				basic := config.Components.SecuritySchemes["foo"].(*openapi.HttpSecurityScheme)
				require.Equal(t, "http", basic.Type)
				require.Equal(t, "basic", basic.Scheme)
				require.Equal(t, "foo", basic.Description)
			},
		},
		{
			name:   "security requirement",
			config: `{"swagger": "2.0", "security": [ { "foo": [] } ] }`,
			test: func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Security, 1)
				require.Equal(t, openapi.SecurityRequirement{"foo": nil}, config.Security[0])
			},
		},
		{
			name:   "security requirement on operation",
			config: `{"swagger": "2.0", "paths": { "/pet": { "get": { "security": [ { "foo": [] } ] } } } }`,
			test: func(t *testing.T, config *openapi.Config) {
				security := config.Paths["/pet"].Value.Get.Security
				require.Len(t, security, 1)
				require.Equal(t, openapi.SecurityRequirement{"foo": nil}, security[0])
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
			tc.test(t, converted)
		})
	}
}
