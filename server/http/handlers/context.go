package handlers

import (
	"fmt"
	"mokapi/providers/data"
	"mokapi/service"
	"net/http"
	"regexp"
	"strings"
)

type Context struct {
	Response     http.ResponseWriter
	Request      *http.Request
	ServiceUrl   string
	Parameters   map[string]string
	DataProvider data.Provider
	Responses    map[service.HttpStatus]*service.Response
}

type ContextParameter struct {
	Parameter *service.Parameter
	Value     string
}

func NewContext(serviceUrl string, response http.ResponseWriter, request *http.Request) *Context {
	return &Context{Response: response, Request: request, ServiceUrl: serviceUrl, Parameters: make(map[string]string)}
}

func NewContextParameter(p *service.Parameter, s string) *ContextParameter {
	return &ContextParameter{Parameter: p, Value: s}
}

func (c *Context) Update(e *service.Endpoint, provider data.Provider) error {
	operation := e.GetOperation(c.Request.Method)
	c.Responses = operation.Responses

	c.DataProvider = provider

	// resolving parameters from current context

	segments := strings.Split(c.Request.URL.Path, "/")

	pathParameterIndex := make(map[string]int)
	paramRegex := regexp.MustCompile(`\{(?P<name>.+)\}`)
	for i, part := range strings.Split(e.Path, "/") {
		match := paramRegex.FindStringSubmatch(part)
		if len(match) > 1 {
			paramName := match[1]
			pathParameterIndex[paramName] = i
		}
	}

	parameters := append(e.Parameters, operation.Parameters...)
	for _, p := range parameters {
		value := ""

		switch p.Type {
		case service.CookieParameter:
		case service.QueryParameter:
			value = c.Request.URL.Query().Get(p.Name)
		case service.HeaderParameter:
			value = c.Request.Header.Get(p.Name)
		case service.PathParameter:
			key := fmt.Sprintf("{%v}", p.Name)
			if i, ok := pathParameterIndex[key]; ok {
				value = segments[i]
			} else {
				return fmt.Errorf("Path parameter %v not found in request %v", key, c.Request.URL)
			}
		}

		c.Parameters[p.Name] = value
	}

	return nil
}
