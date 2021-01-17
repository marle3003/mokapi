package web

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"mokapi/models"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ParameterParser interface {
	parse() (interface{}, error)
}

type Parameter struct {
	Path   map[string]interface{}
	Query  map[string]interface{}
	Header map[string]interface{}
	Cookie map[string]interface{}
}

func newParamter() *Parameter {
	return &Parameter{
		Path:   make(map[string]interface{}),
		Query:  make(map[string]interface{}),
		Header: make(map[string]interface{}),
		Cookie: make(map[string]interface{}),
	}
}

type HttpContext struct {
	Response        http.ResponseWriter
	Request         *http.Request
	Parameters      *Parameter
	ResponseType    *models.Response
	ServicPath      string
	CurrentEndpoint *models.Endpoint
	ContentType     *models.ContentType
	Schema          *models.Schema
	MokapiFile      string
	body            string
	metric          *models.RequestMetric
}

func NewHttpContext(request *http.Request, response http.ResponseWriter, servicePath string) *HttpContext {
	return &HttpContext{Response: response,
		Request:    request,
		ServicPath: servicePath,
		Parameters: newParamter(),
	}
}

func (context *HttpContext) Body() string {
	if context.Request.ContentLength == 0 {
		return ""
	}
	if context.body != "" {
		return context.body
	}

	data, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		log.Errorf("Error while reading body: %v", err.Error())
		return ""
	}
	context.body = string(data)
	return context.body
}

func (context *HttpContext) SetCurrentEndpoint(endpoint *models.Endpoint) error {
	context.CurrentEndpoint = endpoint
	operation := endpoint.GetOperation(context.Request.Method)
	error := context.setFirstSuccessResponse(operation)
	if error != nil {
		return error
	}
	error = context.setContentType()
	if error != nil {
		return error
	}
	context.Schema = context.ResponseType.ContentTypes[context.ContentType.Key()].Schema

	segments := strings.Split(context.Request.URL.Path, "/")

	pathParameterIndex := make(map[string]int)
	paramRegex := regexp.MustCompile(`\{(?P<name>.+)\}`)
	for i, segment := range strings.Split(endpoint.Path, "/") {
		match := paramRegex.FindStringSubmatch(segment)
		if len(match) > 1 {
			paramName := match[1]
			pathParameterIndex[paramName] = i
		}
	}

	params := append(endpoint.Parameters, operation.Parameters...)
	sort.SliceStable(params, func(i, j int) bool {
		return params[i].Location < params[j].Location
	})
	for _, p := range params {
		var parser ParameterParser
		var store map[string]interface{}
		switch p.Location {
		//case models.CookieParameter:
		case models.PathParameter:
			if i, ok := pathParameterIndex[p.Name]; ok {
				parser = newPathParam(p, segments[i], context)
				store = context.Parameters.Path
			} else {
				return fmt.Errorf("path parameter %v not found in request %v", p.Name, context.Request.URL)
			}
		case models.QueryParameter:
			parser = newQueryParam(p, context)
			store = context.Parameters.Query
			//case models.HeaderParameter:
			//	value = context.Request.Header.Get(p.Name)
			//case models.PathParameter:

		}
		if parser != nil {
			v, err := parser.parse()
			if err != nil {
				return errors.Wrapf(err, "parse param '%v' from location %v", p.Name, p.Location)
			}
			store[p.Name] = v
		}
	}

	return nil
}

func (context *HttpContext) setFirstSuccessResponse(operation *models.Operation) error {
	successStatus := make([]models.HttpStatus, 0, 1)
	for httpStatus := range operation.Responses {
		if httpStatus >= 200 && httpStatus < 300 {
			successStatus = append(successStatus, httpStatus)
		}
	}

	if len(successStatus) == 0 {
		return fmt.Errorf("No success response in configuration found")
	}

	sort.SliceStable(successStatus, func(i, j int) bool { return i < j })

	context.ResponseType = operation.Responses[successStatus[0]]
	return nil
}

func (context *HttpContext) setContentType() error {
	accept := context.Request.Header.Get("accept")

	// search for a matching content type
	if accept != "" {
		for _, mimeType := range strings.Split(accept, ",") {
			contentType := models.ParseContentType(mimeType)
			if _, ok := context.ResponseType.ContentTypes[contentType.Key()]; ok {
				context.ContentType = contentType
				return nil
			}
		}
	}

	// no matching content found => returning first in list
	for contentType := range context.ResponseType.ContentTypes {
		// return first element
		context.ContentType = models.ParseContentType(contentType)
		return nil
	}

	return fmt.Errorf("No content type found")
}

func (ctx *HttpContext) parse(s string, schema *models.Schema) (interface{}, error) {
	switch schema.Type {
	case "string":
		return s, nil
	case "integer":
		return strconv.Atoi(s)
	case "number":
		return strconv.ParseFloat(s, 64)
	case "boolean":
		return strconv.ParseBool(s)
		//case "array":
		//case "object":
	}
	return nil, errors.Errorf("unable to parse '%v'", s)
}
