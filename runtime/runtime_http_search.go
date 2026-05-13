package runtime

import (
	"fmt"
	"maps"
	"mokapi/providers/openapi"
	openApiSchema "mokapi/providers/openapi/schema"
	"mokapi/runtime/search"
	"mokapi/schema/json/schema"
	"slices"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

type httpSearchIndexData struct {
	Type          string            `json:"type"`
	Discriminator string            `json:"discriminator"`
	Api           string            `json:"api"`
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Description   string            `json:"description"`
	Contact       *openapi.Contact  `json:"contact"`
	Servers       []*openapi.Server `json:"servers"`
}

type httpPathSearchIndexData struct {
	Type          string                         `json:"type"`
	Discriminator string                         `json:"discriminator"`
	Api           string                         `json:"api"`
	Path          string                         `json:"path"`
	Summary       string                         `json:"summary"`
	Description   string                         `json:"description"`
	Parameters    []httpParameterSearchIndexData `json:"parameters"`
	Meta          map[string]string              `json:"meta"`
}

type httpParameterSearchIndexData struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Location    string            `json:"location"`
	Schema      *schema.IndexData `json:"schema"`
}

type httpOperationSearchIndexData struct {
	Type           string                           `json:"type"`
	Discriminator  string                           `json:"discriminator"`
	Api            string                           `json:"api"`
	Path           string                           `json:"path"`
	Method         string                           `json:"method"`
	Summary        string                           `json:"summary"`
	Description    string                           `json:"description"`
	OperationId    string                           `json:"operationId"`
	Tags           []string                         `json:"tags"`
	Parameters     []httpParameterSearchIndexData   `json:"parameters"`
	StatusCode     int                              `json:"statusCode"`
	StatusCodeText string                           `json:"statusCodeText"`
	RequestBodies  []httpRequestBodySearchIndexData `json:"requestBodies"`
	Responses      []httpResponseSearchIndexData    `json:"responses"`
}

type httpRequestBodySearchIndexData struct {
	Description string            `json:"description"`
	ContentType string            `json:"contentType"`
	Schema      *schema.IndexData `json:"schema"`
}

type httpResponseSearchIndexData struct {
	Description string            `json:"description"`
	ContentType string            `json:"contentType"`
	Schema      *schema.IndexData `json:"schema"`
}

