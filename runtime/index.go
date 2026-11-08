package runtime

import (
	"context"
	"fmt"
	"mokapi/config/static"
	"mokapi/runtime/events"
	"mokapi/runtime/search"
	"mokapi/safe"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sync/atomic"
	"time"

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
	cfg static.Search
	// idx needs no lock: start() writes it before the worker starts and before
	// close(ready), and Search reads it only after <-ready.
	idx     bleve.Index
	enabled atomic.Bool // set to false if the index cannot be created
	ready   chan struct{}
	queue   chan indexOp
	pending atomic.Int64 // operations queued but not yet flushed to the index
}

func newSearchIndex(cfg static.Search) *SearchIndex {
	s := &SearchIndex{cfg: cfg}
	if cfg.Enabled {
		s.enabled.Store(true)
		s.ready = make(chan struct{})
		s.queue = make(chan indexOp, 1000)
		if cfg.NumIndexWorker == 0 {
			s.cfg.NumIndexWorker = 1
		}
	}
	return s
}

func (s *SearchIndex) start(pool *safe.Pool) {
	if !s.enabled.Load() {
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

	metaMapping := bleve.NewDocumentMapping()
	metaMapping.AddFieldMappingsAt("*", disableIndex)
	docMapping.AddSubDocumentMapping("meta", metaMapping)

	apiField := bleve.NewTextFieldMapping()
	apiField.Analyzer = "mokapi_analyzer"
	apiField.IncludeInAll = false // Exclude from default search
	apiField.Store = true
	apiField.Index = true
	apiField.IncludeTermVectors = true
	docMapping.AddFieldMappingsAt("api", apiField)

	// enable term vectors for all fields, allowing phrase queries (like "Swagger Petstore")
	defaultField := bleve.NewTextFieldMapping()
	defaultField.IncludeTermVectors = true
	docMapping.AddFieldMappingsAt("*", defaultField)

	AddMappings(docMapping)

	mapping := bleve.NewIndexMapping()
	mapping.DefaultMapping = docMapping
	mapping.DefaultAnalyzer = "mokapi_analyzer"

	stemmer := porter.Name

	err := mapping.AddCustomAnalyzer("mokapi_analyzer", map[string]any{
		"type":         custom.Name,
		"tokenizer":    unicode.Name,
		"char_filters": []any{asciifolding.Name},
		"token_filters": []any{
			// Bleve token filters run in sequence
			// camelcase must precede lowercase
			camelcase.Name,
			lowercase.Name,
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
		s.enabled.Store(false)
		close(s.ready)
		s.drainQueue()
		return
	}

	for i := 0; i < 1; i++ {
		pool.Go(s.runWorker)
	}

	// never finishes if the pool is cancelled before the workers drain the queue (same as the old Wait)
	for s.pending.Load() > 0 {
		time.Sleep(10 * time.Millisecond)
	}
	// ensure last batch is flushed
	time.Sleep(500 * time.Millisecond)
	close(s.ready)
}

// drainQueue drops queued operations. The queue is not closed because a
// concurrent Add may still send to it.
func (s *SearchIndex) drainQueue() {
	for {
		select {
		case <-s.queue:
		default:
			return
		}
	}
}

type opType int

const (
	opAdd opType = iota
	opDelete
)

type indexOp struct {
	id   string
	data any
	typ  opType
}

func (s *SearchIndex) Add(id string, data any) {
	if !s.enabled.Load() {
		return
	}

	s.pending.Add(1)
	s.queue <- indexOp{
		id:   id,
		data: data,
		typ:  opAdd,
	}
}

func (s *SearchIndex) runWorker(ctx context.Context) {
	batch := s.idx.NewBatch()
	batchSize := 0

	const maxBatchSize = 100
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	flush := func() {
		if batchSize == 0 {
			return
		}
		if s.idx == nil {
			return
		}

		if err := s.idx.Batch(batch); err != nil {
			println("batch failed:", err.Error())
		}

		s.pending.Add(-int64(batchSize))

		batch = s.idx.NewBatch()
		batchSize = 0
	}

	for {
		select {
		case op, ok := <-s.queue:
			if !ok {
				return
			}

			if op.typ == opAdd {
				err := batch.Index(op.id, op.data)
				if err != nil {
					log.Errorf("add '%s' to search index failed: %v", op.id, err)
				}
			} else if op.typ == opDelete {
				batch.Delete(op.id)
			}
			batchSize++

			if batchSize >= maxBatchSize {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-ctx.Done():
			return
		}
	}
}

func (s *SearchIndex) Delete(id string) {
	if !s.enabled.Load() {
		return
	}
	s.pending.Add(1)
	s.queue <- indexOp{
		id:  id,
		typ: opDelete,
	}
}

func (s *SearchIndex) Search(r search.Request) (search.Result, error) {
	result := search.Result{}

	// without this check Search would wait forever: ready only exists if enabled
	if !s.enabled.Load() {
		return result, &search.ErrNotEnabled{}
	}
	<-s.ready
	// start() may have disabled search while we were waiting
	if !s.enabled.Load() {
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

	for _, p := range params {
		term := bleve.NewMatchPhraseQuery(p.value)
		term.SetField(p.key)
		bq := bleve.NewBooleanQuery()
		switch p.operator {
		case "+":
			bq.AddMust(term)
		case "-":
			bq.AddMustNot(term)
		default:
			bq.AddShould(term)
		}
		clauses = append(clauses, bq)
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
		case "mqtt":
			item, err = getMqttSearchResult(fields, discriminators)
		case "websocket":
			item, err = getWebsocketSearchResult(fields, discriminators)
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
		var value string
		switch f := field.(type) {
		case index.NumericField:
			v, _ := f.Number()
			value = fmt.Sprintf("%v", v)
		default:
			value = string(field.Value())
		}
		m[field.Name()] = value
	})
	return m
}

type param struct {
	key      string
	value    string
	operator string
}

func parseQuery(query string) (string, []param) {
	re := regexp.MustCompile(`([+-]?)([\w.]+):("[^"]+"|\S+)`)

	matches := re.FindAllStringSubmatch(query, -1)

	var params []param
	s := query
	for _, m := range matches {
		key := m[2]
		if !slices.Contains(fieldsNotIncludedInAll, key) {
			continue
		}
		value := strings.Trim(m[3], `"`)
		params = append(params, param{key: key, value: value, operator: m[1]})
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
	case "mqtt":
		facet.Value = "MQTT"
	case "websocket":
		facet.Value = "Websocket"
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
	indexPath := cfg.IndexPath
	if indexPath == "" {
		indexPath = os.TempDir()
	}
	return filepath.Join(indexPath, "mokapi-bleve-index")
}
