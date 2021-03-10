package web

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"mokapi/config/dynamic/mokapi"
	"mokapi/config/dynamic/openapi"
	"mokapi/models"
	"mokapi/models/media"
	"net/http"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ParameterParser interface {
	parse() (interface{}, error)
}

type HttpContext struct {
	Response     http.ResponseWriter
	Request      *http.Request
	Parameters   RequestParameters
	ResponseType *openapi.ResponseRef
	ServicPath   string
	Operation    *openapi.Operation
	ContentType  *media.ContentType
	Schema       *openapi.SchemaRef
	Mokapi       *mokapi.Config
	body         string
	metric       *models.RequestMetric
}

func NewHttpContext(request *http.Request, response http.ResponseWriter, servicePath string) *HttpContext {
	return &HttpContext{Response: response,
		Request:    request,
		ServicPath: servicePath,
	}
}

func (context *HttpContext) Body() string {
	if context.Request.ContentLength == 0 {
		return ""
	}
	if context.body != "" {
		return context.body
	}

	data, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		log.Errorf("Error while reading body: %v", err.Error())
		return ""
	}
	context.body = string(data)
	return context.body
}

func (context *HttpContext) Init() error {
	error := context.setFirstSuccessResponse(context.Operation)
	if error != nil {
		return error
	}
	error = context.setContentType()
	if error != nil {
		return error
	}
	context.Schema = context.ResponseType.Value.Content[context.ContentType.Key()].Schema

	return nil
}

func (context *HttpContext) setFirstSuccessResponse(operation *openapi.Operation) error {
	successStatus := make([]openapi.HttpStatus, 0, 1)
	for httpStatus := range operation.Responses {
		if httpStatus >= 200 && httpStatus < 300 {
			successStatus = append(successStatus, httpStatus)
		}
	}

	if len(successStatus) == 0 {
		return fmt.Errorf("no success response in configuration found")
	}

	sort.SliceStable(successStatus, func(i, j int) bool { return i < j })

	context.ResponseType = operation.Responses[successStatus[0]]
	return nil
}

func (context *HttpContext) setContentType() error {
	accept := context.Request.Header.Get("accept")

	// search for a matching content type
	if accept != "" {
		for _, mimeType := range strings.Split(accept, ",") {
			contentType := media.ParseContentType(mimeType)
			if _, ok := context.ResponseType.Value.Content[contentType.Key()]; ok {
				context.ContentType = contentType
				return nil
			}
		}
	}

	// no matching content found => returning first in list
	// The iteration order over maps is not specified and is not
	// guaranteed to be the same from one iteration to the next
	for i, _ := range context.ResponseType.Value.Content {
		// return first element
		context.ContentType = media.ParseContentType(i)
		return nil
	}

	return errors.New("no content type found")
}
