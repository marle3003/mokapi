package web

import (
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic/openapi"
	"mokapi/engine"
	"mokapi/models"
	"mokapi/models/media"
	"net/http"
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
	emitter        engine.EventEmitter
	Response       *openapi.MediaType
	Headers        map[string]*openapi.HeaderRef
}

func NewHttpContext(request *http.Request, response http.ResponseWriter) *HttpContext {
	metric := models.NewRequestMetric(request.Method, getUrl(request))
	return &HttpContext{ResponseWriter: response,
		Request: request,
		metric:  metric,
	}
}

func (context *HttpContext) getFirstSuccessResponse() (openapi.HttpStatus, *openapi.ResponseRef, error) {
	var successStatus openapi.HttpStatus
	for it := context.Operation.Responses.Iter(); it.Next(); {
		status := it.Key().(openapi.HttpStatus)
		if status.IsSuccess() {
			successStatus = status
			break
		}
	}

	if successStatus == 0 {
		return 0, nil, fmt.Errorf("no success response (HTTP 2xx) in configuration")
	}

	v := context.Operation.Responses.GetResponse(successStatus)

	return successStatus, v, nil
}

func (context *HttpContext) setResponse() error {
	status, response, err := context.getFirstSuccessResponse()
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
			if mt := response.Value.GetContent(contentType); mt != nil {
				context.ContentType = contentType
				context.Response = mt
				return nil
			}
		}
		return newHttpErrorf(415, "none of requests content type(s) are supported: %v", accept)
	}

	for i, c := range response.Value.Content {
		// return first element
		context.ContentType = media.ParseContentType(i)
		context.Response = c
		return nil
	}

	return fmt.Errorf("no content type found for accept header %q", accept)

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

func (context *HttpContext) Event(request *Request, response *Response) {
	if context.emitter == nil {
		return
	}
	context.emitter.Emit("http", request, response)
}
