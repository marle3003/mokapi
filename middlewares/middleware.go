package middlewares

import (
	"fmt"
	"mokapi/models"
	"strings"

	"gopkg.in/xmlpath.v2"
)

type Middleware interface {
	ServeData(data *Data, context *Context)
}

type EmptyMiddleware struct {
}

func NewEmptyMiddleware() Middleware {
	return &EmptyMiddleware{}
}

func (m *EmptyMiddleware) ServeData(data *Data, context *Context) {
}

type Data struct {
	Content interface{}
}

func NewData(content interface{}) *Data {
	return &Data{Content: content}
}

type Context struct {
	Parameters map[string]string
	Schema     *models.Schema
	Body       *Body
}

type Body struct {
	Content     string
	ContentType string
}

func (b *Body) Select(selector string) (string, error) {
	switch b.ContentType {
	case "application/xml", "text/xml":
		path, error := xmlpath.Compile(selector)
		if error != nil {
			return "", fmt.Errorf("Expecting xpath as selector with content type %v", b.ContentType)
		}
		r := strings.NewReader(b.Content)
		x, error := xmlpath.Parse(r)
		if error != nil {
			return "", fmt.Errorf("Error in xml parsing request body")
		}
		if v, ok := path.String(x); ok {
			return v, nil
		}
	default:
		return "", fmt.Errorf("Content type '%v' of request body is not supported", b.ContentType)
	}

	return "", nil
}
