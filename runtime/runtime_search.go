package runtime

import (
	"github.com/blevesearch/bleve/v2"
	"mokapi/config/dynamic"
	"strings"
)

type config struct {
	Discriminator string
	Provider      string
	Name          string
	Id            string
	Data          string
}

func addConfigToIndex(index bleve.Index, cfg *dynamic.Config) error {
	return index.Index(cfg.Info.Key(), config{
		Discriminator: "config",
		Provider:      cfg.Info.Provider,
		Name:          cfg.Info.Path(),
		Id:            cfg.Info.Key(),
		Data:          string(cfg.Raw),
	})
}

func removeConfigFromIndex(index bleve.Index, cfg *dynamic.Config) {
	_ = index.Delete(cfg.Info.Key())
}

func getConfigSearchResult(fields map[string]string, _ []string) (*SearchResult, error) {
	return &SearchResult{
		Type:   "Config",
		Domain: strings.ToUpper(fields["Provider"]),
		Title:  fields["Name"],
		Params: map[string]string{
			"type": "config",
			"id":   fields["Id"],
		},
	}, nil
}
