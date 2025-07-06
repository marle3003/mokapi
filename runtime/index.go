package runtime

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/char/html"
	_ "github.com/blevesearch/bleve/v2/analysis/token/ngram"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/unicode"
	"github.com/blevesearch/bleve/v2/search/query"
	index "github.com/blevesearch/bleve_index_api"
	log "github.com/sirupsen/logrus"
	"mokapi/config/static"

	"strings"
)

type SearchResult struct {
	Type       string   `json:"type"`
	ConfigName string   `json:"configName"`
	Title      string   `json:"title"`
	Fragments  []string `json:"fragments,omitempty"`
}

func newIndex(cfg *static.Config) bleve.Index {
	if !cfg.Api.Search.Enabled {
		return nil
	}

	mapping := bleve.NewIndexMapping()

	err := mapping.AddCustomTokenFilter("ngram_filter", map[string]any{
		"type": "ngram",
		"min":  cfg.Api.Search.Ngram.Min,
		"max":  cfg.Api.Search.Ngram.Max,
	})
	if err != nil {
		panic(err)
	}

	err = mapping.AddCustomAnalyzer("ngram", map[string]any{
		"type":          custom.Name,
		"tokenizer":     unicode.Name,
		"token_filters": []any{"to_lower", "ngram_filter"},
	})
	if err != nil {
		panic(err)
	}

	mapping.DefaultAnalyzer = cfg.Api.Search.Analyzer

	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		log.Error(err)
	}

	return idx
}

func add(index bleve.Index, id string, data any) {
	err := index.Index(id, data)
	if err != nil {
		log.Errorf("add '%s' to search index failed: %v", id, err)
	}
}

func (a *App) Search(queryText string) ([]SearchResult, error) {
	var q query.Query
	if queryText == "" {
		q = bleve.NewMatchAllQuery()
	} else {
		q = bleve.NewQueryStringQuery(queryText)
	}

	sr := bleve.NewSearchRequest(q)
	sr.Highlight = bleve.NewHighlightWithStyle(html.Name)
	result, err := a.index.Search(sr)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, hit := range result.Hits {

		doc, err := a.index.Document(hit.ID)
		if err != nil {
			return nil, err
		}
		fields := getSearchFields(doc)

		discriminators := strings.Split(fields["Discriminator"], "_")
		var searchResult *SearchResult
		switch discriminators[0] {
		case "http":
			searchResult, err = getHttpSearchResult(fields, discriminators)
		case "config":
			searchResult, err = getConfigSearchResult(fields, discriminators)
		default:
			log.Errorf("unknown discriminator: %s", strings.Join(discriminators, "_"))
			continue
		}

		if err != nil {
			log.Errorf("error on search result: %v", err)
			continue
		}

		c := 0
		for _, fragments := range hit.Fragments {
			if c > 3 {
				break
			}
			c++
			searchResult.Fragments = append(searchResult.Fragments, fragments...)
		}
		results = append(results, *searchResult)
	}
	return results, nil
}

func getSearchFields(doc index.Document) map[string]string {
	m := make(map[string]string)
	doc.VisitFields(func(field index.Field) {
		m[field.Name()] = string(field.Value())
	})
	return m
}
