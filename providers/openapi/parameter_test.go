package openapi_test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"net/url"
	"testing"
)

func TestParameterHeader_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "name",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := json.Unmarshal([]byte(`{ "name": "foo" }`), &param)
				require.NoError(t, err)
				require.Equal(t, "foo", param.Name)
			},
		},
		{
			name: "type",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := json.Unmarshal([]byte(`{ "in": "cookie" }`), &param)
				require.NoError(t, err)
				require.Equal(t, openapi.ParameterCookie, param.Type)
			},
		},
		{
			name: "description",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := json.Unmarshal([]byte(`{ "description": "foo" }`), &param)
				require.NoError(t, err)
				require.Equal(t, "foo", param.Description)
			},
		},
		{
			name: "required",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := json.Unmarshal([]byte(`{ "required": true  }`), &param)
				require.NoError(t, err)
				require.True(t, param.Required)
			},
		},
		{
			name: "deprecated",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := json.Unmarshal([]byte(`{ "deprecated": true  }`), &param)
				require.NoError(t, err)
				require.True(t, param.Deprecated)
			},
		},
		{
			name: "style",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := json.Unmarshal([]byte(`{ "style": "simple"  }`), &param)
				require.NoError(t, err)
				require.Equal(t, "simple", param.Style)
			},
		},
		{
			name: "explode",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := json.Unmarshal([]byte(`{ "explode": false  }`), &param)
				require.NoError(t, err)
				require.NotNil(t, param.Explode)
				require.False(t, *param.Explode)
			},
		},
		{
			name: "default explode when style is form",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := json.Unmarshal([]byte(`{ "style": "form" }`), &param)
				require.NoError(t, err)
				require.Nil(t, param.Explode)
				require.True(t, param.IsExplode(), "When style is form, the default value is true")
			},
		},
		{
			name: "default explode when style is not form",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := json.Unmarshal([]byte(`{ "style": "simple" }`), &param)
				require.NoError(t, err)
				require.Nil(t, param.Explode)
				require.False(t, param.IsExplode(), "For all other styles, the default value is false.")
			},
		},
		{
			name: "schema",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := json.Unmarshal([]byte(`{ "schema": {}  }`), &param)
				require.NoError(t, err)
				require.NotNil(t, param.Schema)
			},
		},
		{
			name: "reference",
			test: func(t *testing.T) {
				ref := &openapi.ParameterRef{}
				err := json.Unmarshal([]byte(`{ "$ref": "foo.yml"  }`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo.yml", ref.Ref)
			},
		},
		{
			name: "value",
			test: func(t *testing.T) {
				ref := &openapi.ParameterRef{}
				err := json.Unmarshal([]byte(`{ "description": "foo"  }`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo", ref.Value.Description)
			},
		},
		{
			name: "set default style",
			test: func(t *testing.T) {
				ref := &openapi.ParameterRef{}
				err := json.Unmarshal([]byte(`{ "in": "query" }`), &ref)
				require.NoError(t, err)
				require.Equal(t, "form", ref.Value.Style)
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

func TestParameterHeader_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "name",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := yaml.Unmarshal([]byte(`name: foo`), &param)
				require.NoError(t, err)
				require.Equal(t, "foo", param.Name)
			},
		},
		{
			name: "type",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := yaml.Unmarshal([]byte(`in: cookie`), &param)
				require.NoError(t, err)
				require.Equal(t, openapi.ParameterCookie, param.Type)
			},
		},
		{
			name: "description",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := yaml.Unmarshal([]byte(`description: foo`), &param)
				require.NoError(t, err)
				require.Equal(t, "foo", param.Description)
			},
		},
		{
			name: "required",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := yaml.Unmarshal([]byte(`required: true`), &param)
				require.NoError(t, err)
				require.True(t, param.Required)
			},
		},
		{
			name: "deprecated",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := yaml.Unmarshal([]byte(`deprecated: true`), &param)
				require.NoError(t, err)
				require.True(t, param.Deprecated)
			},
		},
		{
			name: "style",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := yaml.Unmarshal([]byte(`style: simple`), &param)
				require.NoError(t, err)
				require.Equal(t, "simple", param.Style)
			},
		},
		{
			name: "explode",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := yaml.Unmarshal([]byte(`explode: false`), &param)
				require.NoError(t, err)
				require.NotNil(t, param.Explode)
				require.False(t, *param.Explode)
			},
		},
		{
			name: "default explode when style is form",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := yaml.Unmarshal([]byte(`style: form`), &param)
				require.NoError(t, err)
				require.Nil(t, param.Explode)
				require.True(t, param.IsExplode(), "When style is form, the default value is true")
			},
		},
		{
			name: "default explode when style is not form",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := yaml.Unmarshal([]byte(`style: simple`), &param)
				require.NoError(t, err)
				require.Nil(t, param.Explode)
				require.False(t, param.IsExplode(), "For all other styles, the default value is false.")
			},
		},
		{
			name: "schema",
			test: func(t *testing.T) {
				param := &openapi.Parameter{}
				err := yaml.Unmarshal([]byte(`schema: {}`), &param)
				require.NoError(t, err)
				require.NotNil(t, param.Schema)
			},
		},
		{
			name: "reference",
			test: func(t *testing.T) {
				ref := &openapi.ParameterRef{}
				err := yaml.Unmarshal([]byte(`$ref: foo.yml`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo.yml", ref.Ref)
			},
		},
		{
			name: "value",
			test: func(t *testing.T) {
				ref := &openapi.ParameterRef{}
				err := yaml.Unmarshal([]byte(`description: foo`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo", ref.Value.Description)
			},
		},
		{
			name: "set default style",
			test: func(t *testing.T) {
				ref := &openapi.ParameterRef{}
				err := yaml.Unmarshal([]byte(`in: query`), &ref)
				require.NoError(t, err)
				require.Equal(t, "form", ref.Value.Style)
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

func TestParameterHeader_Parse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "reference is nil",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				param := openapi.Parameters{nil}
				c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: param}
				err := param.Parse(c, reader)
				require.NoError(t, err)
			},
		},
		{
			name: "reference",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					cfg := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}, Data: &openapi.Parameter{Description: "foo"}}
					return cfg, nil
				})
				param := openapi.Parameters{&openapi.ParameterRef{Reference: dynamic.Reference{Ref: "foo.yml"}}}
				err := param.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: param}, reader)
				require.NoError(t, err)
				require.Equal(t, "foo", param[0].Value.Description)
			},
		},
		{
			name: "schema reference",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					cfg := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}, Data: schematest.New("string")}
					return cfg, nil
				})
				param := openapi.Parameters{&openapi.ParameterRef{Value: &openapi.Parameter{Schema: &schema.Schema{Ref: "foo.yml"}}}}
				err := param.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: param}, reader)
				require.NoError(t, err)
				require.Equal(t, "string", param[0].Value.Schema.Type.String())
			},
		},
		{
			name: "error by resolving example ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				param := openapi.Parameters{&openapi.ParameterRef{Reference: dynamic.Reference{Ref: "foo.yml"}}}
				err := param.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: param}, reader)
				require.EqualError(t, err, "parse parameter index '0' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "error by resolving example ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				param := openapi.Parameters{&openapi.ParameterRef{Value: &openapi.Parameter{Schema: &schema.Schema{Ref: "foo.yml"}}}}
				err := param.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: param}, reader)
				require.EqualError(t, err, "parse parameter index '0' failed: parse schema failed: resolve reference 'foo.yml' failed: TEST ERROR")
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

