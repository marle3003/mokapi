package web

import (
	"fmt"
	"io/ioutil"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
	"mokapi/providers/encoding"
	"mokapi/providers/pipeline"
	"mokapi/providers/pipeline/lang/types"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type OperationHandler struct {
}

func NewOperationHandler() *OperationHandler {
	return &OperationHandler{}
}

func (handler *OperationHandler) ProcessRequest(context *HttpContext) {
	log.WithFields(log.Fields{
		"url":    context.Request.URL.String(),
		"host":   context.Request.Host,
		"method": context.Request.Method,
	}).Info("Processing http request")

	operation := context.Operation

	var bodyParam interface{} = nil
	if operation.RequestBody != nil {
		contentType := media.ParseContentType(context.Request.Header.Get("content-type"))
		body, err := readBody(context)
		if err != nil {
			respond(err.Error(), http.StatusInternalServerError, context)
			return
		}
		if operation.RequestBody.Value.Required && len(body) == 0 {
			respond("request body expected", http.StatusBadRequest, context)
			return
		}
		if mediaType, ok := operation.RequestBody.Value.Content[contentType.Key()]; ok {
			bodyParam, err = parseBody(body, contentType, mediaType.Schema)
			if err != nil {
				respond(err.Error(), http.StatusBadRequest, context)
				return
			}
		} else {
			respond(
				fmt.Sprintf("content type '%v' of request body is not defined. Check your service configuration", contentType.String()),
				http.StatusInternalServerError,
				context)
			return
		}
	}

	pipelineName := operation.Endpoint.Pipeline
	if len(operation.Pipeline) > 0 {
		pipelineName = operation.Pipeline
	}

	if context.Mokapi != nil {
		err := pipeline.RunConfig(
			pipelineName,
			context.Mokapi,
			pipeline.WithGlobalVars(map[types.Type]interface{}{
				"response": &Response{httpContext: context, err: func(err error, status int) {
					respond(err.Error(), http.StatusInternalServerError, context)
				}},
				"Operation": operation,
			}),
			pipeline.WithParams(map[string]interface{}{
				"method": context.Request.Method,
				"body":   bodyParam,
				"query":  context.Parameters[openapi.QueryParameter],
				"path":   context.Parameters[openapi.PathParameter],
				"header": context.Parameters[openapi.HeaderParameter],
				"cookie": context.Parameters[openapi.CookieParameter],
			}),
		)
		if err != nil {
			respond(err.Error(), http.StatusInternalServerError, context)
		}
	}
}

func readBody(ctx *HttpContext) (string, error) {
	if ctx.Request.ContentLength == 0 {
		return "", nil
	}

	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func respond(message string, status int, ctx *HttpContext) {
	ctx.metric.Error = message
	ctx.metric.HttpStatus = status
	log.WithFields(log.Fields{"url": ctx.Request.URL.String()}).Errorf(message)
	http.Error(ctx.Response, message, status)
}

func parseBody(s string, contentType *media.ContentType, schema *openapi.SchemaRef) (interface{}, error) {
	switch contentType.Subtype {
	case "xml":
		return encoding.UnmarshalXml(s, schema)
	default:
		log.Debugf("unsupported content type '%v' from body", contentType)
		return s, nil
	}
}
