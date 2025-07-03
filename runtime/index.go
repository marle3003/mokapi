package runtime

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/char/html"
	"github.com/blevesearch/bleve/v2/search/query"
	index "github.com/blevesearch/bleve_index_api"
	log "github.com/sirupsen/logrus"

	"strings"
)

type SearchResult struct {
	Type       string   `json:"type"`
	ConfigName string   `json:"configName"`
	Title      string   `json:"title"`
	Fragments  []string `json:"fragments,omitempty"`
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
