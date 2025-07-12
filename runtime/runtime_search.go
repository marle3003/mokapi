package runtime

import (
	"github.com/blevesearch/bleve/v2"
	"mokapi/config/dynamic"
	"mokapi/runtime/search"
	"strings"
)

type config struct {
	Discriminator string `json:"discriminator"`
	Provider      string `json:"provider"`
	Name          string `json:"name"`
	Id            string `json:"id"`
	Data          string `json:"data"`
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

func getConfigSearchResult(fields map[string]string, _ []string) (search.ResultItem, error) {
	return search.ResultItem{
		Type:   "Config",
		Domain: strings.ToUpper(fields["provider"]),
		Title:  fields["name"],
		Params: map[string]string{
			"type": "config",
			"id":   fields["id"],
		},
	}, nil
}
