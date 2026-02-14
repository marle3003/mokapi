package common

import (
	"fmt"
	"reflect"
	"strings"
)

type HttpEventResponse struct {
	Headers    map[string]any `json:"headers"`
	StatusCode int            `json:"statusCode"`
	Body       string         `json:"body"`
	Data       any            `json:"data"`
}

type HttpEventRequest struct {
	Method      string         `json:"method"`
	Url         Url            `json:"url"`
	Body        interface{}    `json:"body"`
	Path        map[string]any `json:"path"`
	Query       map[string]any `json:"query"`
	Header      map[string]any `json:"header"`
	Cookie      map[string]any `json:"cookie"`
	QueryString any            `json:"querystring"`

	Api         string `json:"api"`
	Key         string `json:"key"`
	OperationId string `json:"operationId"`
}

type Url struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Path   string `json:"path"`
	Query  string `json:"query"`
}

func (r *HttpEventRequest) String() string {
	s := r.Method + " " + r.Url.String()
	if r.Api != "" {
		s += fmt.Sprintf(" [API: %s]", r.Api)
	}
	if r.OperationId != "" {
		s += fmt.Sprintf(" [OperationId: %s]", r.OperationId)
	}
	return s
}

func (u Url) String() string {
	sb := strings.Builder{}
	sb.WriteString(u.Scheme)
	if sb.Len() > 0 {
		sb.WriteString("://")
	}
	sb.WriteString(u.Host)
	sb.WriteString(u.Path)
	if len(u.Query) > 0 {
		sb.WriteString("?" + u.Query)
	}
	return sb.String()
}

func (r *HttpEventResponse) HasBody() bool {
	return len(r.Body) > 0 || r.Data != nil
}

func HttpEventHandler(req *HttpEventRequest, res *HttpEventResponse, resources interface{}) (bool, error) {
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
