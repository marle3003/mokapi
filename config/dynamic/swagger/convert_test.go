package swagger

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/parameter"
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
			"header",
			`{"swagger": "2.0"}`,
			func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "3.0.1", config.OpenApi)
			},
		},
		{
			"info title",
			`{"swagger": "2.0", "info": {"title": "foo"}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "foo", config.Info.Name)
			},
		},
		{
			"info description",
			`{"swagger": "2.0", "info": {"description": "foo"}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "foo", config.Info.Description)
			},
		},
		{
			"contact",
			`{"swagger": "2.0", "info": {"contact": {"name": "foo","url":"http://foo.bar","email":"foo@bar.com"}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "foo", config.Info.Contact.Name)
				require.Equal(t, "http://foo.bar", config.Info.Contact.Url)
				require.Equal(t, "foo@bar.com", config.Info.Contact.Email)
			},
		},
		{
			"version",
			`{"swagger": "2.0", "info": {"version": "1.0"}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "1.0", config.Info.Version)
			},
		},
		{
			"host",
			`{"swagger": "2.0", "host": "server:8080"}`,
			func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "http://server:8080", config.Servers[0].Url)
			},
		},
		{
			"basePath",
			`{"swagger": "2.0", "basePath": "/foo"}`,
			func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "/foo", config.Servers[0].Url)
			},
		},
		{
			"host",
			`{"swagger": "2.0", "host": "foo"}`,
			func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "http://foo", config.Servers[0].Url)
			},
		},
		{
			"host with port",
			`{"swagger": "2.0", "host": "foo:8080"}`,
			func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "http://foo:8080", config.Servers[0].Url)
			},
		},
		{
			"host with port and basePath",
			`{"swagger": "2.0", "host": "foo:8080", "basePath": "/bar"}`,
			func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "http://foo:8080/bar", config.Servers[0].Url)
			},
		},
		{
			"scheme with basePath",
			`{"swagger": "2.0","schemes":["https"],"basePath": "/bar"}`,
			func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "https:///bar", config.Servers[0].Url)
			},
		},
		{
			"scheme with host and basePath",
			`{"swagger": "2.0","schemes":["https"],"host":"foo","basePath": "/bar"}`,
			func(t *testing.T, config *openapi.Config) {
				require.Len(t, config.Servers, 1)
				require.Equal(t, "https://foo/bar", config.Servers[0].Url)
			},
		},
		{
			"path ref",
			`{"swagger": "2.0","paths":{"/foo":{"$ref":"./foo.json"}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Equal(t, "./foo.json", config.Paths.Value["/foo"].Ref)
			},
		},
		{
			"GET /foo",
			`{"swagger": "2.0","paths":{"/foo":{"get":{}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths.Value["/foo"].Value.Get)
			},
		},
		{
			"PUT /foo",
			`{"swagger": "2.0","paths":{"/foo":{"put":{}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths.Value["/foo"].Value.Put)
			},
		},
		{
			"POST /foo",
			`{"swagger": "2.0","paths":{"/foo":{"post":{}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths.Value["/foo"].Value.Post)
			},
		},
		{
			"DELETE /foo",
			`{"swagger": "2.0","paths":{"/foo":{"delete":{}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths.Value["/foo"].Value.Delete)
			},
		},
		{
			"OPTIONS /foo",
			`{"swagger": "2.0","paths":{"/foo":{"options":{}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths.Value["/foo"].Value.Options)
			},
		},
		{
			"HEAD /foo",
			`{"swagger": "2.0","paths":{"/foo":{"head":{}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths.Value["/foo"].Value.Head)
			},
		},
		{
			"PATCH /foo",
			`{"swagger": "2.0","paths":{"/foo":{"patch":{}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.NotNil(t, config.Paths.Value["/foo"].Value.Patch)
			},
		},
		{
			"path parameter",
			`{"swagger": "2.0", "paths": {"/foo/{id}":{"parameters": [{"name": "id","in":"path","required":true,"description":"id parameter","type":"integer","format":"int64"}]}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo/{id}")
				p := config.Paths.Value["/foo/{id}"].Value
				require.Equal(t, parameter.Path, p.Parameters[0].Value.Type)
				require.Equal(t, "id", p.Parameters[0].Value.Name)
				require.True(t, p.Parameters[0].Value.Required)
				require.Equal(t, "integer", p.Parameters[0].Value.Schema.Value.Type)
				require.Equal(t, "int64", p.Parameters[0].Value.Schema.Value.Format)
				require.Equal(t, "id parameter", p.Parameters[0].Value.Description)
			},
		},
		{
			"operation tags",
			`{"swagger": "2.0", "paths": {"/foo": {"get": {"tags": ["foo","bar"]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.Equal(t, []string{"foo", "bar"}, get.Tags)
			},
		},
		{
			"operation summary",
			`{"swagger": "2.0", "paths": {"/foo": {"get": {"summary": "foo"}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.Equal(t, "foo", get.Summary)
			},
		},
		{
			"operation summary",
			`{"swagger": "2.0", "paths": {"/foo": {"get": {"summary": "foo"}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.Equal(t, "foo", get.Summary)
			},
		},
		{
			"operation description",
			`{"swagger": "2.0", "paths": {"/foo": {"get": {"description": "foo"}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.Equal(t, "foo", get.Description)
			},
		},
		{
			"operation operationId",
			`{"swagger": "2.0", "paths": {"/foo": {"get": {"operationId": "foo"}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.Equal(t, "foo", get.OperationId)
			},
		},
		{
			"operation consumes without parameter body",
			`{"swagger": "2.0", "paths": {"/foo": {"get": {"consumes": ["application/json"]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.Nil(t, get.RequestBody)
			},
		},
		{
			"operation parameter body and empty consumes",
			`{"swagger": "2.0", "paths": {"/foo": {"get": {"parameters": [{"in":"body","name":"body","schema":{"type":"string"}}]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.Contains(t, get.RequestBody.Value.Content, "*/*")
				require.Equal(t, "string", get.RequestBody.Value.Content["*/*"].Schema.Value.Type)
			},
		},
		{
			"operation parameter body required",
			`{"swagger": "2.0", "paths": {"/foo": {"get": {"parameters": [{"in":"body","name":"body","required":true,"schema":{"type":"string"}}]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.True(t, get.RequestBody.Value.Required)
			},
		},
		{
			"operation parameter body and consumes",
			`{"swagger": "2.0", "paths": {"/foo": {"get": {"consumes":["application/json"],"parameters": [{"in":"body","name":"body","schema":{"type":"string"}}]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.Contains(t, get.RequestBody.Value.Content, "application/json")
				require.Equal(t, "string", get.RequestBody.Value.Content["application/json"].Schema.Value.Type)
			},
		},
		{
			"operation parameter body and global consumes",
			`{"swagger": "2.0","consumes":["application/json"],"paths": {"/foo": {"get": {"parameters": [{"in":"body","name":"body","schema":{"type":"string"}}]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.Contains(t, get.RequestBody.Value.Content, "application/json")
				require.Equal(t, "string", get.RequestBody.Value.Content["application/json"].Schema.Value.Type)
			},
		},
		{
			"operation parameter body empty consumes and global consumes",
			`{"swagger": "2.0","consumes":["application/json"],"paths": {"/foo": {"get": {"consumes":[],"parameters": [{"in":"body","name":"body","schema":{"type":"string"}}]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.Contains(t, get.RequestBody.Value.Content, "*/*")
				require.Equal(t, "string", get.RequestBody.Value.Content["*/*"].Schema.Value.Type)
			},
		},
		{
			"operation parameter body empty consumes and global consumes",
			`{"swagger": "2.0","consumes":["text/plain"],"paths": {"/foo": {"get": {"consumes":["application/json"],"parameters": [{"in":"body","name":"body","schema":{"type":"string"}}]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.NotNil(t, get.RequestBody)
				require.Contains(t, get.RequestBody.Value.Content, "application/json")
				require.Equal(t, "string", get.RequestBody.Value.Content["application/json"].Schema.Value.Type)
			},
		},
		{
			"operation parameter path",
			`{"swagger": "2.0","paths": {"/foo": {"get": {"parameters": [{"in":"path","name":"id"}]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.Equal(t, "path", get.Parameters[0].Value.Type.String())
			},
		},
		{
			"operation parameter query",
			`{"swagger": "2.0","paths": {"/foo": {"get": {"parameters": [{"in":"query","name":"id"}]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.Equal(t, "query", get.Parameters[0].Value.Type.String())
			},
		},
		{
			"operation parameter header",
			`{"swagger": "2.0","paths": {"/foo": {"get": {"parameters": [{"in":"header","name":"id"}]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo")
				get := config.Paths.Value["/foo"].Value.Get
				require.Equal(t, "header", get.Parameters[0].Value.Type.String())
			},
		},
		{
			"path parameter body",
			`{"swagger": "2.0", "paths": {"/foo/{id}":{"parameters": [{"name": "id","in":"body","required":true,"description":"id parameter","schema":{"type": "string"}}],"get":{"consumes":["application/json"]}}}}`,
			func(t *testing.T, config *openapi.Config) {
				require.Contains(t, config.Paths.Value, "/foo/{id}")
				p := config.Paths.Value["/foo/{id}"].Value
				require.NotNil(t, p.Get.RequestBody)
				require.NotNil(t, p.Get.RequestBody.Value)
				require.Equal(t, "id parameter", p.Get.RequestBody.Value.Description)
				require.Contains(t, p.Get.RequestBody.Value.Content, "application/json")
				content := p.Get.RequestBody.Value.Content["application/json"]
				require.Equal(t, "string", content.Schema.Value.Type)
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
				foo := config.Components.Schemas.Get("Foo")
				require.Equal(t, "string", foo.Value.Type)
				bar := config.Components.Schemas.Get("Bar")
				require.Equal(t, "object", bar.Value.Type)
				title := bar.Value.Properties.Value.Get("title")
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
