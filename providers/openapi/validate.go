package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"mokapi/media"
	"mokapi/schema/json/parser"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

const responseBodyLocation = "response.body"

// ValidationError describes one violation of the contract. Location names the checked part,
// e.g. "request.query.id" or "response.body". Schema is the JSON pointer of a schema error.
type ValidationError struct {
	Location string `json:"location"`
	Schema   string `json:"schema,omitempty"`
	Message  string `json:"message"`
}

// OperationMatch is the operation a request URL resolved to. It keeps the URL without the base
// path of the matched server, as the mock server hands it to the operation.
type OperationMatch struct {
	Operation *Operation
	url       *url.URL
}

// ResolveOperation finds the operation for method and requestUrl like a mock request would:
// literal paths beat parameterized ones and the base path of a server URL is stripped. It
// returns nil if no operation matches. Without a server context, the base paths of all servers
// are tried, the longest first, and the unstripped path last.
func (c *Config) ResolveOperation(method, requestUrl string) (*OperationMatch, error) {
	u, err := url.Parse(requestUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	for _, basePath := range c.basePathsOf(u.Path) {
		stripped := withoutBasePath(u, basePath)
		op, _ := findOperation(method, trimTrailingSlash(stripped.Path), c.Paths)
		if op != nil {
			return &OperationMatch{Operation: op, url: stripped}, nil
		}
	}
	return nil, nil
}

// basePathsOf returns the parts of requestPath that match a server base path, longest first,
// followed by "" for the unstripped path. Server variables are not supported, therefore a
// server URL with a variable never matches.
func (c *Config) basePathsOf(requestPath string) []string {
	var matches []string
	for _, server := range c.Servers {
		if strings.Contains(server.Url, "{") {
			continue
		}
		u, err := url.Parse(server.Url)
		if err != nil {
			continue
		}
		basePath := strings.TrimRight(u.Path, "/")
		if basePath != "" && len(requestPath) >= len(basePath) && strings.EqualFold(requestPath[:len(basePath)], basePath) {
			matches = append(matches, requestPath[:len(basePath)])
		}
	}
	slices.SortFunc(matches, func(a, b string) int { return len(b) - len(a) })
	return append(slices.Compact(matches), "")
}

func withoutBasePath(u *url.URL, basePath string) *url.URL {
	stripped := *u
	stripped.Path = stripBasePath(trimTrailingSlash(u.Path), basePath)
	stripped.RawPath = ""
	if u.RawPath != "" && len(u.RawPath) >= len(basePath) && strings.EqualFold(u.RawPath[:len(basePath)], basePath) {
		stripped.RawPath = stripBasePath(trimTrailingSlash(u.RawPath), u.RawPath[:len(basePath)])
	}
	return &stripped
}

// ValidateRequest checks the request parameters and, if the operation defines a request body,
// the body against the contract. In contrast to a mock request, all errors are collected.
// Security requirements are not checked and extra headers are allowed.
func (m *OperationMatch) ValidateRequest(headers map[string]string, body string) []ValidationError {
	r := &http.Request{
		Method:        m.Operation.Method,
		URL:           m.url,
		Header:        toHttpHeader(headers),
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
	}

	errs := m.validateParameters(r)
	if m.Operation.RequestBody != nil && m.Operation.RequestBody.Value != nil {
		if _, err := BodyFromRequest(r, m.Operation); err != nil {
			errs = append(errs, newValidationErrors("request.body", err)...)
		}
	}
	return errs
}

func (m *OperationMatch) validateParameters(r *http.Request) []ValidationError {
	var errs []ValidationError
	var pathParams []*Parameter
	for _, ref := range slices.Concat(m.Operation.Path.Parameters, m.Operation.Parameters) {
		if ref.Value == nil {
			continue
		}
		param := ref.Value
		if param.Type == ParameterPath {
			pathParams = append(pathParams, param)
			continue
		}
		if _, err := parseParameter(param, r); err != nil {
			errs = append(errs, newValidationErrors(fmt.Sprintf("request.%s.%s", param.Type, param.Name), err)...)
		}
	}
	return append(errs, m.validatePathParameters(pathParams, r)...)
}

func (m *OperationMatch) validatePathParameters(params []*Parameter, r *http.Request) []ValidationError {
	route := trimTrailingSlash(m.Operation.Path.Path)
	values, err := findPathValues(params, route, r)
	if err == nil {
		var errs []ValidationError
		for _, param := range params {
			if _, ok := values[param.Name]; !ok {
				notFound := fmt.Errorf("path parameter %s not found in route %s", param.Name, route)
				errs = append(errs, newValidationErrors("request.path."+param.Name, notFound)...)
			}
		}
		return errs
	}

	// findPathValues stops at the first error. Parsing each parameter on its own names the
	// failing ones, an error that no single parameter explains is reported for the first.
	var errs []ValidationError
	for _, param := range params {
		if _, err := findPathValues([]*Parameter{param}, route, r); err != nil {
			errs = append(errs, newValidationErrors("request.path."+param.Name, err)...)
		}
	}
	if len(errs) == 0 {
		location := "request.path"
		if len(params) > 0 {
			location += "." + params[0].Name
		}
		errs = newValidationErrors(location, err)
	}
	return errs
}

func toHttpHeader(headers map[string]string) http.Header {
	h := http.Header{}
	for name, value := range headers {
		h.Set(name, value)
	}
	return h
}

// newValidationErrors turns a schema error into one entry per violation, using the format of
// the endpoint /api/schema/validate. Any other error becomes a single entry with its message.
func newValidationErrors(location string, err error) []ValidationError {
	var schemaError *parser.Error
	if !errors.As(err, &schemaError) {
		return []ValidationError{{Location: location, Message: err.Error()}}
	}

	var violations []any
	if json.Unmarshal([]byte(parser.Marshal(schemaError)), &violations) != nil {
		return []ValidationError{{Location: location, Message: err.Error()}}
	}
	return flattenViolations(location, violations)
}

func flattenViolations(location string, violations []any) []ValidationError {
	var errs []ValidationError
	for _, violation := range violations {
		switch v := violation.(type) {
		case string:
			errs = append(errs, ValidationError{Location: location, Message: v})
		case map[string]any:
			schema, _ := v["schema"].(string)
			message, _ := v["message"].(string)
			errs = append(errs, ValidationError{Location: location, Schema: schema, Message: message})
			if nested, ok := v["errors"].([]any); ok {
				errs = append(errs, flattenViolations(location, nested)...)
			}
		}
	}
	return errs
}

// ValidateResponse checks status, headers and, for operations other than HEAD, the body against
// the response the contract defines for statusCode. Only JSON bodies are validated against a
// schema.
func (m *OperationMatch) ValidateResponse(statusCode int, headers map[string]string, body string) []ValidationError {
	res := m.Operation.getResponse(statusCode)
	if res == nil {
		return []ValidationError{{Location: "response.status", Message: fmt.Sprintf("no response defined for status code %d", statusCode)}}
	}

	h := toHttpHeader(headers)
	errs := validateResponseHeaders(res.Headers, h)
	if m.Operation.Method != http.MethodHead {
		errs = append(errs, validateResponseBody(res, statusCode, h.Get("Content-Type"), body)...)
	}
	return errs
}

func validateResponseHeaders(definitions Headers, h http.Header) []ValidationError {
	var errs []ValidationError
	for _, name := range slices.Sorted(maps.Keys(definitions)) {
		ref := definitions[name]
		// a defined Content-Type header is ignored as required by OpenAPI
		if ref == nil || ref.Value == nil || strings.EqualFold(name, "Content-Type") {
			continue
		}
		param := ref.Value.Parameter
		param.Name = name
		// header values arrive as strings and are converted like request headers
		if _, err := parseHeader(&param, h); err != nil {
			errs = append(errs, newValidationErrors("response.header."+name, err)...)
		}
	}
	return errs
}

func validateResponseBody(res *Response, statusCode int, contentTypeHeader, body string) []ValidationError {
	data := []byte(body)

	contentType := media.ParseContentType(contentTypeHeader)
	if contentType.IsEmpty() {
		return validateResponseBodyDetectContentType(res, data)
	}

	mt := res.GetContent(contentType)
	if mt == nil {
		return []ValidationError{{Location: responseBodyLocation, Message: fmt.Sprintf("content type '%s' is not defined for status code %d", contentType, statusCode)}}
	}
	if !isJson(contentType) {
		return nil
	}
	if _, err := mt.ParseData(data, contentType); err != nil {
		return newValidationErrors(responseBodyLocation, err)
	}
	return nil
}

func validateResponseBodyDetectContentType(res *Response, data []byte) []ValidationError {
	if len(data) == 0 {
		if len(res.Content) == 0 {
			return nil
		}
		return []ValidationError{{Location: responseBodyLocation, Message: "response body is required"}}
	}

	_, err := parseBodyDetectContentType(data, res.Content)
	if err == nil {
		return nil
	}
	// a missing Content-Type must not hide the actual error
	if mt := singleJsonMediaType(res.Content); mt != nil {
		_, err = mt.ParseData(data, mt.ContentType)
	}
	return newValidationErrors(responseBodyLocation, err)
}

func singleJsonMediaType(content Content) *MediaType {
	var found *MediaType
	for _, mt := range content {
		if isJson(mt.ContentType) {
			if found != nil {
				return nil
			}
			found = mt
		}
	}
	return found
}

func isJson(contentType media.ContentType) bool {
	return contentType.Subtype == "json" || strings.HasSuffix(contentType.Subtype, "+json")
}
