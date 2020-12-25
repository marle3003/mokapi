package middlewares

import (
	"fmt"
	"mokapi/models"
	"mokapi/providers/parser"
	"mokapi/server/web"

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

func (m *filterContent) ServeData(request *Request, context *web.HttpContext) {
	if list, ok := request.Data.([]interface{}); ok {
		result := make([]interface{}, 0)
		for _, d := range list {
			match, error := m.filter.IsTrue(func(factor string, tag parser.ExpressionTag) string {
				switch tag {
				case parser.Body:
					s, error := context.SelectFromBody(factor)
					if error != nil {
						log.Error(error.Error())
						return ""
					}
					return s
				case parser.Parameter:
					return context.Parameters[factor]
				case parser.Property:
					if request.Data != nil {
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
			request.Data = nil
		}
		request.Data = result
	}

	if m.next != nil {
		m.next.ServeData(request, context)
	}
}
