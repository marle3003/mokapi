package swagger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema"
	"mokapi/sortedmap"
	"net/http"
	"strconv"
)

type Config struct {
	Swagger      string                    `yaml:"swagger" json:"swagger"`
	Info         openapi.Info              `yaml:"info" json:"info"`
	Schemes      []string                  `yaml:"schemes,omitempty" json:"schemes,omitempty"`
	Consumes     []string                  `yaml:"consumes,omitempty" json:"consumes,omitempty"`
	Produces     []string                  `yaml:"produces,omitempty" json:"produces,omitempty"`
	Host         string                    `yaml:"host,omitempty" json:"host,omitempty"`
	BasePath     string                    `yaml:"basePath,omitempty" json:"basePath,omitempty"`
	Paths        PathItems                 `yaml:"paths,omitempty" json:"paths,omitempty"`
	Definitions  map[string]*schema.Schema `yaml:"definitions,omitempty" json:"definitions,omitempty"`
	Parameters   map[string]*Parameter     `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Responses    map[string]*Response      `yaml:"responses,omitempty" json:"responses,omitempty"`
	ExternalDocs *openapi.ExternalDocs     `yaml:"externalDocs,omitempty" json:"externalDocs,omitempty"`
}

type PathItems map[string]*PathItem

type PathItem struct {
	Ref        string     `yaml:"$ref,omitempty" json:"$ref,omitempty"`
	Delete     *Operation `yaml:"delete,omitempty" json:"delete,omitempty"`
	Get        *Operation `yaml:"get,omitempty" json:"get,omitempty"`
	Head       *Operation `yaml:"head,omitempty" json:"head,omitempty"`
	Options    *Operation `yaml:"options,omitempty" json:"options,omitempty"`
	Patch      *Operation `yaml:"patch,omitempty" json:"patch,omitempty"`
	Post       *Operation `yaml:"post,omitempty" json:"post,omitempty"`
	Put        *Operation `yaml:"put,omitempty" json:"put,omitempty"`
	Parameters Parameters `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

type Operation struct {
	Summary     string     `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string     `yaml:"description,omitempty" json:"description,omitempty"`
	Tags        []string   `yaml:"tags,omitempty" json:"tags,omitempty"`
	OperationID string     `yaml:"operationId,omitempty" json:"operationId,omitempty"`
	Deprecated  bool       `yaml:"deprecated" json:"deprecated"`
	Parameters  Parameters `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Responses   *Responses `yaml:"responses" json:"responses"`
	Consumes    []string   `yaml:"consumes,omitempty" json:"consumes,omitempty"`
	Produces    []string   `yaml:"produces,omitempty" json:"produces,omitempty"`
	Schemes     []string   `yaml:"schemes,omitempty" json:"schemes,omitempty"`
}

type Responses struct {
	sortedmap.LinkedHashMap[string, *Response]
} // map[HttpStatus]*ResponseRef

type Response struct {
	Ref         string                 `yaml:"$ref,omitempty" json:"$ref,omitempty"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Schema      *schema.Schema         `yaml:"schema,omitempty" json:"schema,omitempty"`
	Headers     map[string]*Header     `yaml:"headers,omitempty" json:"headers,omitempty"`
	Examples    map[string]interface{} `yaml:"examples,omitempty" json:"examples,omitempty"`
}

type Parameters []*Parameter

type Parameter struct {
	Ref              string         `yaml:"$ref,omitempty" json:"$ref,omitempty"`
	In               string         `yaml:"in,omitempty" json:"in,omitempty"`
	Name             string         `yaml:"name,omitempty" json:"name,omitempty"`
	Description      string         `yaml:"description,omitempty" json:"description,omitempty"`
	CollectionFormat string         `yaml:"collectionFormat,omitempty" json:"collectionFormat,omitempty"`
	Type             string         `yaml:"type,omitempty" json:"type,omitempty"`
	Format           string         `yaml:"format,omitempty" json:"format,omitempty"`
	Pattern          string         `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	AllowEmptyValue  bool           `yaml:"allowEmptyValue,omitempty" json:"allowEmptyValue,omitempty"`
	Required         bool           `yaml:"required,omitempty" json:"required,omitempty"`
	Deprecated       bool           `yaml:"deprecated" json:"deprecated"`
	UniqueItems      bool           `yaml:"uniqueItems,omitempty" json:"uniqueItems,omitempty"`
	ExclusiveMin     bool           `yaml:"exclusiveMinimum,omitempty" json:"exclusiveMinimum,omitempty"`
	ExclusiveMax     bool           `yaml:"exclusiveMaximum,omitempty" json:"exclusiveMaximum,omitempty"`
	Schema           *schema.Schema `yaml:"schema,omitempty" json:"schema,omitempty"`
	Items            *schema.Schema `yaml:"items,omitempty" json:"items,omitempty"`
	Enum             []interface{}  `yaml:"enum,omitempty" json:"enum,omitempty"`
	MultipleOf       *float64       `yaml:"multipleOf,omitempty" json:"multipleOf,omitempty"`
	Minimum          *float64       `yaml:"minimum,omitempty" json:"minimum,omitempty"`
	Maximum          *float64       `yaml:"maximum,omitempty" json:"maximum,omitempty"`
	MaxLength        *uint64        `yaml:"maxLength,omitempty" json:"maxLength,omitempty"`
	MaxItems         *int           `yaml:"maxItems,omitempty" json:"maxItems,omitempty"`
	MinLength        int64          `yaml:"minLength,omitempty" json:"minLength,omitempty"`
	MinItems         int            `yaml:"minItems,omitempty" json:"minItems,omitempty"`
	Default          interface{}    `yaml:"default,omitempty" json:"default,omitempty"`
}

type Header struct {
	Parameter
}

func (p *PathItem) Operations() map[string]*Operation {
	operations := make(map[string]*Operation, 7)
	if p.Get != nil {
		operations[http.MethodGet] = p.Get
	}
	if p.Post != nil {
		operations[http.MethodPost] = p.Post
	}
	if p.Put != nil {
		operations[http.MethodPut] = p.Put
	}
	if p.Patch != nil {
		operations[http.MethodPatch] = p.Patch
	}
	if p.Head != nil {
		operations[http.MethodHead] = p.Head
	}
	if p.Delete != nil {
		operations[http.MethodDelete] = p.Delete
	}
	if p.Options != nil {
		operations[http.MethodOptions] = p.Options
	}
	if p.Post != nil {
		operations[http.MethodPost] = p.Post
	}
	return operations
}

func (p PathItems) Resolve(token string) (interface{}, error) {
	if v, ok := p["/"+token]; ok {
		return v, nil
	}
	return nil, nil
}

func (r *Responses) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("expected openapi.Responses map, got %s", token)
	}
	r.LinkedHashMap = sortedmap.LinkedHashMap[string, *Response]{}
	for {
		token, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}
		key := token.(string)
		val := &Response{}
		err = dec.Decode(&val)
		if err != nil {
			return err
		}
		switch m := any(&r.LinkedHashMap).(type) {
		case *sortedmap.LinkedHashMap[string, *Response]:
			m.Set(key, val)
		case *sortedmap.LinkedHashMap[int, *Response]:
			if key == "default" {
				m.Set(0, val)
			} else {
				statusCode, err := strconv.Atoi(key)
				if err != nil {
					return fmt.Errorf("unable to parse http status %v", key)
				}
				m.Set(statusCode, val)
			}
		}
	}
}

func (r *Responses) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected openapi.Responses map, got %v", value.Tag)
	}
	r.LinkedHashMap = sortedmap.LinkedHashMap[string, *Response]{}
	for i := 0; i < len(value.Content); i += 2 {
		var key string
		err := value.Content[i].Decode(&key)
		if err != nil {
			return err
		}
		val := &Response{}
		err = value.Content[i+1].Decode(&val)
		if err != nil {
			return err
		}
		switch m := any(&r.LinkedHashMap).(type) {
		case *sortedmap.LinkedHashMap[string, *Response]:
			m.Set(key, val)
		case *sortedmap.LinkedHashMap[int, *Response]:
			if key == "default" {
				m.Set(0, val)
			} else {
				statusCode, err := strconv.Atoi(key)
				if err != nil {
					return fmt.Errorf("unable to parse http status %v", key)
				}
				m.Set(statusCode, val)
			}
		}
	}

	return nil
}

func (c *Config) UnmarshalJSON(b []byte) error {
	type alias Config
	a := alias(*c)
	err := dynamic.UnmarshalJSON(b, &a)
	if err != nil {
		return err
	}
	*c = Config(a)
	return nil
}
