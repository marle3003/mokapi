package handlers

import (
	"fmt"
	"mokapi/models"
	"mokapi/providers/encoding"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ResponseHandler struct {
}

func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

func (handler *ResponseHandler) ServeHTTP(context *Context) {
	log.WithFields(log.Fields{"url": context.Request.URL.String(), "host": context.Request.Host, "method": context.Request.Method}).Info("Serve http request")

	// first we select the possible response. With that schema we selects the data.
	// In OpenApi each http status code and content type can defined its own schema
	// Depending on data selection we change http status and schema of that http status code definition

	response, ok := context.Responses[models.Ok]
	if !ok {
		log.WithFields(log.Fields{"url": context.Request.URL.String()}).Errorf("No 200 response in configuration found")
		http.Error(context.Response, "No 200 response in configuration found", http.StatusInternalServerError)
		return
	}

	contentType, error := getContentType(response, context)
	if error != nil {
		log.WithFields(log.Fields{"url": context.Request.URL.String()}).Errorf(error.Error())
		http.Error(context.Response, error.Error(), http.StatusInternalServerError)
		return
	}

	schema := response.ContentTypes[contentType].Schema

	// get content
	data, error := context.DataProvider.Provide(context.Parameters, schema)
	if error != nil {
		// TODO select correct response from config
		http.Error(context.Response, error.Error(), http.StatusInternalServerError)
		return
	} else if data == nil {
		context.Response.WriteHeader(404)
		return
	}

	// set content type
	context.Response.Header().Add("Content-Type", contentType.String())

	// write content
	bytes, error := handler.Encode(data, contentType, schema)
	if error != nil {
		http.Error(context.Response, error.Error(), http.StatusInternalServerError)
		return
	}
	context.Response.Write(bytes)
}

func (handler *ResponseHandler) Encode(data interface{}, contentType models.ContentType, schema *models.Schema) ([]byte, error) {
	switch contentType {
	case "application/json", "application/json;odata=verbose":
		return encoding.MarshalJSON(data, schema)
	case "application/rss+xml":
		return encoding.MarshalXML(data, schema)
	default:
		if s, ok := data.(string); ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("Unspupported encoding for content type %v", contentType)
	}
}

func getContentType(r *models.Response, c *Context) (models.ContentType, error) {
	accept := c.Request.Header.Get("accept")

	// search for a matching content type
	if accept != "" {
		for _, s := range strings.Split(accept, ",") {
			contentType, error := models.ParseContentType(strings.TrimSpace(s))
			if error != nil {
				continue
			}
			if _, ok := r.ContentTypes[contentType]; ok {
				return contentType, nil
			}
		}
	}

	// no matching content found => returning first in list
	for contentType := range r.ContentTypes {
		// return first element
		return contentType, nil
	}

	return "", fmt.Errorf("No content type found")
}
