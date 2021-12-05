package web

import (
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic/openapi"
	"mokapi/models"
	"mokapi/models/media"
	"net/http"
	"sort"
	"strings"
)

type ParameterParser interface {
	parse() (interface{}, error)
}

type HttpContext struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Parameters     RequestParameters
	ServicePath    string
	ServiceName    string
	EndpointPath   string
	Operation      *openapi.Operation
	ContentType    *media.ContentType
	metric         *models.RequestMetric
	statusCode     openapi.HttpStatus
	eventHandler   eventHandler
	Response       *openapi.MediaType
	Headers        map[string]*openapi.HeaderRef
}

func NewHttpContext(request *http.Request, response http.ResponseWriter, eh eventHandler) *HttpContext {
	metric := models.NewRequestMetric(request.Method, getUrl(request))
	return &HttpContext{ResponseWriter: response,
		Request:      request,
		eventHandler: eh,
		metric:       metric,
	}
}

func (context *HttpContext) getFirstSuccessResponse(operation *openapi.Operation) (openapi.HttpStatus, *openapi.ResponseRef, error) {
	successStatus := make([]openapi.HttpStatus, 0, 1)
	for httpStatus := range operation.Responses {
		if httpStatus.IsSuccess() {
			successStatus = append(successStatus, httpStatus)
		}
	}

	if len(successStatus) == 0 {
		return 0, nil, fmt.Errorf("no success response (HTTP 2xx) in configuration")
	}

	sort.SliceStable(successStatus, func(i, j int) bool { return i < j })

	return successStatus[0], operation.Responses[successStatus[0]], nil
}

func (context *HttpContext) setResponse() error {
	status, response, err := context.getFirstSuccessResponse(context.Operation)
	if err != nil {
		return err
	}

	context.statusCode = status
	context.Headers = response.Value.Headers

	accept := context.Request.Header.Get("accept")

	// search for a matching content type
	if accept != "" {
		for _, mimeType := range strings.Split(accept, ",") {
			contentType := media.ParseContentType(mimeType)
			if mt, ok := response.Value.GetContent(contentType); ok {
				context.ContentType = contentType
				context.Response = mt
				return nil
			}
		}
		return newHttpErrorf(415, "none of requests content type(s) are supported: %v", accept)
	}

	return nil
}

func (context *HttpContext) updateMetricWithError(statusCode int, body string) {
	context.updateMetric(statusCode, "text/plain; charset=utf-8", body)
	context.metric.IsError = true
}

func (context *HttpContext) updateMetric(statusCode int, contentType, body string) {
	context.metric.HttpStatus = statusCode
	context.metric.ContentType = contentType
	context.metric.ResponseBody = body

	// headers
	for k, v := range context.Request.Header {
		p := models.RequestParamter{
			Name: k,
			Raw:  fmt.Sprintf("%v", v),
			Type: "header",
		}
		if v, ok := context.Parameters[openapi.HeaderParameter][k]; ok {
			data, _ := json.Marshal(v.Value)
			p.Value = string(data)
		}
		context.metric.Parameters = append(context.metric.Parameters, p)
	}

	// path
	for k, v := range context.Parameters[openapi.PathParameter] {
		data, _ := json.Marshal(v.Value)
		p := models.RequestParamter{
			Name:  k,
			Value: string(data),
			Raw:   v.Raw,
			Type:  "path",
		}
		context.metric.Parameters = append(context.metric.Parameters, p)
	}

	// cookies
	for _, cookie := range context.Request.Cookies() {
		p := models.RequestParamter{
			Name: cookie.Name,
			Raw:  cookie.Raw,
			Type: "cookie",
		}
		if v, ok := context.Parameters[openapi.CookieParameter][cookie.Name]; ok {
			data, _ := json.Marshal(v.Value)
			p.Value = string(data)
		}
		context.metric.Parameters = append(context.metric.Parameters, p)
	}

	// query
	for k, v := range context.Parameters[openapi.QueryParameter] {
		data, _ := json.Marshal(v.Value)
		p := models.RequestParamter{
			Name:  k,
			Value: string(data),
			Raw:   v.Raw,
			Type:  "query",
		}
		context.metric.Parameters = append(context.metric.Parameters, p)
	}
}

func getUrl(r *http.Request) string {
	if r.URL.IsAbs() {
		return r.URL.String()
	}

	var scheme string
	if r.TLS == nil {
		scheme = "http"
	} else {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.String())
}
