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

type emptyMiddleware struct {
}

func Create(middlewares []interface{}) Middleware {
	middleware := newEmptyMiddleware()

	for i := len(middlewares) - 1; i >= 0; i-- {
		item := middlewares[i]
		if m, ok := item.(*models.FilterContent); ok {
			middleware = NewFilterContent(m, middleware)
		}
		if m, ok := item.(*models.ReplaceContent); ok {
			middleware = NewReplaceContent(m, middleware)
		}
		if m, ok := item.(*models.Template); ok {
			middleware = NewTemplate(m, middleware)
		}
		if m, ok := item.(*models.Selection); ok {
			middleware = NewSelection(m, middleware)
		}
	}

	return middleware
}

func newEmptyMiddleware() Middleware {
	return &emptyMiddleware{}
}

func (m *emptyMiddleware) ServeData(data *Data, context *Context) {
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
	ContentType *models.ContentType
}

func (b *Body) Select(selector string) (string, error) {
	switch b.ContentType.Subtype {
	case "xml":
		path, error := xmlpath.Compile(selector)
		if error != nil {
			return "", fmt.Errorf("Expecting xpath as selector with content type %v", b.ContentType)
		}
		r := strings.NewReader(b.Content)
		x, error := xmlpath.Parse(r)
		if error != nil {
			return "", fmt.Errorf("Error in xml parsing request body: %v", error.Error())
		}
		if v, ok := path.String(x); ok {
			return v, nil
		}
	default:
		return "", fmt.Errorf("Selection of Content type '%v' of request body is not supported", b.ContentType)
	}

	return "", nil
}
