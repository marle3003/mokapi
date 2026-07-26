package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/json/parser"
	"net/http"
	"net/url"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type PathItems map[string]*PathRef

type PathRef struct {
	dynamic.Reference[*PathRef]
	Value *Path
}

type Path struct {
	// An optional, string summary, intended to apply to all operations
	// in this path.
	Summary string `yaml:"summary,omitempty" json:"summary,omitempty"`

	// An optional, string description, intended to apply to all operations
	// in this path. CommonMark syntax MAY be used for rich text representation.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// A definition of a GET operation on this path.
	Get *Operation `yaml:"get,omitempty" json:"get,omitempty"`

	// A definition of a POST operation on this path.
	Post *Operation `yaml:"post,omitempty" json:"post,omitempty"`

	// A definition of a PUT operation on this path.
	Put *Operation `yaml:"put,omitempty" json:"put,omitempty"`

	// A definition of a PATCH operation on this path.
	Patch *Operation `yaml:"patch,omitempty" json:"patch,omitempty"`

	// A definition of a DELETE operation on this path.
	Delete *Operation `yaml:"delete,omitempty" json:"delete,omitempty"`

	// A definition of a HEAD operation on this path.
	Head *Operation `yaml:"head,omitempty" json:"head,omitempty"`

	// A definition of an OPTIONS operation on this path.
	Options *Operation `yaml:"options,omitempty" json:"options,omitempty"`

	// A definition of a TRACE operation on this path.
	Trace *Operation `yaml:"trace,omitempty" json:"trace,omitempty"`

	Query *Operation `yaml:"query,omitempty" json:"query,omitempty"`

	AdditionalOperations map[string]*Operation `yaml:"additionalOperations,omitempty" json:"additionalOperations,omitempty"`

	// A list of parameters that are applicable for all
	// the operations described under this path. These
	// parameters can be overridden at the operation level,
	// but cannot be removed there
	Parameters Parameters `yaml:"parameters,omitempty" json:"parameters,omitempty"`

	Path   string  `yaml:"-" json:"-"`
	Status Status  `yaml:"-" json:"-"`
	Errors []Error `yaml:"-" json:"-"`
}

func (r *PathRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *PathRef) MarshalJSON() ([]byte, error) {
	if r.Value != nil {
		return json.Marshal(r.Value)
	} else {
		return json.Marshal(r.Ref)
	}
}

func (r *PathRef) MarshalYAML() (any, error) {
	if r.Value != nil {
		return r.Value, nil
	}
	return r.Ref, nil
}

func (r *PathRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (p *Path) Operations() map[string]*Operation {
	m := make(map[string]*Operation)
	for name, op := range p.AdditionalOperations {
		m[name] = op
	}

	if p.Get != nil {
		m[http.MethodGet] = p.Get
	}
	if p.Post != nil {
		m[http.MethodPost] = p.Post
	}
	if p.Put != nil {
		m[http.MethodPut] = p.Put
	}
	if p.Patch != nil {
		m[http.MethodPatch] = p.Patch
	}
	if p.Delete != nil {
		m[http.MethodDelete] = p.Delete
	}
	if p.Head != nil {
		m[http.MethodHead] = p.Head
	}
	if p.Options != nil {
		m[http.MethodOptions] = p.Options
	}
	if p.Trace != nil {
		m[http.MethodTrace] = p.Trace
	}
	if p.Query != nil {
		m["QUERY"] = p.Query
	}
	return m
}

func (p *Path) Operation(method string) *Operation {
	method = strings.ToUpper(method)
	switch method {
	case http.MethodGet:
		return p.Get
	case http.MethodPost:
		return p.Post
	case http.MethodPut:
		return p.Put
	case http.MethodPatch:
		return p.Patch
	case http.MethodDelete:
		return p.Delete
	case http.MethodHead:
		return p.Head
	case http.MethodOptions:
		return p.Options
	case http.MethodTrace:
		return p.Trace
	case "QUERY":
		return p.Query
	default:
		if op, ok := p.AdditionalOperations[method]; ok {
			return op
		}
	}

	return nil
}

func (p PathItems) Resolve(token string) (interface{}, error) {
	if v, ok := p["/"+token]; ok {
		return v, nil
	}
	if v, ok := p[token]; ok {
		return v, nil
	}
	return nil, nil
}

func (p PathItems) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if p == nil {
		return nil
	}

	for name, e := range p {
		if e == nil {
			continue
		}
		if err := e.Parse(config, reader); err != nil {
			return fmt.Errorf("parse path '%v' failed: %w", name, err)
		}
		if e.Value != nil {
			e.Value.Path = name
		}
	}
	return nil
}

