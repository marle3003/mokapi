package handlers

import (
	"fmt"
	"mokapi/models"
	"mokapi/providers/data"
	"mokapi/providers/encoding"
	"mokapi/server/web"
	"mokapi/server/web/middlewares"
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

	resource := context.GetResourceName()
	data, error := context.DataProvider.Provide(resource, context.Schema)

	if error != nil {
		// TODO select correct response from config
		ProcessError(error.Error(), context)
		return
	} else if data == nil {
		log.Infof("No data found responding with 404")
		context.Response.WriteHeader(404)
		return
	}

	data = runMiddlewares(data, context)

	if data == nil {
		log.Infof("No data found responding with 404")
		context.Response.WriteHeader(404)
		return
	}

	contentType := context.ContentType

	//encode
	bytes, ok := data.([]byte)
	if ok {
		if context.ContentType.Subtype == "*" {
			// detect content type by data
			contentType = models.ParseContentType(http.DetectContentType(bytes))
		}
	} else {
		bytes, error = encodeData(data, contentType, context.Schema)
		if error != nil {
			ProcessError(error.Error(), context)
			return
		}
	}

	// set content type
	context.Response.Header().Add("Content-Type", contentType.String())

	// write content
	context.Response.Header().Add("Content-Length", fmt.Sprint(len(bytes)))
	context.Response.Write(bytes)
}

func ProcessError(message string, context *web.HttpContext) {
	log.WithFields(log.Fields{"url": context.Request.URL.String()}).Errorf(message)
	http.Error(context.Response, message, http.StatusInternalServerError)
}

func encodeData(data interface{}, contentType *models.ContentType, schema *data.Schema) ([]byte, error) {
	switch contentType.Subtype {
	case "json":
		return encoding.MarshalJSON(data, schema)
	case "xml", "rss+xml":
		return encoding.MarshalXML(data, schema)
	default:
		if s, ok := data.(string); ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("Unspupported encoding for content type %v", contentType)
	}
}

func runMiddlewares(data interface{}, context *web.HttpContext) interface{} {
	middleware := middlewares.Create(context.GetMiddleware())
	request := middlewares.NewRequest(data)
	middleware.ServeData(request, context)
	return request.Data
}
