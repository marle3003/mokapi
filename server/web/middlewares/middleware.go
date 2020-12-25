package middlewares

import (
	"mokapi/models"
	"mokapi/server/web"
)

type Middleware interface {
	ServeData(request *Request, context *web.HttpContext)
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
		if m, ok := item.(*models.Delay); ok {
			middleware = NewDelay(m, middleware)
		}
	}

	return middleware
}

func newEmptyMiddleware() Middleware {
	return &emptyMiddleware{}
}

func (m *emptyMiddleware) ServeData(request *Request, context *web.HttpContext) {
}

type Request struct {
	Data interface{}
}

func NewRequest(data interface{}) *Request {
	return &Request{Data: data}
}

// func NewData(content interface{}) *Data {
// 	return &Data{Content: content}
// }

// type Context struct {
// 	Parameters map[string]string
// 	Schema     *data.Schema
// 	Body       *Body
// }

// type Body struct {
// 	Content     string
// 	ContentType *models.ContentType
// }
