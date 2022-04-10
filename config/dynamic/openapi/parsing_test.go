package openapi

import (
	"fmt"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/test"
	"net/url"
	"testing"
)

type testReader struct {
	readFunc func(cfg *common.Config) error
}

func (tr *testReader) Read(u *url.URL, opts ...common.ConfigOptions) (*common.Config, error) {
	cfg := &common.Config{Url: u}
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
		config := &Config{}
		err := config.Parse(&common.Config{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
	})
}

func TestEndpointResolve(t *testing.T) {
	t.Run("nil should not error", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error { return nil }}
		config := &Config{EndPoints: map[string]*EndpointRef{"foo": nil}}
		err := config.Parse(&common.Config{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
	})
	t.Run("file reference", func(t *testing.T) {
		target := &Endpoint{}
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			test.Equals(t, "/foo.yml#/endpoints/foo", cfg.Url.String())
			config := &Config{EndPoints: map[string]*EndpointRef{
				"foo": {Value: target},
			}}
			cfg.Data = config
			return nil
		}}
		config := &Config{EndPoints: map[string]*EndpointRef{
			"foo": {Reference: ref.Reference{Value: "foo.yml#/endpoints/foo"}},
		}}
		err := config.Parse(&common.Config{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, target, config.EndPoints["foo"].Value)
	})
	t.Run("file reference but nil", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			test.Equals(t, "/foo.yml#/endpoints/foo", cfg.Url.String())
			config := &Config{EndPoints: map[string]*EndpointRef{
				"foo": {},
			}}
			cfg.Data = config
			return nil
		}}
		config := &Config{EndPoints: map[string]*EndpointRef{
			"foo": {Reference: ref.Reference{Value: "foo.yml#/endpoints/foo"}},
		}}
		err := config.Parse(&common.Config{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, nil, config.EndPoints["foo"].Value)
	})
	t.Run("reader returns error", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			return test.TestError
		}}
		config := &Config{EndPoints: map[string]*EndpointRef{
			"foo": {Reference: ref.Reference{Value: "foo.yml#/endpoints/foo"}},
		}}
		err := config.Parse(&common.Config{Url: &url.URL{}, Data: config}, reader)
		test.Error(t, err)
		test.Equals(t, fmt.Errorf("unable to read /foo.yml#/endpoints/foo: %v", test.TestError), err)
	})
}
