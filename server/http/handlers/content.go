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
	"time"

	log "github.com/sirupsen/logrus"
)

type ResponseHandler struct {
	middleware middlewares.Middleware
	resources  []*models.Resource

	requestChannel chan *models.RequestMetric
	metric         *models.RequestMetric
}

func NewResponseHandler(middleware middlewares.Middleware, resources []*models.Resource, requestChannel chan *models.RequestMetric) *ResponseHandler {
	return &ResponseHandler{middleware: middleware, resources: resources, requestChannel: requestChannel, metric: &models.RequestMetric{}}
}

func (handler *ResponseHandler) ServeHTTP(context *Context) {
	startTime := time.Now()
	log.WithFields(log.Fields{"url": context.Request.URL.String(), "host": context.Request.Host, "method": context.Request.Method}).Info("Serve http request")
	handler.metric.Method = context.Request.Method
	handler.metric.Url = context.Request.URL.String()

	// first we select the possible response. With that schema we selects the data.
	// In OpenApi each http status code and content type can defined its own schema
	// Depending on data selection we change http status and schema of that http status code definition

	response, ok := context.Responses[models.Ok]
	if !ok {
		handler.error("No 200 response in configuration found", context)
		return
	}

	contentType, error := getContentType(response, context)
	if error != nil {
		handler.error(error.Error(), context)
		return
	}

	schema := response.ContentTypes[contentType.Key()].Schema
	requestContentType := models.NewContentType(context.Request.Header.Get("content-type"))

	dataContext := &middlewares.Context{Parameters: context.Parameters, Schema: schema, Body: &middlewares.Body{ContentType: requestContentType}}
	if context.Request.ContentLength > 0 {
		body, err := ioutil.ReadAll(context.Request.Body)
		if err != nil {
			handler.error(fmt.Sprintf("Error while reading body: %v", err.Error()), context)
			return
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
		handler.error(error.Error(), context)
		return
	} else if raw == nil {
		log.Infof("No data found responding with 404")
		context.Response.WriteHeader(404)
		handler.metric.ResponseTime = time.Now().Sub(startTime)
		handler.metric.HttpStatus = int(models.NotFound)
		handler.requestChannel <- handler.metric
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
			handler.metric.ResponseTime = time.Now().Sub(startTime)
			handler.metric.HttpStatus = int(models.NotFound)
			handler.requestChannel <- handler.metric
			return
		} else {
			data.Content = nil
		}
	}

	if data.Content == nil {
		log.Errorf("Data content is nil responding with 404")
		handler.metric.ResponseTime = time.Now().Sub(startTime)
		handler.metric.HttpStatus = int(models.NotFound)
		handler.requestChannel <- handler.metric
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
			handler.error(error.Error(), context)
			handler.metric.ResponseTime = time.Now().Sub(startTime)
			handler.metric.HttpStatus = int(models.InternalServerError)
			handler.requestChannel <- handler.metric
			return
		}
	}

	// set content type
	context.Response.Header().Add("Content-Type", contentType.String())

	// write content
	context.Response.Header().Add("Content-Length", fmt.Sprint(len(bytes)))
	context.Response.Write(bytes)

	handler.metric.ResponseTime = time.Now().Sub(startTime)
	handler.metric.HttpStatus = int(models.Ok)
	handler.requestChannel <- handler.metric
}

func (handler *ResponseHandler) Encode(data interface{}, contentType *models.ContentType, schema *models.Schema) ([]byte, error) {
	if s, ok := data.(string); ok {
		return []byte(s), nil
	}

	switch contentType.Subtype {
	case "json":
		return encoding.MarshalJSON(data, schema)
	case "xml", "rss+xml":
		return encoding.MarshalXML(data, schema)
	}

	return nil, fmt.Errorf("Unspupported encoding for content type %v", contentType)
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

		if !r.If.IsValid() {
			log.Infof("Skipping invalid expression: %v", r.If.Raw)
			continue
		}

		match, error := r.If.Expr.IsTrue(func(factor string, tag parser.ExpressionTag) string {
			switch tag {
			case parser.Body:
				s, error := context.Body.Select(factor)
				if error != nil {
					log.Error(error.Error())
					return ""
				}
				return s
			case parser.Parameter:
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

func (h *ResponseHandler) error(message string, context *Context) {
	log.WithFields(log.Fields{"url": context.Request.URL.String()}).Errorf(message)
	http.Error(context.Response, message, http.StatusInternalServerError)
	h.metric.Error = message
	h.requestChannel <- h.metric
}
