package api

import (
	"encoding/json"
	"fmt"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"net/http"
	"slices"
	"strings"
)

type httpValidateRequest struct {
	Api     string `json:"api"`
	Request *struct {
		Method  string            `json:"method"`
		Url     string            `json:"url"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	} `json:"request"`
	Response *struct {
		StatusCode int               `json:"statusCode"`
		Headers    map[string]string `json:"headers"`
		Body       string            `json:"body"`
	} `json:"response"`
}

type httpValidateResult struct {
	Valid     bool                      `json:"valid"`
	Operation *httpValidateOperation    `json:"operation,omitempty"`
	Errors    []openapi.ValidationError `json:"errors"`
}

type httpValidateOperation struct {
	Api    string `json:"api"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

type httpOperationMatch struct {
	api   string
	match *openapi.OperationMatch
}

// validateHttpExchange checks a request/response pair against the OpenAPI contracts without
// serving it: no event, no metric and no script.
func (h *handler) validateHttpExchange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, fmt.Errorf("method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
		return
	}

	payload, err := parseHttpValidateRequest(r)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	apis, err := h.selectHttpApis(payload.Api)
	if err != nil {
		writeError(w, err, http.StatusNotFound)
		return
	}

	matches, err := findOperations(apis, payload.Request.Method, payload.Request.Url)
	if err != nil {
		writeError(w, fmt.Errorf("invalid request.url: %w", err), http.StatusBadRequest)
		return
	}

	result := httpValidateResult{Errors: []openapi.ValidationError{}}
	switch len(matches) {
	case 0:
		result.Errors = append(result.Errors, openapi.ValidationError{
			Location: "operation",
			Message:  fmt.Sprintf("no operation found for %s %s", payload.Request.Method, payload.Request.Url),
		})
	case 1:
		m := matches[0]
		result.Operation = &httpValidateOperation{Api: m.api, Method: m.match.Operation.Method, Path: m.match.Operation.Path.Path}
		result.Errors = append(result.Errors, m.match.ValidateRequest(payload.Request.Headers, payload.Request.Body)...)
		result.Errors = append(result.Errors, m.match.ValidateResponse(payload.Response.StatusCode, payload.Response.Headers, payload.Response.Body)...)
	default:
		result.Errors = append(result.Errors, openapi.ValidationError{
			Location: "operation",
			Message: fmt.Sprintf("ambiguous operation: %s %s matches the APIs %s. Set 'api' to select one.",
				payload.Request.Method, payload.Request.Url, quotedNames(matches)),
		})
	}
	result.Valid = len(result.Errors) == 0

	write(w, result)
}

func parseHttpValidateRequest(r *http.Request) (*httpValidateRequest, error) {
	var payload httpValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	switch {
	case payload.Request == nil:
		return nil, fmt.Errorf("request is required")
	case payload.Response == nil:
		return nil, fmt.Errorf("response is required")
	case payload.Request.Method == "":
		return nil, fmt.Errorf("request.method is required")
	case payload.Request.Url == "":
		return nil, fmt.Errorf("request.url is required")
	case payload.Response.StatusCode == 0:
		return nil, fmt.Errorf("response.statusCode is required")
	}
	return &payload, nil
}

// selectHttpApis returns the API named by name or, if name is empty, all HTTP APIs sorted by name.
func (h *handler) selectHttpApis(name string) ([]*runtime.HttpInfo, error) {
	if name == "" {
		apis := h.app.Http.List()
		slices.SortFunc(apis, func(a, b *runtime.HttpInfo) int { return strings.Compare(a.Info.Name, b.Info.Name) })
		return apis, nil
	}

	api := h.app.Http.Get(name)
	if api == nil {
		return nil, fmt.Errorf("api '%s' not found", name)
	}
	return []*runtime.HttpInfo{api}, nil
}

func findOperations(apis []*runtime.HttpInfo, method, url string) ([]httpOperationMatch, error) {
	var matches []httpOperationMatch
	for _, api := range apis {
		m, err := api.ResolveOperation(method, url)
		if err != nil {
			return nil, err
		}
		if m != nil {
			matches = append(matches, httpOperationMatch{api: api.Info.Name, match: m})
		}
	}
	return matches, nil
}

func quotedNames(matches []httpOperationMatch) string {
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, fmt.Sprintf("'%s'", m.api))
	}
	return strings.Join(names, ", ")
}
