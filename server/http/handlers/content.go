package handlers

import (
	"fmt"
	"io/ioutil"
	"mokapi/middlewares"
	"mokapi/models"
	"mokapi/providers/encoding"
	"mokapi/providers/parser"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ResponseHandler struct {
	middleware middlewares.Middleware
	resources  []*models.Resource
}

func NewResponseHandler(middleware middlewares.Middleware, resources []*models.Resource) *ResponseHandler {
	return &ResponseHandler{middleware: middleware, resources: resources}
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

	schema := response.ContentTypes[contentType.Key()].Schema
	requestContentType := models.NewContentType(context.Request.Header.Get("content-type"))

	dataContext := &middlewares.Context{Parameters: context.Parameters, Schema: schema, Body: &middlewares.Body{ContentType: requestContentType}}
	if context.Request.ContentLength > 0 {
		body, err := ioutil.ReadAll(context.Request.Body)
		if err != nil {
			log.Errorf("Error reading body: %v", err)
		} else {
			dataContext.Body.Content = string(body) // should be dataContext.Body.Content []byte?
		}
	}

	r := handler.GetResource(dataContext)
	resourceName := ""
	if r != nil {
		resourceName = r.Name
	}

	// get content
	raw, error := context.DataProvider.Provide(resourceName, schema)
	if error != nil {
		// TODO select correct response from config
		http.Error(context.Response, error.Error(), http.StatusInternalServerError)
		return
	} else if raw == nil {
		context.Response.WriteHeader(404)
		return
	}

	data := middlewares.NewData(raw)
	handler.middleware.ServeData(data, dataContext)

	if a, ok := data.Content.([]interface{}); ok && schema.Type != "array" {
		if len(a) == 1 {
			data.Content = a[0]
		} else if len(a) > 1 {
			// TODO select correct response from config
			http.Error(context.Response, "multiple resources found but schema type is not an array", http.StatusInternalServerError)
			return
		} else {
			data.Content = nil
		}
	}

	if data.Content == nil {
		context.Response.WriteHeader(404)
		return
	}

	//encode
	var bytes []byte
	if bytes, ok = data.Content.([]byte); ok {
		if contentType.Subtype == "*" {
			s := http.DetectContentType(bytes)
			contentType = models.NewContentType(s)
		}
	} else {
		bytes, error = handler.Encode(data.Content, contentType, schema)
		if error != nil {
			http.Error(context.Response, error.Error(), http.StatusInternalServerError)
			return
		}
	}

	// set content type
	context.Response.Header().Add("Content-Type", contentType.String())

	// write content
	context.Response.Header().Add("Content-Length", fmt.Sprint(len(bytes)))
	context.Response.Write(bytes)
}

func (handler *ResponseHandler) Encode(data interface{}, contentType *models.ContentType, schema *models.Schema) ([]byte, error) {
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

func getContentType(r *models.Response, c *Context) (*models.ContentType, error) {
	accept := c.Request.Header.Get("accept")

	// search for a matching content type
	if accept != "" {
		for _, s := range strings.Split(accept, ",") {
			contentType := models.NewContentType(s)
			if _, ok := r.ContentTypes[contentType.Key()]; ok {
				return contentType, nil
			}
		}
	}

	// no matching content found => returning first in list
	for contentType := range r.ContentTypes {
		// return first element
		return models.NewContentType(contentType), nil
	}

	return nil, fmt.Errorf("No content type found")
}

func (h *ResponseHandler) GetResource(context *middlewares.Context) *models.Resource {
	for _, r := range h.resources {
		if r.If == nil {
			return r
		}
		match, error := r.If.IsTrue(func(factor string, tag parser.FilterTag) string {
			switch tag {
			case parser.FilterBody:
				s, error := context.Body.Select(factor)
				if error != nil {
					log.Error(error.Error())
					return ""
				}
				return s
			case parser.FilterParameter:
				return context.Parameters[factor]
			default:
				return factor
			}
		})
		if error != nil {
			log.Error(error.Error())
		} else if match {
			return r
		}
	}
	return nil
}
