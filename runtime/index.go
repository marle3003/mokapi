package runtime

import (
	"context"
	"mokapi/config/static"
	"mokapi/runtime/events"
	"mokapi/runtime/search"
	"mokapi/safe"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/char/asciifolding"
	"github.com/blevesearch/bleve/v2/analysis/char/html"
	_ "github.com/blevesearch/bleve/v2/analysis/lang/en"
	"github.com/blevesearch/bleve/v2/analysis/token/camelcase"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	_ "github.com/blevesearch/bleve/v2/analysis/token/ngram"
	"github.com/blevesearch/bleve/v2/analysis/token/porter"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/unicode"
	bleveSearch "github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	index "github.com/blevesearch/bleve_index_api"
	log "github.com/sirupsen/logrus"

	"strings"
)

var fieldsNotIncludedInAll = []string{"api"}
var SupportedFacets = []string{"type"}

type SearchIndex struct {
	cfg   static.Search
	idx   bleve.Index
	ready chan struct{}
	queue chan func()
}

func newSearchIndex(cfg static.Search) *SearchIndex {
	s := &SearchIndex{cfg: cfg}
	if cfg.Enabled {
		s.ready = make(chan struct{})
		s.queue = make(chan func(), 1000)
	}
	return s
}

func (s *SearchIndex) start(pool *safe.Pool) {
	if !s.cfg.Enabled {
		return
	}

	// 💡 Disable indexing for "_title"
	disableIndex := bleve.NewTextFieldMapping()
	disableIndex.Index = false
	disableIndex.Store = true

	docMapping := bleve.NewDocumentMapping()
	docMapping.AddFieldMappingsAt("_title", disableIndex)
	docMapping.AddFieldMappingsAt("_time", disableIndex)
	docMapping.AddFieldMappingsAt("discriminator", disableIndex)

	apiField := bleve.NewTextFieldMapping()
	apiField.Analyzer = "mokapi_analyzer"
	apiField.IncludeInAll = false // Exclude from default search
	apiField.Store = true
	apiField.Index = true
	docMapping.AddFieldMappingsAt("api", apiField)

	// enable term vectors for all fields, allowing phrase queries (like "Swagger Petstore")
	defaultField := bleve.NewTextFieldMapping()
	defaultField.IncludeTermVectors = true
	docMapping.AddFieldMappingsAt("*", defaultField)

	mapping := bleve.NewIndexMapping()
	mapping.DefaultMapping = docMapping
	mapping.DefaultAnalyzer = "mokapi_analyzer"

	stemmer := porter.Name

	err := mapping.AddCustomAnalyzer("mokapi_analyzer", map[string]any{
		"type":         custom.Name,
		"tokenizer":    unicode.Name,
		"char_filters": []any{asciifolding.Name},
		"token_filters": []any{
			lowercase.Name,
			camelcase.Name,
			stemmer,
		},
	})
	if err != nil {
		panic(err)
	}

	if !s.cfg.InMemory {
		indexPath := getSearchIndexPath(s.cfg)
		_ = os.RemoveAll(indexPath)
		s.idx, err = bleve.New(indexPath, mapping)
	} else {
		s.idx, err = bleve.NewMemOnly(mapping)
	}

	if err != nil {
		log.Errorf("disabling search due to error: %s", err)
		s.cfg.Enabled = false
		close(s.ready)
		close(s.queue)
		return
	}

initialization:
	for {
		select {
		case op := <-s.queue:
			op()
		default:
			close(s.ready)
			break initialization
		}
	}

	pool.Go(func(ctx context.Context) {
		for {
			select {
			case op, ok := <-s.queue:
				if !ok {
					return
				}
				op()
			case <-ctx.Done():
				close(s.queue)

				if !s.cfg.InMemory {
					indexPath := getSearchIndexPath(s.cfg)
					if indexPath != "" {
						_ = os.RemoveAll(indexPath)
					}
				}

				return
			}
		}
	})
}

func (s *SearchIndex) Add(id string, data any) {
	if !s.cfg.Enabled {
		return
	}
	s.queue <- func() {
		s.add(id, data)
	}
}

func (s *SearchIndex) add(id string, data any) {
	if s.idx == nil {
		return
	}
	err := s.idx.Index(id, data)
	if err != nil {
		log.Errorf("add '%s' to search index failed: %v", id, err)
	}
}

func (s *SearchIndex) Delete(id string) {
	if !s.cfg.Enabled {
		return
	}
	s.queue <- func() {
		_ = s.idx.Delete(id)
	}
}

