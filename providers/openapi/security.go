package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net/http"
)

type SecuritySchemes map[string]SecurityScheme

type SecurityScheme interface {
	Serve(req *http.Request) error
}

type SecurityRequirement map[string][]string

type HttpSecurityScheme struct {
	Type         string `yaml:"type" json:"type"`
	Description  string `yaml:"description" json:"description"`
	Scheme       string `yaml:"scheme" json:"scheme"`
	BearerFormat string `yaml:"bearerFormat" json:"bearerFormat"`
}

func (s *HttpSecurityScheme) Serve(req *http.Request) error {
	request := EventRequestFromContext(req.Context())
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return fmt.Errorf("no authorization header")
	}

	switch s.Scheme {
	case "bearer":
		request.Header["Authorization"] = auth[len("Bearer "):]
	case "basic":
		request.Header["Authorization"] = auth[len("Basic "):]
	default:
		return fmt.Errorf("security scheme not supported: %v", s.Scheme)
	}

	return nil
}

type ApiKeySecurityScheme struct {
	Type string `yaml:"type" json:"type"`
	In   string `yaml:"in" json:"in"`
	Name string `yaml:"name" json:"name"`
}

func (s *ApiKeySecurityScheme) Serve(req *http.Request) error {
	request := EventRequestFromContext(req.Context())
	switch s.In {
	case "header":
		auth := req.Header.Get(s.Name)
		if auth == "" {
			return fmt.Errorf("missing header for API Key: %v", s.Name)
		}
		request.Header[s.Name] = auth
	case "query":
		q := req.URL.Query()
		if !q.Has(s.Name) {
			return fmt.Errorf("no API key in query: %v", s.Name)
		}
		key := req.URL.Query().Get(s.Name)
		request.Query[s.Name] = key
	case "cookie":
		c, err := req.Cookie(s.Name)
		if errors.Is(err, http.ErrNoCookie) {
			return fmt.Errorf("no API key in cookie: %v", s.Name)
		}
		request.Cookie[s.Name] = c.Value
	default:
		return fmt.Errorf("security scheme API Key in not supported: %v", s.In)
	}
	return nil
}

type NotSupportedSecurityScheme struct {
	Type string `yaml:"type" json:"type"`
}

func (s *NotSupportedSecurityScheme) Serve(req *http.Request) error {
	log.Warnf("security scheme %v not supported", s.Type)
	return nil
}

func (s *SecuritySchemes) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("unexpected token %s; expected '{'", token)
	}
	for {
		token, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}
		key := token.(string)

		if delim, ok := token.(json.Delim); ok && delim != '{' {
			return fmt.Errorf("unexpected token %s; expected '{'", token)
		}

		m := map[string]interface{}{}
		err = dec.Decode(&m)
		if err != nil {
			return err
		}
		jsonString, err := json.Marshal(m)
		if err != nil {
			return err
		}
		var v SecurityScheme
		switch m["type"] {
		case "http":
			v = &HttpSecurityScheme{}
		case "apiKey":
			v = &ApiKeySecurityScheme{}
		default:
			v = &NotSupportedSecurityScheme{}
		}
		err = json.Unmarshal(jsonString, v)
		if err != nil {
			return err
		}

		if *s == nil {
			*s = map[string]SecurityScheme{}
		}

		(*s)[key] = v
	}
}

func (s *SecuritySchemes) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping")
	}

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		var values map[string]string
		_ = node.Content[i+1].Decode(&values)

		if *s == nil {
			*s = map[string]SecurityScheme{}
		}

		var v SecurityScheme
		switch values["type"] {
		case "http":
			v = &HttpSecurityScheme{}
		case "apiKey":
			v = &ApiKeySecurityScheme{}
		default:
			v = &NotSupportedSecurityScheme{}
		}
		(*s)[key] = v
		err := node.Content[i+1].Decode(v)
		if err != nil {
			return err
		}
	}

	return nil
}
