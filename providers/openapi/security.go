package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"net/http"
	"strings"
)

type SecuritySchemes map[string]SecurityScheme

type SecurityScheme interface {
	Serve(req *http.Request) error
}

type SecurityRequirement map[string][]string

type HttpSecurityScheme struct {
	Type         string `yaml:"type" json:"type"`
	Scheme       string `yaml:"scheme" json:"scheme"`
	BearerFormat string `yaml:"bearerFormat" json:"bearerFormat"`
	Description  string `yaml:"description,omitempty" json:"description,omitempty"`
}

func (s *HttpSecurityScheme) Serve(req *http.Request) error {
	request := EventRequestFromContext(req.Context())
	log, ok := LogEventFromContext(req.Context())

	auth := req.Header.Get("Authorization")

	if auth == "" {
		if ok {
			log.Request.Parameters = append(log.Request.Parameters, HttpParameter{
				Name: "Authorization",
				Type: "header",
			})
		}
		return fmt.Errorf("no authorization header")
	}

	switch s.Scheme {
	case "bearer":
		request.Header["Authorization"] = auth
	case "basic":
		request.Header["Authorization"] = auth
	default:
		return fmt.Errorf("security scheme not supported: %v", s.Scheme)
	}

	if ok {
		for i, p := range log.Request.Parameters {
			if p.Name == "Authorization" {
				p.Value = auth
				log.Request.Parameters[i] = p
				break
			}
		}
	}

	return nil
}

type ApiKeySecurityScheme struct {
	Type        string `yaml:"type" json:"type"`
	In          string `yaml:"in" json:"in"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

func (s *ApiKeySecurityScheme) Serve(req *http.Request) error {
	request := EventRequestFromContext(req.Context())
	var val string
	switch s.In {
	case "header":
		auth := req.Header.Get(s.Name)
		if auth == "" {
			return fmt.Errorf("missing header for API Key: %v", s.Name)
		}
		request.Header[s.Name] = auth
		val = auth
	case "query":
		q := req.URL.Query()
		if !q.Has(s.Name) {
			return fmt.Errorf("no API key in query: %v", s.Name)
		}
		val = req.URL.Query().Get(s.Name)
		request.Query[s.Name] = val
	case "cookie":
		c, err := req.Cookie(s.Name)
		if errors.Is(err, http.ErrNoCookie) {
			return fmt.Errorf("no API key in cookie: %v", s.Name)
		}
		request.Cookie[s.Name] = c.Value
		val = c.Value
	default:
		return fmt.Errorf("security scheme API Key in not supported: %v", s.In)
	}

	if log, ok := LogEventFromContext(req.Context()); ok {
		if s.In == "query" {
			log.Request.Parameters = append(log.Request.Parameters, HttpParameter{
				Name:  s.Name,
				Type:  "query",
				Value: val,
				Raw:   &val,
			})
		} else if s.In == "cookie" {
			log.Request.Parameters = append(log.Request.Parameters, HttpParameter{
				Name:  s.Name,
				Type:  "cookie",
				Value: val,
				Raw:   &val,
			})
		} else {
			name := strings.ToLower(s.Name)
			for i, p := range log.Request.Parameters {
				if strings.ToLower(p.Name) == name && p.Type == s.In {
					// Golang enforces a canonical format, where each word's first letter is capitalized.
					// Update the name to be identical to the specification name
					p.Name = s.Name
					p.Value = val
					log.Request.Parameters[i] = p
					break
				}
			}
		}
	}

	return nil
}

type OAuth2SecurityScheme struct {
	Type        string                 `yaml:"type" json:"type"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Flows       map[string]*OAuth2Flow `yaml:"flows" json:"flows"`
}

type OAuth2Flow struct {
	AuthorizationUrl string `yaml:"authorizationUrl" json:"authorizationUrl"`
	TokenUrl         string `yaml:"tokenUrl" json:"tokenUrl"`
	RefreshUrl       string `yaml:"refreshUrl" json:"refreshUrl"`
	// Scopes map between scope name and a short description
	Scopes map[string]string `yaml:"scopes" json:"scopes"`
}

func (s *OAuth2SecurityScheme) Serve(req *http.Request) error {
	request := EventRequestFromContext(req.Context())
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return fmt.Errorf("missing authorization header")
	}
	request.Header["Authorization"] = auth

	if log, ok := LogEventFromContext(req.Context()); ok {
		for i, p := range log.Request.Parameters {
			if p.Name == "Authorization" {
				p.Value = auth
				log.Request.Parameters[i] = p
				break
			}
		}
	}

	return nil
}

type NotSupportedSecuritySchemeError struct {
	Scheme string
}

type NotSupportedSecurityScheme struct {
	Type string `yaml:"type" json:"type"`
}

func (s *NotSupportedSecurityScheme) Serve(req *http.Request) error {
	return &NotSupportedSecuritySchemeError{Scheme: s.Type}
}

func (e *NotSupportedSecuritySchemeError) Error() string {
	return fmt.Sprintf("security scheme %v not supported", e.Scheme)
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
		case "oauth2":
			v = &OAuth2SecurityScheme{}
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
		case "oauth2":
			v = &OAuth2SecurityScheme{}
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

func getSecuritySchemeType(schema SecurityScheme) string {
	switch v := schema.(type) {
	case *HttpSecurityScheme:
		return v.Type
	case *ApiKeySecurityScheme:
		return v.Type
	case *OAuth2SecurityScheme:
		return v.Type
	case *NotSupportedSecurityScheme:
		return v.Type
	default:
		return "unknown"
	}
}