func (s *HttpStore) addToIndex(cfg *openapi.Config) {
	if cfg == nil || cfg.Info.Name == "" {
		return
	}

	c := httpSearchIndexData{
		Type:          "http",
		Discriminator: "http",
		Api:           cfg.Info.Name,
		Name:          cfg.Info.Name,
		Version:       cfg.Info.Version,
		Description:   cfg.Info.Description,
		Contact:       cfg.Info.Contact,
		Servers:       cfg.Servers,
	}

	s.index.Add(fmt.Sprintf("http_%s", cfg.Info.Name), c)

	for path, p := range cfg.Paths {
		if p.Value == nil {
			continue
		}
		pathData := httpPathSearchIndexData{
			Type:          "http",
			Discriminator: "http_path",
			Api:           cfg.Info.Name,
			Path:          path,
			Summary:       p.Summary,
			Description:   p.Description,
			Meta:          map[string]string{},
		}
		if pathData.Summary == "" {
			pathData.Summary = p.Value.Summary
		}
		if pathData.Description == "" {
			pathData.Description = p.Value.Description
		}
		for _, param := range p.Value.Parameters {
			ps := openApiSchema.ConvertToJsonSchema(param.Value.Schema)

			pathData.Parameters = append(pathData.Parameters, httpParameterSearchIndexData{
				Name:        param.Value.Name,
				Description: param.Value.Description,
				Location:    param.Value.Type.String(),
				Schema:      schema.NewIndexData(ps),
			})
		}
		methods := slices.Collect(maps.Keys(p.Value.Operations()))
		pathData.Meta["methods"] = strings.Join(methods, ",")

		s.index.Add(fmt.Sprintf("http_%s_%s", cfg.Info.Name, path), pathData)

		for method, op := range p.Value.Operations() {
			id := fmt.Sprintf("http_%s_%s_%s", cfg.Info.Name, path, method)

			params := pathData.Parameters
			for _, param := range op.Parameters {
				ps := openApiSchema.ConvertToJsonSchema(param.Value.Schema)

				params = append(params, httpParameterSearchIndexData{
					Name:        param.Value.Name,
					Description: param.Value.Description,
					Location:    param.Value.Type.String(),
					Schema:      schema.NewIndexData(ps),
				})
			}

			var requestBodies []httpRequestBodySearchIndexData
			if op.RequestBody != nil && op.RequestBody.Value != nil {
				v := op.RequestBody.Value
				for ct, mt := range v.Content {
					rs := openApiSchema.ConvertToJsonSchema(mt.Schema)
					requestBodies = append(requestBodies, httpRequestBodySearchIndexData{
						Description: v.Description,
						ContentType: ct,
						Schema:      schema.NewIndexData(rs),
					})
				}
			}

			if op.Responses != nil && op.Responses.Len() > 0 {
				for it := op.Responses.Iter(); it.Next(); {
					v := it.Value().Value
					if v == nil {
						continue
					}
					statusCode := 0
					if i, err := strconv.Atoi(it.Key()); err == nil {
						statusCode = i
					}

					var responses []httpResponseSearchIndexData
					for ct, mt := range v.Content {
						rs := openApiSchema.ConvertToJsonSchema(mt.Schema)
						responses = append(responses, httpResponseSearchIndexData{
							Description: v.Description,
							ContentType: ct,
							Schema:      schema.NewIndexData(rs),
						})
					}

					opData := httpOperationSearchIndexData{
						Type:           "http",
						Discriminator:  "http_operation",
						Api:            cfg.Info.Name,
						Path:           path,
						Method:         method,
						Summary:        op.Summary,
						Description:    op.Description,
						OperationId:    op.OperationId,
						Tags:           op.Tags,
						Parameters:     params,
						StatusCode:     statusCode,
						StatusCodeText: it.Key(),
						RequestBodies:  requestBodies,
						Responses:      responses,
					}

					s.index.Add(id, opData)
				}
			} else {
				opData := httpOperationSearchIndexData{
					Type:          "http",
					Discriminator: "http_operation",
					Api:           cfg.Info.Name,
					Path:          path,
					Method:        method,
					Summary:       op.Summary,
					Description:   op.Description,
					OperationId:   op.OperationId,
					Tags:          op.Tags,
					Parameters:    params,
					RequestBodies: requestBodies,
				}
				s.index.Add(id, opData)
			}
		}
	}
}

func getHttpSearchResult(fields map[string]string, discriminator []string) (search.ResultItem, error) {
	result := search.ResultItem{
		Type: "HTTP",
	}

	if len(discriminator) == 1 {
		result.Title = fields["name"]
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Title,
		}
		return result, nil
	}

	switch discriminator[1] {
	case "path":
		result.Domain = fields["api"]
		result.Title = fields["path"]
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Domain,
			"path":    fields["path"],
			"methods": fields["meta.methods"],
		}
	case "operation":
		result.Domain = fields["api"]
		result.Title = fields["path"]
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Domain,
			"path":    fields["path"],
			"method":  strings.ToUpper(fields["method"]),
		}
		if s, ok := fields["statusCode"]; ok && s != "0" {
			result.Params["statusCode"] = s
		}
	default:
		return result, fmt.Errorf("unsupported search result: %s", strings.Join(discriminator, "_"))
	}
	return result, nil
}

func (s *HttpStore) removeFromIndex(cfg *openapi.Config) {
	s.index.Delete(fmt.Sprintf("http_%s", cfg.Info.Name))

	for path, p := range cfg.Paths {
		s.index.Delete(fmt.Sprintf("http_%s_%s", cfg.Info.Name, path))
		for method := range p.Value.Operations() {
			s.index.Delete(fmt.Sprintf("http_%s_%s_%s", cfg.Info.Name, path, method))
		}
	}
}

func AddMappings(m *mapping.DocumentMapping) {
	// Statt tief zu verschachteln, mappe den Pfad direkt:
	statusFieldMapping := bleve.NewNumericFieldMapping()
	statusFieldMapping.Name = "statusCode"
	statusFieldMapping.Store = true

	// Direkt auf das oberste Document Mapping anwenden
	// Bleve sucht im JSON trotzdem nach dem Pfad, indiziert ihn aber flach
	m.AddFieldMappingsAt("event.data.response.statusCode", statusFieldMapping)
}
