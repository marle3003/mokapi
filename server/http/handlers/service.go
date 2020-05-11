package handlers

import (
	"fmt"
	"mokapi/providers/data"
	"mokapi/service"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ServiceHandler struct {
	service      *service.Service
	dataProvider data.Provider
}

func NewServiceHandler(s *service.Service, dataProvider data.Provider) *ServiceHandler {
	return &ServiceHandler{service: s, dataProvider: dataProvider}
}

func (s *ServiceHandler) ServeHTTP(context *Context) {
	s.resolveEndpoint(context)
}

func (s *ServiceHandler) Close() {
	s.dataProvider.Close()
}

func (h *ServiceHandler) resolveEndpoint(context *Context) {

	for _, e := range h.service.Endpoint {
		o := e.GetOperation(context.Request.Method)
		if o == nil {
			continue
		}

		p := append(e.Parameters, o.Parameters...)

		if isMatchingPath(e.Path, p, context) {
			context.Update(e, h.dataProvider)

			handler := NewResponseHandler()
			handler.ServeHTTP(context)

			return
		}
	}

	context.Response.WriteHeader(404)
	fmt.Fprintf(context.Response, "No endpoint found %s %s", context.Request.Method, context.Request.URL.String())
	log.Infof("No endpoint found %s %v", context.Request.Method, context.Request.URL)
}

func isMatchingPath(path string, params []*service.Parameter, c *Context) bool {
	pathToValidate := c.Request.URL.Path
	var parts []string
	if c.ServiceUrl != "/" {
		pathToValidate = c.Request.URL.Path[len(c.ServiceUrl):]
	}
	parts = strings.Split(pathToValidate, "/")

	parameters := make(map[string]*service.Parameter)
	for _, v := range params {
		parameters[v.Name] = v
	}

	i := 0
	paramRegex := regexp.MustCompile(`\{(?P<name>.+)\}`)
	for _, part := range strings.Split(path, "/") {
		match := paramRegex.FindStringSubmatch(part)
		if len(match) > 1 {
			paramName := match[1]
			if p, ok := parameters[paramName]; ok && p.Type == service.PathParameter {
				if !isValidParameterValue(p, parts[i]) {
					log.Errorf("Invalid parameter value %v found in path %v", parts[i], path)
					return false
				}
			} else {
				log.Errorf("No parameter definition %v found in path %v", paramName, path)
				return false
			}
		} else {
			if part != parts[i] {
				return false
			}
		}
		i++
	}

	return true
}

func isValidParameterValue(p *service.Parameter, s string) bool {
	if p.Schema == nil {
		return true
	}

	switch strings.ToLower(p.Schema.Type) {
	case "string":
		return true
	case "number":
		if _, error := strconv.ParseFloat(s, 64); error == nil {
			return true
		}
	case "integer":
		if _, error := strconv.Atoi(s); error == nil {
			return true
		}
	case "boolean":
		if _, error := strconv.ParseBool(s); error == nil {
			return true
		}
	case "array":
		log.Error("Paramter type array not supported")
	case "object":
		log.Error("Paramter type array not supported")
	default:
		log.Errorf("Paramter type %v not supported", p.Schema.Type)
	}

	return false
}
