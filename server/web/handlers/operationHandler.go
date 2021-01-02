package handlers

import (
	"mokapi/providers/data"
	"mokapi/providers/pipeline"
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
	response := &Response{httpContext: context}

	if len(context.MokapiFile) >= 0 && operation.Pipeline != nil {
		err := pipeline.Run(context.MokapiFile, *operation.Pipeline,
			pipeline.WithGlobalVars(map[pipeline.Type]interface{}{
				"response":  response,
				"Operation": operation,
			}))
		if err != nil {
			ProcessError(err.Error(), context)
		}
	} else {
		provider := data.NewRandomDataProvider()
		d, err := provider.Provide(context.Schema)
		if err != nil {
			ProcessError(err.Error(), context)
		} else {
			err = response.Write(d)
			if err != nil {
				ProcessError(err.Error(), context)
			}
		}
	}
}

func ProcessError(message string, context *web.HttpContext) {
	log.WithFields(log.Fields{"url": context.Request.URL.String()}).Errorf(message)
	http.Error(context.Response, message, http.StatusInternalServerError)
}
