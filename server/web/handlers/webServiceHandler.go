package handlers

import (
	"fmt"
	"mokapi/models"
	"mokapi/server/web"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type WebServiceHandler struct {
	WebService *models.WebService
}

func NewWebServiceHandler(service *models.WebService) *WebServiceHandler {
	return &WebServiceHandler{WebService: service}
}

func (handler *WebServiceHandler) ServeHTTP(context *web.HttpContext) {
	endpoint := handler.resolveEndpoint(context)
	if endpoint == nil {
		message := fmt.Sprintf("No endpoint found in service %v. Request %v %v",
			handler.WebService.Name,
			context.Request.Method,
			context.Request.URL.String())
		http.Error(context.Response, message, http.StatusNotFound)
		log.Infof(message)
		return
	}

	operation := endpoint.GetOperation(context.Request.Method)
	if operation == nil {
		message := fmt.Sprintf("No operation found in endpoint %v in service %v. Request %v %v",
			endpoint.Path,
			handler.WebService.Name,
			context.Request.Method,
			context.Request.URL.String())
		http.Error(context.Response, message, http.StatusMethodNotAllowed)
		log.Infof(message)
		return
	}

	context.MokapiFile = handler.WebService.MokapiFile
	context.SetCurrentEndpoint(endpoint)

	operationHandler := NewOperationHandler()
	operationHandler.ProcessRequest(context)
}

func (handler *WebServiceHandler) resolveEndpoint(context *web.HttpContext) *models.Endpoint {
endpointLoop:
	for _, endpoint := range handler.WebService.Endpoint {
		operation := endpoint.GetOperation(context.Request.Method)
		if operation == nil {
			continue
		}

		requestSegments := getRequestSegments(context)
		operationSegments := strings.Split(endpoint.Path, "/")

		if len(requestSegments) != len(operationSegments) {
			continue
		}

		parameters := toMap(append(endpoint.Parameters, operation.Parameters...))
		parameterRegex := regexp.MustCompile(`\{(?P<name>.+)\}`) // parameter format "/{param}/"

		for index, segment := range operationSegments {
			match := parameterRegex.FindStringSubmatch(segment)
			if len(match) > 1 {
				parameterName := match[1]
				if parameter, ok := parameters[parameterName]; ok && parameter.Location == models.PathParameter {
					parameterValue := requestSegments[index]
					if err := validateParameterValue(parameterValue, parameter); err != nil {
						log.Debug(err)
						continue endpointLoop
					}
				} else {
					log.Debugf("No parameter definition %v found in path %v", parameterName, endpoint.Path)
					continue endpointLoop
				}
			} else {
				if segment != requestSegments[index] {
					continue endpointLoop
				}
			}
		}
		return endpoint
	}

	return nil
}

func getRequestSegments(context *web.HttpContext) []string {
	path := context.Request.URL.Path
	if context.ServicPath != "/" {
		// remove service path
		path = path[len(context.ServicPath):]
	}
	return strings.Split(path, "/")
}

func toMap(parameters []*models.Parameter) map[string]*models.Parameter {
	result := make(map[string]*models.Parameter)
	for _, v := range parameters {
		result[v.Name] = v
	}
	return result
}

func validateParameterValue(value string, parameter *models.Parameter) error {
	if parameter.Schema == nil {
		return nil
	}

	switch strings.ToLower(parameter.Schema.Type) {
	case "string":
		return nil
	case "number":
		if _, error := strconv.ParseFloat(value, 64); error == nil {
			return nil
		}
	case "integer":
		if _, error := strconv.Atoi(value); error == nil {
			return nil
		}
	case "boolean":
		if _, error := strconv.ParseBool(value); error == nil {
			return nil
		}
	case "array":
		return fmt.Errorf("Paramter type array is not supported in url.")
	case "object":
		return fmt.Errorf("Paramter type array is not supported in url.")
	default:
		return fmt.Errorf("Paramter type %v not supported", parameter.Schema.Type)
	}

	return fmt.Errorf("Invalid type of value %v for parameter %v. Expected type is %v", value, parameter.Name, parameter.Schema.Type)
}