func (s *SearchIndex) Search(r search.Request) (search.Result, error) {
	result := search.Result{}

	<-s.ready

	if s.idx == nil || !s.cfg.Enabled {
		return result, &search.ErrNotEnabled{}
	}

	queryText, params := parseQuery(r.QueryText)

	var clauses []query.Query
	if queryText == "" {
		clauses = append(clauses, bleve.NewMatchAllQuery())
	} else {
		q := bleve.NewQueryStringQuery(queryText)
		clauses = append(clauses, q)
	}

	for k, v := range params {
		term := bleve.NewMatchPhraseQuery(v)
		term.SetField(k)
		clauses = append(clauses, term)
	}

	qFacetsValues := make([]query.Query, len(clauses))
	copy(qFacetsValues, clauses)
	for name, val := range r.Facets {
		facet := bleve.NewMatchPhraseQuery(val)
		facet.SetField(name)
		clauses = append(clauses, facet)
	}

	q := bleve.NewConjunctionQuery(clauses...)
	sr := bleve.NewSearchRequest(q)
	sr.Size = r.Limit
	sr.From = r.Limit * r.Index
	sr.SortBy([]string{"-_score", "_id"})
	sr.Highlight = bleve.NewHighlightWithStyle(html.Name)

	searchResult, err := s.idx.Search(sr)
	if err != nil {
		return result, err
	}

	result.Total = searchResult.Total
	for _, hit := range searchResult.Hits {
		item := search.ResultItem{}

		doc, err := s.idx.Document(hit.ID)
		if err != nil {
			return result, err
		}
		if doc == nil {
			continue
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
		case "kafka":
			item, err = getKafkaSearchResult(fields, discriminators)
		case "mail":
			item, err = getMailSearchResult(fields, discriminators)
		case "ldap":
			item, err = getLdapSearchResult(fields, discriminators)
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

	// get facet values
	q = bleve.NewConjunctionQuery(qFacetsValues...)
	sr = bleve.NewSearchRequest(q)
	sr.Size = 0
	sr.AddFacet("type", bleve.NewFacetRequest("type", 6))
	searchResult, err = s.idx.Search(sr)
	if err != nil {
		return result, err
	}
	result.Facets = getFacets(searchResult)

	return result, nil
}

func getSearchFields(doc index.Document) map[string]string {
	m := make(map[string]string)
	doc.VisitFields(func(field index.Field) {
		m[field.Name()] = string(field.Value())
	})
	return m
}

func parseQuery(query string) (string, map[string]string) {
	re := regexp.MustCompile(`([\w.]+):("[^"]+"|\S+)`)

	params := make(map[string]string)
	matches := re.FindAllStringSubmatch(query, -1)

	s := query
	for _, m := range matches {
		key := m[1]
		if !slices.Contains(fieldsNotIncludedInAll, key) {
			continue
		}
		value := strings.Trim(m[2], `"`)
		params[key] = value
		s = strings.Replace(s, m[0], "", 1)
	}

	s = strings.TrimSpace(s)
	return s, params
}

func getFacets(sr *bleve.SearchResult) map[string][]search.FacetValue {
	if sr.Facets == nil {
		return nil
	}

	m := make(map[string][]search.FacetValue)
	for name, facet := range sr.Facets {
		var selectFunc func(*bleveSearch.TermFacet) search.FacetValue
		switch name {
		case "type":
			selectFunc = getTypeFacet
		default:
			log.Warnf("unknown facet: %s", name)
			continue
		}
		for _, term := range facet.Terms.Terms() {
			m[name] = append(m[name], selectFunc(term))
		}
	}
	return m
}

func getTypeFacet(term *bleveSearch.TermFacet) search.FacetValue {
	facet := search.FacetValue{Count: term.Count}
	switch term.Term {
	case "http":
		facet.Value = "HTTP"
	case "kafka":
		facet.Value = "Kafka"
	case "mail":
		facet.Value = "Mail"
	case "event":
		facet.Value = "Event"
	case "config":
		facet.Value = "Config"
	default:
		log.Errorf("unknown facet type: %s", term.Term)
		facet.Value = term.Term
	}
	return facet
}

func getSearchIndexPath(cfg static.Search) string {
	if cfg.InMemory {
		return ""
	}

	indexPath := cfg.IndexPath
	if indexPath == "" {
		indexPath = os.TempDir()
	}
	return filepath.Join(indexPath, "mokapi-bleve-index")
}
