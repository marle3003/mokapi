package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
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