func (r *PathRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		resolved, err := r.Resolve(config, reader)
		if err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
	}

	return r.Value.Parse(config, reader)
}

func (p *Path) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if p == nil {
		return nil
	}

	for index, param := range p.Parameters {
		if err := param.Parse(config, reader); err != nil {
			return fmt.Errorf("parse parameter '%v' failed: %w", index, err)
		}
	}

	for method, op := range p.Operations() {
		err := op.Parse(config, reader)
		if err != nil {
			op.Status = StatusInvalid
			method = strings.ToUpper(method)
			op.Errors = append(op.Errors, Error{Message: err.Error()})
			log.
				WithField("api", getName(config)).
				WithField("method", method).
				WithField("path", p.Path).
				WithField("namespace", "http").
				Error(err)
		}
		op.Path = p
	}

	for name, op := range p.AdditionalOperations {
		if err := op.Parse(config, reader); err != nil {
			op.Status = StatusInvalid
			name = strings.ToUpper(name)
			log.
				WithField("api", getName(config)).
				WithField("method", name).
				WithField("path", p.Path).
				WithField("namespace", "http").
				Error(err)
		} else {
			op.Path = p
		}
	}

	return nil
}

func (p PathItems) patch(patch PathItems) {
	for path, v := range patch {
		if r, ok := p[path]; ok && r != nil {
			r.patch(v)
		} else {
			p[path] = v
		}
	}
}

func (r *PathRef) patch(patch *PathRef) {
	if patch == nil || patch.Value == nil {
		return
	}

	if r.Value == nil {
		r.Value = patch.Value
		return
	} else {
		r.Value.patch(patch.Value)
	}
}

func (p *Path) patch(patch *Path) {
	if p == nil || patch == nil {
		return
	}

	if len(patch.Summary) > 0 {
		p.Summary = patch.Summary
	}

	if len(patch.Description) > 0 {
		p.Description = patch.Description
	}

	if p.Get == nil {
		p.Get = patch.Get
	} else {
		p.Get.patch(patch.Get)
	}

	if p.Post == nil {
		p.Post = patch.Post
	} else {
		p.Post.patch(patch.Post)
	}

	if p.Put == nil {
		p.Put = patch.Put
	} else {
		p.Put.patch(patch.Put)
	}

	if p.Patch == nil {
		p.Patch = patch.Patch
	} else {
		p.Patch.patch(patch.Patch)
	}

	if p.Delete == nil {
		p.Delete = patch.Delete
	} else {
		p.Delete.patch(patch.Delete)
	}

	if p.Head == nil {
		p.Head = patch.Head
	} else {
		p.Head.patch(patch.Head)
	}

	if p.Options == nil {
		p.Options = patch.Options
	} else {
		p.Options.patch(patch.Options)
	}

	if p.Trace == nil {
		p.Trace = patch.Trace
	} else {
		p.Trace.patch(patch.Trace)
	}

	p.Parameters.Patch(patch.Parameters)
}

