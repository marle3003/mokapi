package openapi

import (
	"context"
	"fmt"
	"math/rand"
	"mokapi/engine/common"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/json/generator"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const eventKey = "event"

func NewEventResponse(status int, ct media.ContentType) *common.EventResponse {
	r := &common.EventResponse{
		Headers:    make(map[string]any),
		StatusCode: status,
	}

	if !ct.IsEmpty() {
		r.Headers["Content-Type"] = ct.String()
	}

	return r
}

func EventRequestFromContext(ctx context.Context) *common.EventRequest {
	e := ctx.Value(eventKey).(*common.EventRequest)
	return e
}

func NewEventRequest(r *http.Request, contentType media.ContentType, api string) (*common.EventRequest, context.Context) {
	ctx := r.Context()
	endpointPath := ctx.Value("endpointPath").(string)
	op, _ := OperationFromContext(ctx)

	req := &common.EventRequest{
		Api:         api,
		Key:         endpointPath,
		OperationId: op.OperationId,
		Method:      r.Method,
		Path:        make(map[string]any),
		Query:       make(map[string]any),
		Header:      make(map[string]any),
		Cookie:      make(map[string]any),
	}

	req.Url = common.Url{
		Scheme: "",
		Path:   r.URL.Path,
		Query:  r.URL.RawQuery,
	}
	req.Url.Host, req.Url.Port = getHostAndPort(r)

	if strings.HasPrefix(r.Proto, "HTTPS") {
		req.Url.Scheme = "https"
	} else if strings.HasPrefix(r.Proto, "HTTP") {
		req.Url.Scheme = "http"
	}

	setParam := func(target map[string]any, params map[string]RequestParameterValue) {
		for k, v := range params {
			target[k] = v.Value
		}
	}

	// Mokapi's goal is to provide better APIs.
	// Therefore, we only add headers that are defined in the specification
	params, _ := FromContext(ctx)
	setParam(req.Path, params.Path)
	setParam(req.Query, params.Query)
	setParam(req.Header, params.Header)
	setParam(req.Cookie, params.Cookie)
	if params.QueryString != nil {
		req.QueryString = params.QueryString
	}

	// Accept header is defined in the response object and not as parameter
	req.Header["Accept"] = contentType.String()

	return req, context.WithValue(ctx, eventKey, req)
}

func setResponseData(r *common.EventResponse, m *MediaType, request *common.EventRequest) error {
	if m != nil {
		if len(m.Examples) > 0 {
			keys := reflect.ValueOf(m.Examples).MapKeys()
			key := keys[rand.Intn(len(keys))].String()
			v := m.Examples[key]
			if v.Value != nil {
				r.Data = v.Value.Value
				return nil
			}
		} else if m.Example != nil {
			r.Data = m.Example.Value
			return nil
		}
		segments := strings.Split(request.Key, "/")
		var names []string
		for _, seg := range segments[1:] {
			if !strings.HasPrefix(seg, "{") {
				names = append(names, seg)
			}
		}

		req := generator.NewRequest(
			names,
			schema.ConvertToJsonSchema(m.Schema),
			getGeneratorContext(request),
		)
		data, err := generator.New(req)
		if err != nil {
			return fmt.Errorf("generate response data failed: %v", err)
		} else {
			r.Data = data
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

func getGeneratorContext(r *common.EventRequest) map[string]interface{} {
	ctx := map[string]interface{}{}
	for k, v := range r.Cookie {
		if v != nil {
			ctx[k] = v
		}
	}
	for k, v := range r.Header {
		if v != nil {
			ctx[k] = v
		}
	}
	for k, v := range r.Path {
		if v != nil {
			ctx[k] = v
		}
	}
	for k, v := range r.Query {
		if v != nil {
			ctx[k] = v
		}
	}
	return ctx
}

func getHostAndPort(r *http.Request) (string, int) {
	// 1. Try to extract from Host header (can include port)
	host := r.Host
	portString := ""
	if strings.Contains(host, ":") {
		hostAndPort := strings.Split(host, ":")
		host = hostAndPort[0]
		portString = hostAndPort[1]
	} else {
		// 2. Try from URL
		portString = r.URL.Port()
		if portString == "" {
			// 3. Default based on scheme
			if r.TLS != nil {
				portString = "443"
			} else {
				portString = "80"
			}
		}
	}

	port, _ := strconv.Atoi(portString)
	return host, port
}
