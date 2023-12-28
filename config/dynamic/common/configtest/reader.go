package configtest

import (
	"mokapi/config/dynamic/common"
	"net/url"
)

type Reader struct {
	ReadFunc func(cfg *common.Config) error
}

func (tr *Reader) Read(u *url.URL, opts ...common.ConfigOptions) (*common.Config, error) {
	cfg := common.NewConfig(common.ConfigInfo{Url: u})
	for _, opt := range opts {
		opt(cfg, true)
	}
	if err := tr.ReadFunc(cfg); err != nil {
		return cfg, err
	}
	if p, ok := cfg.Data.(common.Parser); ok {
		return cfg, p.Parse(cfg, tr)
	}
	return cfg, nil
}

func (tr *Reader) Close() {}