func (p *Path) RunWebhook(urlString string, args common.WebhookArgs) (*common.WebhookResponse, error) {
	method := args.Method
	ops := p.Operations()

	if len(ops) == 0 {
		return nil, fmt.Errorf("no operations specified")
	}

	if method == "" {
		methods := slices.Collect(maps.Keys(ops))
		switch len(methods) {
		case 1:
			method = methods[0]
		default:
			return nil, fmt.Errorf("multiple operations specified: use args.method to refine")
		}
	}

	o, ok := ops[strings.ToUpper(method)]
	if !ok {
		return nil, fmt.Errorf("method %s not found", method)
	}

	u, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url '%v': %w", urlString, err)
	}

	r := &http.Request{Method: method, URL: u}

	r.Header, err = parseRequestHeader(args, o)
	if err != nil {
		return nil, err
	}

	r.Body, err = parseRequestBody(args, o)
	if err != nil {
		return nil, err
	}

	c := &http.Client{Timeout: args.Timeout}
	res, err := c.Do(r)
	if err != nil {
		return nil, err
	}

	resp := o.getResponse(res.StatusCode)
	if resp == nil {
		return nil, fmt.Errorf("no response for '%v' specified", res.StatusCode)
	}

	result := &common.WebhookResponse{
		StatusCode: res.StatusCode,
		Data:       nil,
		Headers:    map[string]any{},
	}

	if res.Body != http.NoBody {
		defer func() { _ = res.Body.Close() }()

		ct := media.ParseContentType(res.Header.Get("Content-Type"))
		result.Headers["Content-Type"] = ct.String()

		mt := resp.GetContent(ct)
		if mt == nil {
			return nil, fmt.Errorf("content type '%s' for '%v' not specified", ct, res.StatusCode)
		}

		var b []byte
		b, err = io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		var body *Body
		body, err = parseBody(b, ct, mt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response body: %w", err)
		}
		if body != nil {
			result.Data = body.Value
		}
	}

	for name, ref := range resp.Headers {
		if ref.Value == nil {
			continue
		}
		v, err := parseHeader(new(ref.Value.Parameter), res.Header)
		if err != nil {
			return nil, fmt.Errorf("failed to parse header '%v': %w", name, err)
		}
		if v != nil {
			result.Headers[name] = v.Value
		} else {
			result.Headers[name] = nil
		}
	}

	return result, nil
}

func parseRequestHeader(args common.WebhookArgs, o *Operation) (http.Header, error) {
	header := http.Header{}
	params := o.Parameters
	if o.Path != nil && o.Path.Parameters != nil {
		params = append(params, o.Path.Parameters...)
	}
	for _, refParam := range o.Parameters {
		param := refParam.Value
		if param == nil {
			continue
		}

		if param.Type != ParameterHeader {
			continue
		}

		s, ok := args.Headers[param.Name]
		if !ok {
			if param.Required {
				return nil, fmt.Errorf("required header parameter %s not found", param.Name)
			}
			continue
		}

		ps := parser.Parser{Schema: schema.ConvertToJsonSchema(param.Schema)}
		v, err := ps.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse header parameter %s: %w", param.Name, err)
		}
		header.Set(param.Name, fmt.Sprintf("%v", v))
	}
	return header, nil
}

func parseRequestBody(args common.WebhookArgs, o *Operation) (io.ReadCloser, error) {
	if o.RequestBody != nil && o.RequestBody.Value != nil {
		rb := o.RequestBody.Value
		if rb.Required && args.Body == "" && args.Data == nil {
			return nil, fmt.Errorf("request body is required")
		}

		if args.Body != "" {
			r := io.NopCloser(bytes.NewReader([]byte(args.Body)))
			return r, nil
		}

		ct, err := getContentType(args.Headers)
		if err != nil {
			return nil, err
		}
		if ct == "" {
			values := slices.Collect(maps.Keys(rb.Content))
			switch len(values) {
			case 0:
				return nil, fmt.Errorf("request body content is not specified")
			case 1:
				ct = values[0]
			default:
				return nil, fmt.Errorf("multiple request body contents specified: use args.headers to refine")
			}
		}
		contentType := media.ParseContentType(ct)
		c := rb.Content[ct]
		if c == nil {
			return nil, fmt.Errorf("request body not specified for content type %s", contentType)
		}
		b, err := c.Schema.Marshal(args.Data, contentType)
		if err != nil {
			return nil, err
		}
		r := io.NopCloser(bytes.NewReader([]byte(b)))
		return r, nil
	}
	return nil, nil
}
