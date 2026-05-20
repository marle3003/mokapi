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

func BuildDescription(max int, values ...string) string {
	seen := map[string]bool{}
	var parts []string

	for _, v := range values {
		v = normalize(v)

		if v == "" || seen[v] {
			continue
		}

		seen[v] = true
		parts = append(parts, v)
	}

	return truncateWords(strings.Join(parts, " "), max)
}

func truncateWords(s string, max int) string {
	if len(s) <= max {
		return s
	}

	words := strings.Fields(s)
	var result strings.Builder
	for _, w := range words {
		next := w
		if result.Len() > 0 {
			next = " " + w
		}

		if result.Len()+len(next) > max {
			break
		}
		result.WriteString(next)
	}

	r := result.String()
	if r != s {
		r += "..."
	}
	return r
}

func normalize(s string) string {
	s = strings.TrimSpace(s)

	// Replace line breaks/tabs with spaces
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")

	// Collapse multiple spaces
	s = strings.Join(strings.Fields(s), " ")

	return s
}
