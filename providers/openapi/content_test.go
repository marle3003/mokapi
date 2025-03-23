package openapi_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"testing"
)

func TestContent_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "application/json",
			test: func(t *testing.T) {
				c := openapi.Content{}
				err := json.Unmarshal([]byte(`{ "application/json": {} }`), &c)
				require.NoError(t, err)
				require.Len(t, c, 1)
				require.Contains(t, c, "application/json")
				require.Equal(t, "application", c["application/json"].ContentType.Type)
				require.Equal(t, "json", c["application/json"].ContentType.Subtype)
			},
		},
		{
			name: "text/*",
			test: func(t *testing.T) {
				c := openapi.Content{}
				err := json.Unmarshal([]byte(`{ "text/*": {} }`), &c)
				require.NoError(t, err)
				require.Len(t, c, 1)
				require.Contains(t, c, "text/*")
				require.Equal(t, "text", c["text/*"].ContentType.Type)
				require.Equal(t, "*", c["text/*"].ContentType.Subtype)
			},
		},
		{
			name: "content is nil",
			test: func(t *testing.T) {
				var c openapi.Content
				err := json.Unmarshal([]byte(`{ "application/json": {} }`), &c)
				require.NoError(t, err)
				require.Len(t, c, 1)
			},
		},
		{
			name: "unexpected token",
			test: func(t *testing.T) {
				var c openapi.Content
				err := json.Unmarshal([]byte(`[]`), &c)
				require.EqualError(t, err, "expected openapi.Content map, got [")
			},
		},
		{
			name: "MediaType unexpected array",
			test: func(t *testing.T) {
				c := openapi.Content{}
				err := json.Unmarshal([]byte(`{ "application/json": [] }`), &c)
				require.EqualError(t, err, "json: cannot unmarshal array into Go value of type openapi.MediaType")
				require.Len(t, c, 0)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestContent_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "application/json",
			test: func(t *testing.T) {
				c := openapi.Content{}
				err := yaml.Unmarshal([]byte(`'application/json': {}`), &c)
				require.NoError(t, err)
				require.Len(t, c, 1)
				require.Contains(t, c, "application/json")
				require.Equal(t, "application", c["application/json"].ContentType.Type)
				require.Equal(t, "json", c["application/json"].ContentType.Subtype)
			},
		},
		{
			name: "text/*",
			test: func(t *testing.T) {
				c := openapi.Content{}
				err := yaml.Unmarshal([]byte(`'text/*': {}`), &c)
				require.NoError(t, err)
				require.Len(t, c, 1)
				require.Contains(t, c, "text/*")
				require.Equal(t, "text", c["text/*"].ContentType.Type)
				require.Equal(t, "*", c["text/*"].ContentType.Subtype)
			},
		},
		{
			name: "content is nil",
			test: func(t *testing.T) {
				var c openapi.Content
				err := yaml.Unmarshal([]byte(`'application/json': {}`), &c)
				require.NoError(t, err)
				require.Len(t, c, 1)
			},
		},
		{
			name: "unexpected token",
			test: func(t *testing.T) {
				var c openapi.Content
				err := yaml.Unmarshal([]byte(`[]`), &c)
				require.EqualError(t, err, "expected openapi.Content map, got !!seq")
			},
		},
		{
			name: "MediaType unexpected array",
			test: func(t *testing.T) {
				c := openapi.Content{}
				err := yaml.Unmarshal([]byte(`'application/json': []`), &c)
				require.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into openapi.MediaType")
				require.Len(t, c, 0)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestConfig_Patch_Content(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "add MediaType",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{}}),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Len(t, res.Content, 1)
				require.Contains(t, res.Content, "text/plain")
				require.NotNil(t, res.Content["text/plain"])
			},
		},
		{
			name: "append MediaType",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Len(t, res.Content, 2)
				require.NotNil(t, res.Content["text/plain"])
				require.NotNil(t, res.Content["text/html"])
				require.NotEqual(t, res.Content["text/plain"], res.Content["text/html"])
			},
		},
		{
			name: "patch content",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{Schema: schematest.New("string")})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Len(t, res.Content, 1)
				require.NotNil(t, res.Content["text/plain"])
				require.Equal(t, "string", res.Content["text/plain"].Schema.Type.String())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}
