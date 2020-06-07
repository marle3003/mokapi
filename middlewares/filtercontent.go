package middlewares

import (
	"fmt"
	"mokapi/models"
	"mokapi/providers/parser"

	log "github.com/sirupsen/logrus"
)

type filterContent struct {
	filter *parser.Expression
	next   Middleware
}

func NewFilterContent(config *models.FilterContent, next Middleware) Middleware {
	m := &filterContent{filter: config.Filter.Expr, next: next}
	return m
}

func (m *filterContent) ServeData(data *Data, context *Context) {
	if list, ok := data.Content.([]interface{}); ok {
		result := make([]interface{}, 0)
		for _, d := range list {
			match, error := m.filter.IsTrue(func(factor string, tag parser.ExpressionTag) string {
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
				case parser.Property:
					if data != nil {
						o := d.(map[string]interface{})
						if v, ok := o[factor]; ok {
							return fmt.Sprint(v)
						}
					}
					return ""
				default:
					return factor
				}
			})
			if error != nil {
				log.Error(error.Error())
				continue
			} else if match {
				result = append(result, d)
			}
		}
		if len(result) == 0 {
			data.Content = nil
		}
		data.Content = result
	}

	if m.next != nil {
		m.next.ServeData(data, context)
	}
}
