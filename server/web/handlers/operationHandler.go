package handlers

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"mokapi/models"
	"mokapi/providers/encoding"
	"mokapi/providers/pipeline"
	"mokapi/providers/pipeline/lang/types"
	"mokapi/server/web"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type OperationHandler struct {
}

func NewOperationHandler() *OperationHandler {
	return &OperationHandler{}
}

func (handler *OperationHandler) ProcessRequest(context *web.HttpContext) {
	log.WithFields(log.Fields{
		"url":    context.Request.URL.String(),
		"host":   context.Request.Host,
		"method": context.Request.Method,
	}).Info("Processing http request")

	operation := context.CurrentEndpoint.GetOperation(context.Request.Method)

	var bodyParam interface{} = nil
	if operation.RequestBody != nil {
		contentType := models.ParseContentType(context.Request.Header.Get("content-type"))
		body, err := readBody(context)
		if err != nil {
			respond(err.Error(), http.StatusInternalServerError, context)
			return
		}
		if operation.RequestBody.Required && len(body) == 0 {
			respond("request body expected", http.StatusBadRequest, context)
			return
		}
		if mediaType, ok := operation.RequestBody.ContentTypes[contentType.String()]; ok {
			bodyParam, err = parseBody(body, contentType, mediaType.Schema)
			if err != nil {
				respond(err.Error(), http.StatusBadRequest, context)
				return
			}
		} else {
			respond("content type of request body is not definied", http.StatusInternalServerError, context)
			return
		}
	}

	pipelineName := operation.Endpoint.Pipeline
	if len(operation.Pipeline) > 0 {
		pipelineName = operation.Pipeline
	}

	err := pipeline.Run(
		context.MokapiFile,
		pipelineName,
		pipeline.WithGlobalVars(map[types.Type]interface{}{
			"response":  &Response{httpContext: context},
			"Operation": operation,
		}),
		pipeline.WithParams(context.Parameters),
		pipeline.WithParams(map[string]interface{}{
			"method": context.Request.Method,
			"body":   bodyParam,
		}),
	)
	if err != nil {
		respond(err.Error(), http.StatusInternalServerError, context)
	}
}

func readBody(ctx *web.HttpContext) (string, error) {
	if ctx.Request.ContentLength == 0 {
		return "", nil
	}

	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func respond(message string, status int, context *web.HttpContext) {
	log.WithFields(log.Fields{"url": context.Request.URL.String()}).Errorf(message)
	http.Error(context.Response, message, status)
}

func parseBody(s string, contentType *models.ContentType, schema *models.Schema) (interface{}, error) {
	switch contentType.Subtype {
	case "xml":
		return encoding.UnmarshalXml(s, schema)
	}

	return nil, errors.Errorf("unsupported content type '%v' from body", contentType)
}
