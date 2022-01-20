package openapi

import (
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/test"
	"net/url"
	"testing"
)

type testReader struct {
	readFunc func(file *common.File) error
}

func (tr *testReader) Read(u *url.URL, opts ...common.FileOptions) (*common.File, error) {
	file := &common.File{Url: u}
	for _, opt := range opts {
		opt(file)
	}
	if err := tr.readFunc(file); err != nil {
		return file, err
	}
	if p, ok := file.Data.(common.Parser); ok {
		return file, p.Parse(file, tr)
	}
	return file, nil
}

func (tr *testReader) Close() {}

func TestResolve(t *testing.T) {
	t.Run("empty should not error", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.File) error { return nil }}
		config := &Config{}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
	})
}

func TestEndpointResolve(t *testing.T) {
	t.Run("nil should not error", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.File) error { return nil }}
		config := &Config{EndPoints: map[string]*EndpointRef{"foo": nil}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
	})
	t.Run("file reference", func(t *testing.T) {
		target := &Endpoint{}
		reader := &testReader{readFunc: func(file *common.File) error {
			test.Equals(t, "/foo.yml#/endpoints/foo", file.Url.String())
			config := &Config{EndPoints: map[string]*EndpointRef{
				"foo": {Value: target},
			}}
			file.Data = config
			return nil
		}}
		config := &Config{EndPoints: map[string]*EndpointRef{
			"foo": {Reference: ref.Reference{Value: "foo.yml#/endpoints/foo"}},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, target, config.EndPoints["foo"].Value)
	})
	t.Run("file reference but nil", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.File) error {
			test.Equals(t, "/foo.yml#/endpoints/foo", file.Url.String())
			config := &Config{EndPoints: map[string]*EndpointRef{
				"foo": {},
			}}
			file.Data = config
			return nil
		}}
		config := &Config{EndPoints: map[string]*EndpointRef{
			"foo": {Reference: ref.Reference{Value: "foo.yml#/endpoints/foo"}},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, nil, config.EndPoints["foo"].Value)
	})
	t.Run("reader returns error", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.File) error {
			return test.TestError
		}}
		config := &Config{EndPoints: map[string]*EndpointRef{
			"foo": {Reference: ref.Reference{Value: "foo.yml#/endpoints/foo"}},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Error(t, err)
		test.Equals(t, test.TestError, err)
	})
}
