package runtime

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"mokapi/providers/openapi"
	"strings"
)

type HttpConfig struct {
	Discriminator string
	Name          string
	Version       string
	Description   string
}

type HttpPath struct {
	Discriminator string
	ConfigName    string
	Path          string
	Summary       string
	Description   string
	Parameters    []HttpParameter
}

type HttpParameter struct {
	Name        string
	Description string
	Location    string
}

type HttpOperation struct {
	Discriminator string
	ConfigName    string
	Path          string
	Method        string
	Summary       string
	Description   string
	Tags          []string
	Parameters    []HttpParameter
	RequestBody   string
	Responses     string
}

func addHttpToIndex(index bleve.Index, cfg *openapi.Config) error {
	err := index.Index(cfg.Info.Name, HttpConfig{
		Discriminator: "http",
		Name:          cfg.Info.Name,
		Version:       cfg.Info.Version,
		Description:   cfg.Info.Description,
	})
	if err != nil {
		return err
	}

	for path, p := range cfg.Paths {
		if p.Value == nil {
			continue
		}
		pathData := HttpPath{
			Discriminator: "http_path",
			ConfigName:    cfg.Info.Name,
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
			pathData.Parameters = append(pathData.Parameters, HttpParameter{
				Name:        param.Value.Name,
				Description: param.Value.Description,
				Location:    param.Value.Type.String(),
			})
		}

		err = index.Index(fmt.Sprintf("%s_%s", cfg.Info.Name, path), pathData)
		if err != nil {
			return err
		}

		for method, op := range p.Value.Operations() {
			id := fmt.Sprintf("%s_%s_%s", cfg.Info.Name, path, method)

			opData := HttpOperation{
				Discriminator: "http_operation",
				ConfigName:    cfg.Info.Name,
				Path:          path,
				Method:        method,
				Summary:       op.Summary,
				Description:   op.Description,
				Tags:          op.Tags,
				Parameters:    pathData.Parameters,
			}
			for _, param := range op.Parameters {
				opData.Parameters = append(opData.Parameters, HttpParameter{
					Name:        param.Value.Name,
					Description: param.Value.Description,
					Location:    param.Value.Type.String(),
				})
			}

			err = index.Index(id, opData)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func removeHttpFromIndex(index bleve.Index, cfg *openapi.Config) {
	_ = index.Delete(cfg.Info.Name)

	for path, p := range cfg.Paths {
		_ = index.Delete(fmt.Sprintf("%s_%s", cfg.Info.Name, path))
		for method := range p.Value.Operations() {
			_ = index.Delete(fmt.Sprintf("%s_%s_%s", cfg.Info.Name, path, method))
		}
	}
}

func getHttpSearchResult(fields map[string]string, discriminator []string) (*SearchResult, error) {
	result := &SearchResult{
		Type:       "HTTP",
		ConfigName: fields["ConfigName"],
	}

	if len(discriminator) == 1 {
		result.Title = fields["Name"]
		return result, nil
	}

	switch discriminator[1] {
	case "":
	case "path":
		result.Title = fields["Path"]
	case "operation":
		result.Title = fmt.Sprintf("%s %s", fields["Method"], fields["Path"])
	default:
		return nil, fmt.Errorf("unsupported search result: %s", strings.Join(discriminator, "_"))
	}
	return result, nil
}
