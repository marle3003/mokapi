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
	"mokapi/runtime/events"
	"mokapi/runtime/search"

	"strings"
)

func newIndex(cfg *static.Config) bleve.Index {
	if !cfg.Api.Search.Enabled {
		return nil
	}

	// 💡 Disable indexing for "_title"
	disableIndex := bleve.NewTextFieldMapping()
	disableIndex.Index = false
	disableIndex.Store = true

	docMapping := bleve.NewDocumentMapping()
	docMapping.AddFieldMappingsAt("_title", disableIndex)
	docMapping.AddFieldMappingsAt("discriminator", disableIndex)

	// enable term vectors for all fields, allowing phrase queries (like "Swagger Petstore")
	defaultField := bleve.NewTextFieldMapping()
	defaultField.IncludeTermVectors = true
	docMapping.AddFieldMappingsAt("*", defaultField)

	mapping := bleve.NewIndexMapping()
	mapping.DefaultMapping = docMapping
	mapping.DefaultAnalyzer = cfg.Api.Search.Analyzer

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

func (a *App) Search(r search.Request) (search.Result, error) {
	result := search.Result{}

	if a.index == nil {
		return result, &search.ErrNotEnabled{}
	}

	var clauses []query.Query
	if r.Query == "" {
		clauses = append(clauses, bleve.NewMatchAllQuery())
	} else {
		clauses = append(clauses, bleve.NewQueryStringQuery(r.Query))
	}

	for k, v := range r.Terms {
		term := bleve.NewMatchPhraseQuery(v)
		term.SetField(k)
		clauses = append(clauses, term)
	}

	q := bleve.NewConjunctionQuery(clauses...)

	sr := bleve.NewSearchRequest(q)
	sr.Size = r.Limit
	sr.From = r.Limit * r.Index
	sr.SortBy([]string{"-_score", "_id"})
	sr.Highlight = bleve.NewHighlightWithStyle(html.Name)
	searchResult, err := a.index.Search(sr)
	if err != nil {
		return result, err
	}

	result.Total = searchResult.Total
	for _, hit := range searchResult.Hits {
		item := search.ResultItem{}

		doc, err := a.index.Document(hit.ID)
		if err != nil {
			return result, err
		}
		fields := getSearchFields(doc)

		discriminators := strings.Split(fields["discriminator"], "_")
		switch discriminators[0] {
		case "http":
			item, err = getHttpSearchResult(fields, discriminators)
		case "config":
			item, err = getConfigSearchResult(fields, discriminators)
		case "event":
			item, err = events.GetSearchResult(fields, discriminators)
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
			item.Fragments = append(item.Fragments, fragments...)
		}
		result.Results = append(result.Results, item)
	}
	return result, nil
}

func getSearchFields(doc index.Document) map[string]string {
	m := make(map[string]string)
	doc.VisitFields(func(field index.Field) {
		m[field.Name()] = string(field.Value())
	})
	return m
}
