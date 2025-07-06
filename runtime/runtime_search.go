package runtime

import (
	"github.com/blevesearch/bleve/v2"
	"mokapi/config/dynamic"
	"path/filepath"
)

type config struct {
	Discriminator string
	ConfigName    string `json:"configName"`
	Name          string
	Data          string
}

func addConfigToIndex(index bleve.Index, cfg *dynamic.Config) error {
	return index.Index(cfg.Info.Key(), config{
		Discriminator: "config",
		ConfigName:    cfg.Info.Path(),
		Name:          filepath.Base(cfg.Info.Path()),
		Data:          string(cfg.Raw),
	})
}

func removeConfigFromIndex(index bleve.Index, cfg *dynamic.Config) {
	_ = index.Delete(cfg.Info.Key())
}

func getConfigSearchResult(fields map[string]string, _ []string) (*SearchResult, error) {
	return &SearchResult{
		Type:       "Config",
		ConfigName: fields["ConfigName"],
		Title:      fields["Name"],
	}, nil
}
