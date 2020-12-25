package middlewares

import (
	"mokapi/models"
	"mokapi/server/web"
)

type selection struct {
	config *models.Selection
	next   Middleware
}

func NewSelection(config *models.Selection, next Middleware) Middleware {
	m := &selection{config: config, next: next}
	return m
}

func (m *selection) ServeData(request *Request, context *web.HttpContext) {
	if a, ok := request.Data.([]interface{}); ok {
		if m.config.First {
			if len(a) > 0 {
				request.Data = a[0]
			} else {
				request.Data = nil
			}
		} else if m.config.Slice != nil {
			low, high := m.config.Slice.Low, m.config.Slice.High
			if m.config.Slice.High == -1 {
				high = len(a)
			}
			if high > len(a) {
				high = len(a)
			}
			request.Data = a[low:high]
		}
	}

	m.next.ServeData(request, context)
}
