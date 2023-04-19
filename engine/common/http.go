package common

import (
	"reflect"
	"strings"
)

type EventResponse struct {
	Headers    map[string]string `json:"headers"`
	StatusCode int               `json:"statusCode"`
	Body       string            `json:"body"`
	Data       interface{}       `json:"data"`
}

type EventRequest struct {
	Method string                 `json:"method"`
	Url    Url                    `json:"url"`
	Body   interface{}            `json:"body"`
	Path   map[string]interface{} `json:"path"`
	Query  map[string]interface{} `json:"query"`
	Header map[string]interface{} `json:"header"`
	Cookie map[string]interface{} `json:"cookie"`

	Key         string `json:"key"`
	OperationId string `json:"operationId"`
}

type Url struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
	Query  string `json:"query"`
}

func (r *EventRequest) String() string {
	return r.Method + " " + r.Url.String()
}

func (u Url) String() string {
	sb := strings.Builder{}
	sb.WriteString(u.Scheme)
	if sb.Len() > 0 {
		sb.WriteString("://")
	}
	sb.WriteString(u.Host)
	sb.WriteString(u.Path)
	sb.WriteString("?" + u.Query)
	return sb.String()
}

func (r *EventResponse) HasBody() bool {
	return len(r.Body) > 0 || r.Data != nil
}

func EventHandler(req *EventRequest, res *EventResponse, resources interface{}) (bool, error) {
	resource := getResource(req.Url, resources)
	if resource == nil {
		return false, nil
	}
	res.Data = resource
	return true, nil
}

func getResource(u Url, resources interface{}) interface{} {
	paths := strings.Split(u.Path, "/")
	val := reflect.ValueOf(resources)
	for _, path := range paths[:len(paths)-1] {
		v := val
		if v.Kind() == reflect.Interface {
			v = v.Elem()
		}
		if v.Kind() != reflect.Map {
			break
		}
		v = v.MapIndex(reflect.ValueOf(path))
		if v.IsValid() {
			val = v
		}
	}

	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	if val.Kind() != reflect.Map {
		return nil
	}
	resource := val.MapIndex(reflect.ValueOf(paths[len(paths)-1]))
	if !resource.IsValid() {
		return nil
	}
	return resource.Interface()
}
