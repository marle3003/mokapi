package handlers

import (
	"fmt"
	"mokapi/providers/data"
	"mokapi/providers/encoding"
	"mokapi/service"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type ResponseHandler struct {
	Parameters  []*service.Parameter
	ContentType service.ContentType
	schema      *service.Schema
	Provider    data.Provider
}

func NewContentHandler(parameters []*service.Parameter, contentType service.ContentType, schema *service.Schema, provider data.Provider) *ResponseHandler {
	return &ResponseHandler{Parameters: parameters, ContentType: contentType, schema: schema, Provider: provider}
}

func (handler *ResponseHandler) ServeHTTP(context *Context) {
	log.WithFields(log.Fields{"url": context.Request.URL.String(), "host": context.Request.Host, "method": context.Request.Method, "contentType": handler.ContentType}).Info("Serve http request")

	// set content type
	context.Response.Header().Add("Content-Type", handler.ContentType.String())

	// get content
	content, error := handler.Provider.Provide(nil, handler.schema)
	if error != nil {
		http.Error(context.Response, error.Error(), http.StatusInternalServerError)
		return
	}

	// write content
	data, error := handler.Encode(content)
	if error != nil {
		http.Error(context.Response, error.Error(), http.StatusInternalServerError)
		return
	}
	context.Response.Write(data)
}

func (handler *ResponseHandler) Encode(obj interface{}) ([]byte, error) {
	switch handler.ContentType {
	case "application/json", "application/json;odata=verbose":
		return encoding.MarshalJSON(obj, handler.schema)
	case "application/rss+xml":
		return encoding.MarshalXML(obj, handler.schema)
	default:
		return nil, fmt.Errorf("Unspupported encoding for content type %v", handler.ContentType)
	}
}
