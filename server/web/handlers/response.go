package handlers

import (
	"fmt"
	"mokapi/models"
	"mokapi/providers/encoding"
	"mokapi/server/web"
	"net/http"
)

type Response struct {
	httpContext *web.HttpContext
}

func (r *Response) AddHeader(key string, value string) {
	r.httpContext.Response.Header().Add(key, value)
}

func (r *Response) WriteString(s string) error {
	bytes := []byte(s)

	return r.write(bytes)
}

func (r *Response) Write(object interface{}) error {
	bytes, ok := object.([]byte)
	contentType := r.httpContext.ContentType
	var err error
	if ok {
		if contentType.Subtype == "*" {
			// detect content type by data
			contentType = models.ParseContentType(http.DetectContentType(bytes))
		}
	} else {
		bytes, err = r.encodeData(object)
		if err != nil {
			return err
		}
	}

	r.AddHeader("Content-Type", contentType.String())

	return r.write(bytes)
}

func (r *Response) write(bytes []byte) error {
	r.AddHeader("Content-Length", fmt.Sprint(len(bytes)))
	_, err := r.httpContext.Response.Write(bytes)

	return err
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
