package web

import (
	"github.com/pkg/errors"
	"mokapi/models"
	"strings"
)

type queryParam struct {
	ctx   *HttpContext
	param *models.Parameter
}

func newQueryParam(param *models.Parameter, ctx *HttpContext) *queryParam {
	return &queryParam{param: param, ctx: ctx}
}

func (q *queryParam) parse() (interface{}, error) {
	switch q.param.Schema.Type {
	case "array":
		return q.parseArray()
	case "object":
		return q.parseObject()
	}

	s := q.ctx.Request.URL.Query().Get(q.param.Name)
	if len(s) == 0 && q.param.Required {
		return nil, errors.Errorf("required parameter not found")
	}

	return q.ctx.parse(s, q.param.Schema)
}

func (q *queryParam) parseObject() (obj map[string]interface{}, err error) {
	switch q.param.Style {
	case "spaceDelimited", "pipeDelimited":
		return nil, errors.Errorf("not supported object style '%v'", q.param.Style)
	default:
		obj = make(map[string]interface{})
		if q.param.Explode {
			for name, p := range q.param.Schema.Properties {
				s := q.ctx.Request.URL.Query().Get(name)
				if v, err := q.ctx.parse(s, p); err == nil {
					obj[name] = v
				} else {
					return nil, err
				}
			}
		} else {
			s := q.ctx.Request.URL.Query().Get(q.param.Name)
			elements := strings.Split(s, ",")
			i := 0
			for {
				if i >= len(elements) {
					break
				}
				key := elements[i]
				p, ok := q.param.Schema.Properties[key]
				if !ok {
					return nil, errors.Errorf("property '%v' not defined in schema", key)
				}
				i++
				if i >= len(elements) {
					return nil, errors.Errorf("invalid number of property pairs")
				}
				if v, err := q.ctx.parse(elements[i], p); err == nil {
					obj[key] = v
				} else {
					return nil, err
				}
			}
		}
	}
	return
}

func (q *queryParam) parseArray() (result []interface{}, err error) {
	switch q.param.Style {
	case "spaceDelimited", "pipeDelimited", "deepObject":
		return nil, errors.Errorf("not supported arrray style '%v'", q.param.Style)
	default:
		var values []string
		if q.param.Explode {
			var ok bool
			values, ok = q.ctx.Request.URL.Query()[q.param.Name]
			if !ok && q.param.Required {
				return nil, errors.Errorf("required parameter not found")
			}

		} else {
			s := q.ctx.Request.URL.Query().Get(q.param.Name)
			if len(s) == 0 && q.param.Required {
				return nil, errors.Errorf("required parameter not found")
			}
			values = strings.Split(s, ",")
		}
		for _, v := range values {
			if i, err := q.ctx.parse(v, q.param.Schema.Items); err != nil {
				return nil, err
			} else {
				result = append(result, i)
			}
		}
	}
	return
}
