package runtime

import (
	"mokapi/config/dynamic"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/directory"
	"mokapi/providers/mail"
	"mokapi/providers/openapi"
	"mokapi/runtime/search"
	"path/filepath"
	"strings"
)

type config struct {
	Discriminator string `json:"discriminator"`
	Provider      string `json:"provider"`
	Name          string `json:"name"`
	Id            string `json:"id"`
	Data          string `json:"data"`
	Source        string `json:"source"`
}

func (a *App) addConfigToIndex(cfg *dynamic.Config) {
	data := config{
		Discriminator: "config",
		Provider:      cfg.Info.Provider,
		Name:          cfg.Info.Path(),
		Id:            cfg.Info.Key(),
		Data:          string(cfg.Raw),
	}

	switch cfg.Data.(type) {
	case *openapi.Config:
		data.Source = "OpenAPI"
	case *asyncapi3.Config:
		data.Source = "AsyncAPI"
	case *directory.Config:
		data.Source = "LDAP"
	case *mail.Config:
		data.Source = "LDAP"
	default:
		ext := filepath.Ext(cfg.Info.Path())
		switch ext {
		case ".js", ".ts":
			data.Source = "JavaScript"
		case ".ldif":
			data.Source = "LDIF"
		}
	}

	a.searchIndex.Add(cfg.Info.Key(), data)
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
			"type":   "config",
			"id":     fields["id"],
			"source": fields["source"],
		},
	}, nil
}
