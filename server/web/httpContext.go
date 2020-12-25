package web

import (
	"fmt"
	"io/ioutil"
	"mokapi/models"
	"mokapi/providers/data"
	"mokapi/providers/parser"
	"net/http"
	"regexp"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/xmlpath.v2"
)

type HttpContext struct {
	Response http.ResponseWriter
	Request  *http.Request
	// Name-value pairs are added to the collection in the following order:
	// 1. PathParameter
	// 2. QueryParameter
	// 3. HeaderParameter
	// 4. CookieParameter
	Parameters      map[string]string
	DataProvider    data.Provider
	ResponseType    *models.Response
	ServicPath      string
	CurrentEndpoint *models.Endpoint
	ContentType     *models.ContentType
	Schema          *data.Schema

	body string
}

func NewHttpContext(request *http.Request, response http.ResponseWriter, servicePath string) *HttpContext {
	return &HttpContext{Response: response, Request: request, ServicPath: servicePath, Parameters: make(map[string]string)}
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

	parameters := append(endpoint.Parameters, operation.Parameters...)
	sort.SliceStable(parameters, func(i, j int) bool {
		return parameters[i].Type < parameters[j].Type
	})
	for _, parameter := range parameters {
		value := ""

		switch parameter.Type {
		case models.CookieParameter:
		case models.QueryParameter:
			value = context.Request.URL.Query().Get(parameter.Name)
		case models.HeaderParameter:
			value = context.Request.Header.Get(parameter.Name)
		case models.PathParameter:
			if i, ok := pathParameterIndex[parameter.Name]; ok {
				value = segments[i]
			} else {
				return fmt.Errorf("Path parameter %v not found in request %v", parameter.Name, context.Request.URL)
			}
		}

		context.Parameters[parameter.Name] = value
	}

	return nil
}

func (context *HttpContext) GetMiddleware() []interface{} {
	operation := context.CurrentEndpoint.GetOperation(context.Request.Method)
	return operation.Middleware
}

func (context *HttpContext) GetResourceName() string {
	operation := context.CurrentEndpoint.GetOperation(context.Request.Method)
	for _, resource := range operation.Resources {
		if resource.If == nil {
			return resource.Name
		}

		if !resource.If.IsValid() {
			log.Infof("Skipping invalid expression: %v", resource.If.Raw)
			continue
		}

		match, error := resource.If.Expr.IsTrue(func(factor string, tag parser.ExpressionTag) string {
			switch tag {
			case parser.Body:
				s, error := context.SelectFromBody(factor)
				if error != nil {
					log.Error(error.Error())
					return ""
				}
				return s
			case parser.Parameter:
				return context.Parameters[factor]
			default:
				return factor
			}
		})
		if error != nil {
			log.Error(error.Error())
		} else if match {
			return resource.Name
		}
	}
	return ""
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

func (context *HttpContext) SelectFromBody(selector string) (string, error) {
	contentType := models.ParseContentType(context.Request.Header.Get("content-type"))

	switch contentType.Subtype {
	case "xml":
		path, error := xmlpath.Compile(selector)
		if error != nil {
			return "", fmt.Errorf("Expecting xpath as selector with content type %v", contentType)
		}
		reader := strings.NewReader(context.Body())
		node, error := xmlpath.Parse(reader)
		if error != nil {
			return "", fmt.Errorf("Error in xml parsing request body: %v", error.Error())
		}
		if value, ok := path.String(node); ok {
			return value, nil
		}
	default:
		return "", fmt.Errorf("Selection of Content type '%v' of request body is not supported", contentType)
	}

	return "", nil
}
