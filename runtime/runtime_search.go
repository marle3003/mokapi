package runtime

import (
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

func (a *App) addConfigToIndex(cfg *dynamic.Config) {
	a.searchIndex.Add(cfg.Info.Key(), config{
		Discriminator: "config",
		Provider:      cfg.Info.Provider,
		Name:          cfg.Info.Path(),
		Id:            cfg.Info.Key(),
		Data:          string(cfg.Raw),
	})
}

func (a *App) removeConfigFromIndex(cfg *dynamic.Config) {
	a.searchIndex.Delete(cfg.Info.Key())
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
