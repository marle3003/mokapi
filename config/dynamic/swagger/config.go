package swagger

import (
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/schema"
	"net/http"
)

func init() {
	common.Register("swagger", &Config{})
}

type Config struct {
	Swagger     string                 `json:"swagger" yaml:"swagger"`
	Info        openapi.Info           `json:"info" yaml:"info"`
	Schemes     []string               `json:"schemes,omitempty" yaml:"schemes,omitempty"`
	Consumes    []string               `json:"consumes,omitempty" yaml:"consumes,omitempty"`
	Produces    []string               `json:"produces,omitempty" yaml:"produces,omitempty"`
	Host        string                 `json:"host,omitempty" yaml:"host,omitempty"`
	BasePath    string                 `json:"basePath,omitempty" yaml:"basePath,omitempty"`
	Paths       PathItems              `json:"paths,omitempty" yaml:"paths,omitempty"`
	Definitions map[string]*schema.Ref `json:"definitions,omitempty" yaml:"definitions,omitempty"`
	Parameters  map[string]*Parameter  `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Responses   map[string]*Response   `json:"responses,omitempty" yaml:"responses,omitempty"`
}

type PathItems map[string]*PathItem

type PathItem struct {
	Ref        string     `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Delete     *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Get        *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Head       *Operation `json:"head,omitempty" yaml:"head,omitempty"`
	Options    *Operation `json:"options,omitempty" yaml:"options,omitempty"`
	Patch      *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
	Post       *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Put        *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Parameters Parameters `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

type Operation struct {
	Summary     string               `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string             `json:"tags,omitempty" yaml:"tags,omitempty"`
	OperationID string               `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  Parameters           `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Responses   map[string]*Response `json:"responses" yaml:"responses"`
	Consumes    []string             `json:"consumes,omitempty" yaml:"consumes,omitempty"`
	Produces    []string             `json:"produces,omitempty" yaml:"produces,omitempty"`
	Schemes     []string             `json:"schemes,omitempty" yaml:"schemes,omitempty"`
}

type Response struct {
	Ref         string                 `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      *schema.Ref            `json:"schema,omitempty" yaml:"schema,omitempty"`
	Headers     map[string]*Header     `json:"headers,omitempty" yaml:"headers,omitempty"`
	Examples    map[string]interface{} `json:"examples,omitempty" yaml:"examples,omitempty"`
}

type Parameters []*Parameter

type Parameter struct {
	Ref              string        `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	In               string        `json:"in,omitempty" yaml:"in,omitempty"`
	Name             string        `json:"name,omitempty" yaml:"name,omitempty"`
	Description      string        `json:"description,omitempty" yaml:"description,omitempty"`
	CollectionFormat string        `json:"collectionFormat,omitempty" yaml:"collectionFormat,omitempty"`
	Type             string        `json:"type,omitempty" yaml:"type,omitempty"`
	Format           string        `json:"format,omitempty" yaml:"format,omitempty"`
	Pattern          string        `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	AllowEmptyValue  bool          `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Required         bool          `json:"required,omitempty" yaml:"required,omitempty"`
	UniqueItems      bool          `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	ExclusiveMin     bool          `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMax     bool          `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Schema           *schema.Ref   `json:"schema,omitempty" yaml:"schema,omitempty"`
	Items            *schema.Ref   `json:"items,omitempty" yaml:"items,omitempty"`
	Enum             []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
	MultipleOf       *float64      `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Minimum          *float64      `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum          *float64      `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	MaxLength        *uint64       `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MaxItems         *int          `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinLength        int64         `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MinItems         int           `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	Default          interface{}   `json:"default,omitempty" yaml:"default,omitempty"`
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
