package parameter_test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/common/readertest"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/config/dynamic/openapi/schema"
	"net/url"
	"testing"
)

func TestHeader_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "name",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := json.Unmarshal([]byte(`{ "name": "foo" }`), &param)
				require.NoError(t, err)
				require.Equal(t, "foo", param.Name)
			},
		},
		{
			name: "type",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := json.Unmarshal([]byte(`{ "in": "cookie" }`), &param)
				require.NoError(t, err)
				require.Equal(t, parameter.Cookie, param.Type)
			},
		},
		{
			name: "description",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := json.Unmarshal([]byte(`{ "description": "foo" }`), &param)
				require.NoError(t, err)
				require.Equal(t, "foo", param.Description)
			},
		},
		{
			name: "required",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := json.Unmarshal([]byte(`{ "required": true  }`), &param)
				require.NoError(t, err)
				require.True(t, param.Required)
			},
		},
		{
			name: "deprecated",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := json.Unmarshal([]byte(`{ "deprecated": true  }`), &param)
				require.NoError(t, err)
				require.True(t, param.Deprecated)
			},
		},
		{
			name: "style",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := json.Unmarshal([]byte(`{ "style": "simple"  }`), &param)
				require.NoError(t, err)
				require.Equal(t, "simple", param.Style)
			},
		},
		{
			name: "explode",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := json.Unmarshal([]byte(`{ "explode": false  }`), &param)
				require.NoError(t, err)
				require.False(t, param.Explode)
			},
		},
		{
			name: "schema",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := json.Unmarshal([]byte(`{ "schema": {}  }`), &param)
				require.NoError(t, err)
				require.NotNil(t, param.Schema)
			},
		},
		{
			name: "reference",
			test: func(t *testing.T) {
				ref := &parameter.Ref{}
				err := json.Unmarshal([]byte(`{ "$ref": "foo.yml"  }`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo.yml", ref.Ref)
			},
		},
		{
			name: "value",
			test: func(t *testing.T) {
				ref := &parameter.Ref{}
				err := json.Unmarshal([]byte(`{ "description": "foo"  }`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo", ref.Value.Description)
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

func TestHeader_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "name",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := yaml.Unmarshal([]byte(`name: foo`), &param)
				require.NoError(t, err)
				require.Equal(t, "foo", param.Name)
			},
		},
		{
			name: "type",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := yaml.Unmarshal([]byte(`in: cookie`), &param)
				require.NoError(t, err)
				require.Equal(t, parameter.Cookie, param.Type)
			},
		},
		{
			name: "description",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := yaml.Unmarshal([]byte(`description: foo`), &param)
				require.NoError(t, err)
				require.Equal(t, "foo", param.Description)
			},
		},
		{
			name: "required",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := yaml.Unmarshal([]byte(`required: true`), &param)
				require.NoError(t, err)
				require.True(t, param.Required)
			},
		},
		{
			name: "deprecated",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := yaml.Unmarshal([]byte(`deprecated: true`), &param)
				require.NoError(t, err)
				require.True(t, param.Deprecated)
			},
		},
		{
			name: "style",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := yaml.Unmarshal([]byte(`style: simple`), &param)
				require.NoError(t, err)
				require.Equal(t, "simple", param.Style)
			},
		},
		{
			name: "explode",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := yaml.Unmarshal([]byte(`explode: false`), &param)
				require.NoError(t, err)
				require.False(t, param.Explode)
			},
		},
		{
			name: "schema",
			test: func(t *testing.T) {
				param := &parameter.Parameter{}
				err := yaml.Unmarshal([]byte(`schema: {}`), &param)
				require.NoError(t, err)
				require.NotNil(t, param.Schema)
			},
		},
		{
			name: "reference",
			test: func(t *testing.T) {
				ref := &parameter.Ref{}
				err := yaml.Unmarshal([]byte(`$ref: foo.yml`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo.yml", ref.Ref)
			},
		},
		{
			name: "value",
			test: func(t *testing.T) {
				ref := &parameter.Ref{}
				err := yaml.Unmarshal([]byte(`description: foo`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo", ref.Value.Description)
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

func TestHeader_Parse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "reference is nil",
			test: func(t *testing.T) {
				reader := &readertest.Reader{ReadFunc: func(cfg *common.Config) error {
					return nil
				}}
				param := parameter.Parameters{nil}
				err := param.Parse(common.NewConfig(&url.URL{}, common.WithData(param)), reader)
				require.NoError(t, err)
			},
		},
		{
			name: "reference",
			test: func(t *testing.T) {
				reader := &readertest.Reader{ReadFunc: func(cfg *common.Config) error {
					cfg.Data = &parameter.Parameter{Description: "foo"}
					return nil
				}}
				param := parameter.Parameters{&parameter.Ref{Reference: ref.Reference{Ref: "foo.yml"}}}
				err := param.Parse(common.NewConfig(&url.URL{}, common.WithData(param)), reader)
				require.NoError(t, err)
				require.Equal(t, "foo", param[0].Value.Description)
			},
		},
		{
			name: "schema reference",
			test: func(t *testing.T) {
				reader := &readertest.Reader{ReadFunc: func(cfg *common.Config) error {
					cfg.Data = &schema.Schema{Type: "string"}
					return nil
				}}
				param := parameter.Parameters{&parameter.Ref{Value: &parameter.Parameter{Schema: &schema.Ref{Reference: ref.Reference{Ref: "foo.yml"}}}}}
				err := param.Parse(common.NewConfig(&url.URL{}, common.WithData(param)), reader)
				require.NoError(t, err)
				require.Equal(t, "string", param[0].Value.Schema.Value.Type)
			},
		},
		{
			name: "error by resolving example ref",
			test: func(t *testing.T) {
				reader := &readertest.Reader{ReadFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				param := parameter.Parameters{&parameter.Ref{Reference: ref.Reference{Ref: "foo.yml"}}}
				err := param.Parse(common.NewConfig(&url.URL{}, common.WithData(param)), reader)
				require.EqualError(t, err, "parse parameter index '0' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "error by resolving example ref",
			test: func(t *testing.T) {
				reader := &readertest.Reader{ReadFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				param := parameter.Parameters{&parameter.Ref{Value: &parameter.Parameter{Schema: &schema.Ref{Reference: ref.Reference{Ref: "foo.yml"}}}}}
				err := param.Parse(common.NewConfig(&url.URL{}, common.WithData(param)), reader)
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
	testcases := map[parameter.Location]string{
		parameter.Header: "header",
		parameter.Cookie: "cookie",
		parameter.Path:   "path",
		parameter.Query:  "query",
	}

	for k, v := range testcases {
		require.Equal(t, v, k.String())
	}
}

func TestParameters_Patch(t *testing.T) {
	testcases := []struct {
		name    string
		configs []parameter.Parameters
		test    func(t *testing.T, result parameter.Parameters)
	}{
		{
			name: "add parameter",
			configs: []parameter.Parameters{
				{},
				{&parameter.Ref{Value: &parameter.Parameter{Description: "foo"}}},
			},
			test: func(t *testing.T, result parameter.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, "foo", result[0].Value.Description)
			},
		},
		{
			name: "patch type",
			configs: []parameter.Parameters{
				{&parameter.Ref{Value: &parameter.Parameter{Type: parameter.Header}}},
				{&parameter.Ref{Value: &parameter.Parameter{Type: parameter.Cookie}}},
			},
			test: func(t *testing.T, result parameter.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, parameter.Cookie, result[0].Value.Type)
			},
		},
		{
			name: "set schema",
			configs: []parameter.Parameters{
				{&parameter.Ref{Value: &parameter.Parameter{}}},
				{&parameter.Ref{Value: &parameter.Parameter{Schema: &schema.Ref{Value: &schema.Schema{Type: "string"}}}}},
			},
			test: func(t *testing.T, result parameter.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, "string", result[0].Value.Schema.Value.Type)
			},
		},
		{
			name: "patch schema",
			configs: []parameter.Parameters{
				{&parameter.Ref{Value: &parameter.Parameter{Schema: &schema.Ref{Value: &schema.Schema{Type: "number"}}}}},
				{&parameter.Ref{Value: &parameter.Parameter{Schema: &schema.Ref{Value: &schema.Schema{Type: "string"}}}}},
			},
			test: func(t *testing.T, result parameter.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, "string", result[0].Value.Schema.Value.Type)
			},
		},
		{
			name: "patch required",
			configs: []parameter.Parameters{
				{&parameter.Ref{Value: &parameter.Parameter{Required: true}}},
				{&parameter.Ref{Value: &parameter.Parameter{Required: false}}},
			},
			test: func(t *testing.T, result parameter.Parameters) {
				require.Len(t, result, 1)
				require.False(t, result[0].Value.Required)
			},
		},
		{
			name: "patch description",
			configs: []parameter.Parameters{
				{&parameter.Ref{Value: &parameter.Parameter{Description: "foo"}}},
				{&parameter.Ref{Value: &parameter.Parameter{Description: "bar"}}},
			},
			test: func(t *testing.T, result parameter.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, "bar", result[0].Value.Description)
			},
		},
		{
			name: "patch deprecated",
			configs: []parameter.Parameters{
				{&parameter.Ref{Value: &parameter.Parameter{}}},
				{&parameter.Ref{Value: &parameter.Parameter{Deprecated: true}}},
			},
			test: func(t *testing.T, result parameter.Parameters) {
				require.Len(t, result, 1)
				require.True(t, result[0].Value.Deprecated)
			},
		},
		{
			name: "patch style",
			configs: []parameter.Parameters{
				{&parameter.Ref{Value: &parameter.Parameter{Style: "foo"}}},
				{&parameter.Ref{Value: &parameter.Parameter{Style: "bar"}}},
			},
			test: func(t *testing.T, result parameter.Parameters) {
				require.Len(t, result, 1)
				require.Equal(t, "bar", result[0].Value.Style)
			},
		},
		{
			name: "patch explode",
			configs: []parameter.Parameters{
				{&parameter.Ref{Value: &parameter.Parameter{}}},
				{&parameter.Ref{Value: &parameter.Parameter{Explode: false}}},
			},
			test: func(t *testing.T, result parameter.Parameters) {
				require.Len(t, result, 1)
				require.False(t, result[0].Value.Explode)
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
