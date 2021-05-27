package web

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
	"mokapi/providers/encoding"
	"mokapi/providers/workflow/runtime"
	"mokapi/server/event"
	"net/http"
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

	//var bodyParam interface{} = nil
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

		var bodySchema *openapi.SchemaRef
		if c, ok := operation.RequestBody.Value.Content[contentType.String()]; ok {
			if c.Schema == nil {
				respond(fmt.Sprintf("schema of request body %q is not defined", contentType.String()), http.StatusInternalServerError, context)
				return
			}
			bodySchema = c.Schema
		} else if c, ok := operation.RequestBody.Value.Content[contentType.Key()]; ok {
			if c.Schema == nil {
				respond(fmt.Sprintf("schema of response %q is not defined", contentType.String()), http.StatusInternalServerError, context)
				return
			}
			bodySchema = c.Schema
		}

		if bodySchema != nil {
			//bodyParam, err = parseBody(body, contentType, bodySchema)
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

	//pipelineName := operation.Endpoint.Pipeline
	//if len(operation.Pipeline) > 0 {
	//	pipelineName = operation.Pipeline
	//}

	gen := openapi.NewGenerator()

	r := &Response{
		Headers: map[string]string{
			"Content-Type": context.ContentType.String(),
		},
		StatusCode: int(context.statusCode),
		Data:       gen.New(context.Schema),
	}

	context.workflowHandler(event.WithHttpEvent(event.HttpEvent{
		Service: context.ServiceName,
		Method:  context.Request.Method,
		Path:    context.Request.URL.Path,
	}),
		runtime.WithContext("response", r),
		runtime.WithAction("set-response", r))

	write(r, context)

	//if context.Mokapi != nil && len(pipelineName) > 0 {
	//	err := pipeline.RunConfig(
	//		pipelineName,
	//		context.Mokapi,
	//		pipeline.WithGlobalVars(map[types.Type]interface{}{
	//			"response":  r,
	//			"Operation": operation,
	//		}),
	//		pipeline.WithParams(map[string]interface{}{
	//			"method": context.Request.Method,
	//			"body":   bodyParam,
	//			"query":  context.Parameters[openapi.QueryParameter],
	//			"path":   context.Parameters[openapi.PathParameter],
	//			"header": context.Parameters[openapi.HeaderParameter],
	//			"cookie": context.Parameters[openapi.CookieParameter],
	//		}),
	//	)
	//	if err != nil {
	//		respond(err.Error(), http.StatusInternalServerError, context)
	//	}
	//} else {
	//	r.WriteRandom(int(context.statusCode), context.ContentType.String())
	//}
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
	log.WithFields(log.Fields{"url": ctx.Request.URL.String()}).Errorf("path %q, operation %q: %v", ctx.Request.URL, ctx.Request.Method, message)
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

func write(r *Response, ctx *HttpContext) error {
	var body []byte
	contentType := media.ParseContentType(r.Headers["Content-Type"])

	if len(r.Body) > 0 {
		body = []byte(r.Body)
	} else {
		if bytes, ok := r.Data.([]byte); ok {
			if contentType.Subtype == "*" {
				// detect content type by data
				contentType = media.ParseContentType(http.DetectContentType(bytes))
			}
			body = bytes
		} else {
			if bytes, err := encodeData(r.Data, contentType, ctx.Schema); err != nil {
				return err
			} else {
				body = bytes
			}
		}
	}

	for k, v := range r.Headers {
		ctx.Response.Header().Add(k, v)
	}

	if r.StatusCode > 0 {
		ctx.Response.WriteHeader(r.StatusCode)
	}

	_, err := ctx.Response.Write(body)
	return err
}

func encodeData(data interface{}, contentType *media.ContentType, schema *openapi.SchemaRef) ([]byte, error) {
	switch contentType.Subtype {
	case "json":
		return encoding.MarshalJSON(data, schema)
	case "xml", "rss+xml":
		return encoding.MarshalXML(data, schema)
	default:
		if s, ok := data.(string); ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("unspupported encoding for content type %v", contentType)
	}
}