func TestLocation_String(t *testing.T) {
	testcases := map[openapi.Location]string{
		openapi.ParameterHeader:      "header",
		openapi.ParameterCookie:      "cookie",
		openapi.ParameterPath:        "path",
		openapi.ParameterQuery:       "query",
		openapi.ParameterQueryString: "querystring",
	}

	for k, v := range testcases {
		require.Equal(t, v, k.String())
	}
}

func TestParameters_Patch(t *testing.T) {
	testcases := []struct {
		name    string
		configs []openapi.Parameters
		test    func(t *testing.T, result openapi.Parameters)
	}{
		{
			name: "add parameter",
			configs: []openapi.Parameters{
				{},
				{&openapi.ParameterRef{Value: &openapi.Parameter{Description: "foo"}}},
			},
			test: func(t *testing.T, result openapi.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, "foo", result[0].Value.Description)
			},
		},
		{
			name: "patch type",
			configs: []openapi.Parameters{
				{&openapi.ParameterRef{Value: &openapi.Parameter{Type: openapi.ParameterHeader}}},
				{&openapi.ParameterRef{Value: &openapi.Parameter{Type: openapi.ParameterCookie}}},
			},
			test: func(t *testing.T, result openapi.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, openapi.ParameterCookie, result[0].Value.Type)
			},
		},
		{
			name: "set schema",
			configs: []openapi.Parameters{
				{&openapi.ParameterRef{Value: &openapi.Parameter{}}},
				{&openapi.ParameterRef{Value: &openapi.Parameter{Schema: schematest.New("string")}}},
			},
			test: func(t *testing.T, result openapi.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, "string", result[0].Value.Schema.Type.String())
			},
		},
		{
			name: "patch schema",
			configs: []openapi.Parameters{
				{&openapi.ParameterRef{Value: &openapi.Parameter{Schema: schematest.New("number")}}},
				{&openapi.ParameterRef{Value: &openapi.Parameter{Schema: schematest.New("string")}}},
			},
			test: func(t *testing.T, result openapi.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, "[number, string]", result[0].Value.Schema.Type.String())
			},
		},
		{
			name: "patch required",
			configs: []openapi.Parameters{
				{&openapi.ParameterRef{Value: &openapi.Parameter{Required: true}}},
				{&openapi.ParameterRef{Value: &openapi.Parameter{Required: false}}},
			},
			test: func(t *testing.T, result openapi.Parameters) {
				require.Len(t, result, 1)
				require.False(t, result[0].Value.Required)
			},
		},
		{
			name: "patch description",
			configs: []openapi.Parameters{
				{&openapi.ParameterRef{Value: &openapi.Parameter{Description: "foo"}}},
				{&openapi.ParameterRef{Value: &openapi.Parameter{Description: "bar"}}},
			},
			test: func(t *testing.T, result openapi.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, "bar", result[0].Value.Description)
			},
		},
		{
			name: "patch deprecated",
			configs: []openapi.Parameters{
				{&openapi.ParameterRef{Value: &openapi.Parameter{}}},
				{&openapi.ParameterRef{Value: &openapi.Parameter{Deprecated: true}}},
			},
			test: func(t *testing.T, result openapi.Parameters) {
				require.Len(t, result, 1)
				require.True(t, result[0].Value.Deprecated)
			},
		},
		{
			name: "patch style",
			configs: []openapi.Parameters{
				{&openapi.ParameterRef{Value: &openapi.Parameter{Style: "foo"}}},
				{&openapi.ParameterRef{Value: &openapi.Parameter{Style: "bar"}}},
			},
			test: func(t *testing.T, result openapi.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, "bar", result[0].Value.Style)
			},
		},
		{
			name: "patch explode",
			configs: []openapi.Parameters{
				{&openapi.ParameterRef{Value: &openapi.Parameter{}}},
				{&openapi.ParameterRef{Value: &openapi.Parameter{Explode: explode(true)}}},
			},
			test: func(t *testing.T, result openapi.Parameters) {
				require.Len(t, result, 1)
				require.True(t, *result[0].Value.Explode)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := tc.configs[0]
			for _, patch := range tc.configs[1:] {
				p.Patch(patch)
			}
			tc.test(t, p)
		})
	}
}

func explode(b bool) *bool {
	return &b
}
