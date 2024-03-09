package openapi

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"mokapi/engine/common"
	"mokapi/json/generator"
	"mokapi/media"
	"mokapi/providers/openapi/parameter"
	"mokapi/providers/openapi/schema"
	"net/http"
	"reflect"
	"strings"
)

func NewEventResponse(status int, ct media.ContentType) *common.EventResponse {
	r := &common.EventResponse{
		Headers:    make(map[string]string),
		StatusCode: status,
	}

	if !ct.IsEmpty() {
		r.Headers["Content-Type"] = ct.String()
	}

	return r
}

func EventRequestFrom(r *http.Request) *common.EventRequest {
	ctx := r.Context()
	endpointPath := ctx.Value("endpointPath").(string)
	op, _ := OperationFromContext(ctx)

	req := &common.EventRequest{
		Key:         endpointPath,
		OperationId: op.OperationId,
		Method:      r.Method,
		Path:        make(map[string]interface{}),
		Query:       make(map[string]interface{}),
		Header:      make(map[string]interface{}),
		Cookie:      make(map[string]interface{}),
	}

	req.Url = common.Url{
		Scheme: "",
		Host:   r.Host,
		Path:   r.URL.Path,
		Query:  r.URL.RawQuery,
	}

	if strings.HasPrefix(r.Proto, "HTTPS") {
		req.Url.Scheme = "https"
	} else if strings.HasPrefix(r.Proto, "HTTP") {
		req.Url.Scheme = "http"
	}

	// Mokapi's goal is to provide better APIs
	// Therefore, we only add headers that defined in specification
	params, _ := parameter.FromContext(ctx)
	for t, values := range params {
		for k, v := range values {
			switch t {
			case parameter.Path:
				req.Path[k] = v.Value
			case parameter.Query:
				req.Query[k] = v.Value
			case parameter.Header:
				req.Header[k] = v.Value
			case parameter.Cookie:
				req.Cookie[k] = v.Value
			}
		}
	}

	return req
}

func setResponseData(r *common.EventResponse, m *MediaType, path string) error {
	if m != nil {
		if len(m.Examples) > 0 {
			keys := reflect.ValueOf(m.Examples).MapKeys()
			v := keys[rand.Intn(len(keys))].Interface().(*ExampleRef)
			r.Data = v.Value.Value
		} else if m.Example != nil {
			r.Data = m.Example
		} else {
			segments := strings.Split(path, "/")
			var names []string
			for _, seg := range segments[1:] {
				if !strings.HasPrefix(seg, "{") {
					names = append(names, seg)
				}
			}

			data, err := generator.New(generator.NewRequest(generator.Name(names...), generator.Ref(schema.ConvertToJsonSchema(m.Schema))))
			if err != nil {
				return fmt.Errorf("generate response data failed: %v", err)
			} else {
				r.Data = data
			}
		}
	}
	return nil
}

func setResponseHeader(r *common.EventResponse, headers Headers) error {
	for k, v := range headers {
		if v.Value == nil {
			log.Warnf("header ref not resovled: %v", v.Ref)
			continue
		}
		if data, err := schema.CreateValue(v.Value.Schema); err != nil {
			return fmt.Errorf("set response header '%v' failed: %v", k, err)
		} else {
			r.Headers[k] = fmt.Sprintf("%v", data)
		}
	}
	return nil
}
