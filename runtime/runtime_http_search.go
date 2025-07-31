package runtime

import (
	"fmt"
	"mokapi/providers/openapi"
	openApiSchema "mokapi/providers/openapi/schema"
	"mokapi/runtime/search"
	"mokapi/schema/json/schema"
	"strings"
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
}

type httpParameterSearchIndexData struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Location    string            `json:"location"`
	Schema      *schema.IndexData `json:"schema"`
}

type httpOperationSearchIndexData struct {
	Type          string                         `json:"type"`
	Discriminator string                         `json:"discriminator"`
	Api           string                         `json:"api"`
	Path          string                         `json:"path"`
	Method        string                         `json:"method"`
	Summary       string                         `json:"summary"`
	Description   string                         `json:"description"`
	OperationId   string                         `json:"operationId"`
	Tags          []string                       `json:"tags"`
	Parameters    []httpParameterSearchIndexData `json:"parameters"`
	RequestBody   string                         `json:"request_body"`
	Responses     []httpResponseSearchIndexData  `json:"responses"`
}

type httpResponseSearchIndexData struct {
	Description string            `json:"description"`
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

	add(s.index, cfg.Info.Name, c)

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

		add(s.index, fmt.Sprintf("%s_%s", cfg.Info.Name, path), pathData)

		for method, op := range p.Value.Operations() {
			id := fmt.Sprintf("%s_%s_%s", cfg.Info.Name, path, method)

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
				Parameters:    pathData.Parameters,
			}
			for _, param := range op.Parameters {
				ps := openApiSchema.ConvertToJsonSchema(param.Value.Schema)

				opData.Parameters = append(opData.Parameters, httpParameterSearchIndexData{
					Name:        param.Value.Name,
					Description: param.Value.Description,
					Location:    param.Value.Type.String(),
					Schema:      schema.NewIndexData(ps),
				})
			}

			if op.Responses != nil {
				for it := op.Responses.Iter(); it.Next(); {
					v := it.Value().Value
					if v == nil {
						continue
					}
					for _, mt := range v.Content {
						rs := openApiSchema.ConvertToJsonSchema(mt.Schema)

						opData.Responses = append(opData.Responses, httpResponseSearchIndexData{
							Description: v.Description,
							Schema:      schema.NewIndexData(rs),
						})
					}

				}
			}

			add(s.index, id, opData)
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
		}
	case "operation":
		result.Domain = fields["api"]
		result.Title = fmt.Sprintf("%s %s", fields["method"], fields["path"])
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Domain,
			"path":    fields["path"],
			"method":  strings.ToLower(fields["method"]),
		}
	default:
		return result, fmt.Errorf("unsupported search result: %s", strings.Join(discriminator, "_"))
	}
	return result, nil
}

func (s *HttpStore) removeFromIndex(cfg *openapi.Config) {
	_ = s.index.Delete(cfg.Info.Name)

	for path, p := range cfg.Paths {
		_ = s.index.Delete(fmt.Sprintf("%s_%s", cfg.Info.Name, path))
		for method := range p.Value.Operations() {
			_ = s.index.Delete(fmt.Sprintf("%s_%s_%s", cfg.Info.Name, path, method))
		}
	}
}
