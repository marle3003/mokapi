package openapi_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/config/dynamic/openapi/schema"
	"net/url"
	"testing"
)

type testReader struct {
	readFunc func(cfg *common.Config) error
}

func (tr *testReader) Read(u *url.URL, opts ...common.ConfigOptions) (*common.Config, error) {
	cfg := common.NewConfig(u)
	for _, opt := range opts {
		opt(cfg, true)
	}
	if err := tr.readFunc(cfg); err != nil {
		return cfg, err
	}
	if p, ok := cfg.Data.(common.Parser); ok {
		return cfg, p.Parse(cfg, tr)
	}
	return cfg, nil
}

func (tr *testReader) Close() {}

func TestResolve(t *testing.T) {
	t.Run("empty should not error", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error { return nil }}
		config := &openapi.Config{}
		err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
		require.NoError(t, err)
	})
}

func TestEndpointResolve(t *testing.T) {
	t.Run("nil should not error", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error { return nil }}
		config := openapitest.NewConfig("3.0", openapitest.WithEndpoint("foo", nil))
		err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
		require.NoError(t, err)
	})
	t.Run("file reference", func(t *testing.T) {
		target := &openapi.Endpoint{}
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			require.Equal(t, "/foo.yml", cfg.Info.Url.String())
			config := openapitest.NewConfig("3.0", openapitest.WithEndpoint("/foo", target))
			cfg.Data = config
			return nil
		}}
		config := openapitest.NewConfig("3.0", openapitest.WithEndpointRef("/foo", &openapi.EndpointRef{Reference: ref.Reference{Ref: "foo.yml#/paths/foo"}}))
		err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
		require.NoError(t, err)
		require.Equal(t, target, config.Paths.Value["/foo"].Value)
	})
	t.Run("file reference but nil", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			require.Equal(t, "/foo.yml", cfg.Info.Url.String())
			config := openapitest.NewConfig("3.0", openapitest.WithEndpoint("/foo", nil))
			cfg.Data = config
			return nil
		}}
		config := openapitest.NewConfig("3.0", openapitest.WithEndpointRef("/foo", &openapi.EndpointRef{Reference: ref.Reference{Ref: "foo.yml#/paths/foo"}}))
		err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
		require.NoError(t, err)
		require.Nil(t, config.Paths.Value["/foo"].Value)
	})
	t.Run("reader returns error", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			return fmt.Errorf("TEST ERROR")
		}}
		config := openapitest.NewConfig("3.0", openapitest.WithEndpointRef("foo", &openapi.EndpointRef{Reference: ref.Reference{Ref: "foo.yml#/paths/foo"}}))
		err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
		require.EqualError(t, err, "unable to read /foo.yml#/paths/foo: TEST ERROR")
	})
	t.Run("schema ref", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			config := openapitest.NewConfig("3.0", openapitest.WithComponentSchemaRef("Foo", &schema.Ref{Value: &schema.Schema{Type: "string"}}))
			cfg.Data = config
			return nil
		}}
		config := openapitest.NewConfig("3.0", openapitest.WithComponentSchemaRef("Foo", &schema.Ref{Reference: ref.Reference{Ref: "foo.yml#/components/schemas/Foo"}}))
		err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
		require.NoError(t, err)
		s := config.Components.Schemas.Get("Foo")
		require.Equal(t, "string", s.Value.Type)
	})
	t.Run("paths ref", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			config := openapitest.NewConfig("3.0", openapitest.WithEndpoint("/foo", openapitest.NewEndpoint()))
			cfg.Data = config
			return nil
		}}
		config := openapitest.NewConfig("3.0", openapitest.WithEndpointsRef("paths.yml#/paths"))
		err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
		require.NoError(t, err)
		endpoint := config.Paths.Value["/foo"]
		require.NotNil(t, endpoint)
	})
}
