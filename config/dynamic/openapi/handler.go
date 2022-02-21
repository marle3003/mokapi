package openapi

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/engine"
	"mokapi/media"
	"mokapi/server/httperror"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

type operationHandler struct {
	config *Config
	next   http.Handler
}

type responseHandler struct {
	config       *Config
	eventEmitter engine.EventEmitter
}

func NewHandler(config *Config, eventEmitter engine.EventEmitter) http.Handler {
	return &operationHandler{config: config, next: &responseHandler{config: config, eventEmitter: eventEmitter}}
}

func (h *responseHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	op, ok := OperationFromContext(r.Context())
	if !ok {
		return
	}

	status, res, err := op.getFirstSuccessResponse()
	if err != nil {
		writeError(rw, r, err)
		return
	}
	contentType, mediaType, err := ContentTypeFromRequest(r, res)
	if err != nil {
		writeError(rw, r, err)
		return
	}

	request := EventRequestFrom(r)
	response := NewEventResponse(status)

	if op.RequestBody != nil {
		body, err := BodyFromRequest(r, op)
		if err != nil {
			writeError(rw, r, err)
			return
		} else if body == nil {
			writeError(rw, r, errors.New("request body expected"))
			return
		}
		request.Body = body
	}

	response.Headers["Content-Type"] = contentType.String()

	gen := schema.NewGenerator()

	if len(mediaType.Examples) > 0 {
		keys := reflect.ValueOf(mediaType.Examples).MapKeys()
		v := keys[rand.Intn(len(keys))].Interface().(*ExampleRef)
		response.Data = v.Value.Value
	} else if mediaType.Example != nil {
		response.Data = mediaType.Example
	} else {
		response.Data = gen.New(mediaType.Schema)
	}

	for k, v := range res.Headers {
		if v.Value == nil {
			log.Warnf("header ref not resovled: %v", v.Ref())
			continue
		}
		response.Headers[k] = fmt.Sprintf("%v", gen.New(v.Value.Schema))
	}

	h.eventEmitter.Emit("http", request, response)

	res = op.getResponse(response.StatusCode)

	for k, v := range response.Headers {
		rw.Header().Add(k, v)
	}

	if response.StatusCode > 0 {
		rw.WriteHeader(response.StatusCode)
	}

	if len(res.Content) == 0 {
		// no response content is defined which means body is empty
		return
	}

	contentType = media.ParseContentType(response.Headers["Content-Type"])
	contentType, mediaType = res.GetContent(contentType)
	if mediaType == nil {
		writeError(rw, r, fmt.Errorf("response has no definition for content type: %v", contentType))
		return
	}
	var body []byte

	if len(response.Body) > 0 {
		body = []byte(response.Body)
	} else if response.Data != nil {
		if b, ok := response.Data.([]byte); ok {
			body = b
			if contentType.Subtype == "*" {
				contentType = media.ParseContentType(http.DetectContentType(b))
			}
		} else {
			body, err = mediaType.Schema.Marshal(response.Data, contentType)
			if err != nil {
				writeError(rw, r, err)
				return
			}
		}
	}

	_, err = rw.Write(body)
	if err != nil {
		log.Errorf("unable to write body: %v", err)
	}
}

func (h *operationHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	regex := regexp.MustCompile(`\{(?P<name>.+)\}`) // parameter format "/{param}/"
	reqSeg := strings.Split(r.URL.Path, "/")
	var lastError error

endpointLoop:
	for path, ref := range h.config.EndPoints {
		if ref.Value == nil {
			continue
		}
		endpoint := ref.Value
		op := getOperation(r.Method, endpoint)
		if op == nil {
			continue
		}

		routePath := path
		servicePath, ok := r.Context().Value("servicePath").(string)
		if ok && servicePath != "/" {
			routePath = servicePath + routePath
		}
		routeSeg := strings.Split(routePath, "/")

		if len(reqSeg) != len(routeSeg) {
			continue
		}

		for i, s := range routeSeg {
			if len(regex.FindStringSubmatch(s)) > 1 {
				continue // validate in parseParams
			} else if s != reqSeg[i] {
				continue endpointLoop
			}
		}

		params := append(endpoint.Parameters, op.Parameters...)
		rp, err := parameter.FromRequest(params, routePath, r)
		if err != nil {
			lastError = httperror.New(400, err.Error())
			continue
		}

		r = r.WithContext(parameter.NewContext(r.Context(), rp))
		r = r.WithContext(NewOperationContext(r.Context(), op))
		r = r.WithContext(context.WithValue(r.Context(), "endpointPath", path))

		h.next.ServeHTTP(rw, r)
		return
	}

	if lastError == nil {
		writeError(rw, r, httperror.Newf(http.StatusNotFound, "no matching endpoint found at %v", r.URL))
	} else {
		writeError(rw, r, lastError)
	}
}

// Gets the operation for the given method name
func getOperation(method string, e *Endpoint) *Operation {
	switch strings.ToUpper(method) {
	case http.MethodGet:
		return e.Get
	case http.MethodPost:
		return e.Post
	case http.MethodPut:
		return e.Put
	case http.MethodPatch:
		return e.Patch
	case http.MethodDelete:
		return e.Delete
	case http.MethodHead:
		return e.Head
	case http.MethodOptions:
		return e.Options
	case http.MethodTrace:
		return e.Trace
	}

	return nil
}

func writeError(rw http.ResponseWriter, r *http.Request, err error) {
	message := err.Error()
	var status int

	if hErr, ok := err.(*httperror.Error); ok {
		status = hErr.StatusCode
	} else {
		status = http.StatusInternalServerError
	}

	entry := log.WithFields(log.Fields{"url": r.URL, "method": r.Method, "status": status})
	if status == http.StatusInternalServerError {
		entry.Error(message)
	} else {
		entry.Info(message)
	}
	http.Error(rw, message, status)
}
