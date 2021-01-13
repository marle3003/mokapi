package web

import (
	"github.com/pkg/errors"
	"mokapi/models"
	"strings"
)

type pathParam struct {
	ctx   *HttpContext
	param *models.Parameter
	s     string
}

func newPathParam(param *models.Parameter, s string, ctx *HttpContext) *pathParam {
	return &pathParam{param: param, s: s, ctx: ctx}
}

func (q *pathParam) parse() (interface{}, error) {
	switch q.param.Schema.Type {
	case "array":
		return q.parseArray()
	case "object":
		return q.parseObject()
	}

	if len(q.s) == 0 && q.param.Required {
		return nil, errors.Errorf("required parameter not found")
	}

	return q.ctx.parse(q.s, q.param.Schema)
}

func (q *pathParam) parseObject() (obj map[string]interface{}, err error) {
	switch q.param.Style {
	case "spaceDelimited", "pipeDelimited":
		return nil, errors.Errorf("not supported object style '%v'", q.param.Style)
	default:
		obj = make(map[string]interface{})
		values := strings.Split(q.s, ",")
		if q.param.Explode {
			for _, i := range values {
				kv := strings.Split(i, "=")
				if len(kv) != 2 {
					return nil, errors.Errorf("invalid format")
				}
				p, ok := q.param.Schema.Properties[kv[0]]
				if !ok {
					return nil, errors.Errorf("property '%v' not defined in schema", kv[0])
				}

				if v, err := q.ctx.parse(kv[1], p); err == nil {
					obj[kv[0]] = v
				} else {
					return nil, err
				}
			}
		} else {
			i := 0
			for {
				if i >= len(values) {
					break
				}
				key := values[i]
				p, ok := q.param.Schema.Properties[key]
				if !ok {
					return nil, errors.Errorf("property '%v' not defined in schema", key)
				}
				i++
				if i >= len(values) {
					return nil, errors.Errorf("invalid number of property pairs")
				}
				if v, err := q.ctx.parse(values[i], p); err == nil {
					obj[key] = v
				} else {
					return nil, err
				}
			}
		}
	}
	return
}

func (q *pathParam) parseArray() (result []interface{}, err error) {
	switch q.param.Style {
	case "spaceDelimited", "pipeDelimited", "deepObject":
		return nil, errors.Errorf("not supported arrray style '%v'", q.param.Style)
	default:
		values := strings.Split(q.s, ",")

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
