package web

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
	"mokapi/providers/encoding"
	"net/http"
)

type ErrorHandler func(error, int)

type Response struct {
	eh          ErrorHandler
	httpContext *HttpContext
	g           *openapi.Generator
}

func newResponse(ctx *HttpContext, eh ErrorHandler) *Response {
	return &Response{
		eh:          eh,
		httpContext: ctx,
		g:           openapi.NewGenerator(),
	}
}

func (r *Response) AddHeader(key string, value string) {
	r.httpContext.Response.Header().Add(key, value)
}

func (r *Response) WriteString(s string, statusCode int, contentType string) {
	bytes := []byte(s)

	r.write(bytes, statusCode, contentType)
}

func (r *Response) Write(object interface{}, statusCode int, contentType string) {
	bytes, ok := object.([]byte)
	if len(contentType) == 0 {
		contentType = r.httpContext.ContentType.String()
	}

	var err error
	if ok {
		if r.httpContext.ContentType.Subtype == "*" {
			// detect content type by data
			contentType = media.ParseContentType(http.DetectContentType(bytes)).String()
		}
	} else {
		bytes, err = r.encodeData(object)
		if err != nil {
			err = errors.Wrapf(err, "unable to encode to %q", contentType)
			r.eh(err, http.StatusInternalServerError)
			return
		}
	}

	r.write(bytes, statusCode, contentType)
}

func (r *Response) WriteRandom(statusCode int, contentType string) {
	data := r.g.New(r.httpContext.Schema)
	r.httpContext.metric.HttpStatus = statusCode
	r.Write(data, statusCode, contentType)
}

func (r *Response) write(bytes []byte, statusCode int, contentType string) {
	if statusCode > 0 {
		r.httpContext.Response.WriteHeader(statusCode)
	}
	if len(contentType) == 0 {
		r.AddHeader("Content-Type", r.httpContext.ContentType.String())
	} else {
		r.AddHeader("Content-Type", contentType)
	}

	r.AddHeader("Content-Length", fmt.Sprint(len(bytes)))
	_, err := r.httpContext.Response.Write(bytes)
	if err != nil {
		r.eh(err, http.StatusInternalServerError)
	} else {
		r.httpContext.metric.HttpStatus = statusCode
	}
}

func (r *Response) encodeData(data interface{}) ([]byte, error) {
	switch r.httpContext.ContentType.Subtype {
	case "json":
		return encoding.MarshalJSON(data, r.httpContext.Schema)
	case "xml", "rss+xml":
		return encoding.MarshalXML(data, r.httpContext.Schema)
	default:
		if s, ok := data.(string); ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("unspupported encoding for content type %v", r.httpContext.ContentType)
	}
}
