package middlewares

import (
	"mokapi/models"
)

type selection struct {
	config *models.Selection
	next   Middleware
}

func NewSelection(config *models.Selection, next Middleware) Middleware {
	m := &selection{config: config, next: next}
	return m
}

func (m *selection) ServeData(data *Data, context *Context) {
	if a, ok := data.Content.([]interface{}); ok {
		if m.config.First {
			if len(a) > 0 {
				data.Content = a[0]
			} else {
				data.Content = nil
			}
		} else if m.config.Slice != nil {
			low, high := m.config.Slice.Low, m.config.Slice.High
			if m.config.Slice.High == -1 {
				high = len(a)
			}
			if high > len(a) {
				high = len(a)
			}
			data.Content = a[low:high]
		}
	}

	m.next.ServeData(data, context)
}
