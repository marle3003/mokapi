package openapi

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/engine/common"
	"mokapi/media"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type operationHandler struct {
	config *Config
	next   http.Handler
}

type responseHandler struct {
	config       *Config
	eventEmitter common.EventEmitter
	g            *schema.Generator
}

func NewHandler(config *Config, eventEmitter common.EventEmitter) http.Handler {
	return &operationHandler{
		config: config,
		next: &responseHandler{
			config:       config,
			eventEmitter: eventEmitter,
		},
	}
}

func (h *responseHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	op, ok := OperationFromContext(r.Context())
	if !ok {
		return
	}
	logHttp, ok := LogEventFromContext(r.Context())
	if logHttp != nil {
		defer func() {
			i := r.Context().Value("time")
			if i != nil {
				t := i.(time.Time)
				logHttp.Duration = time.Now().Sub(t).Milliseconds()
			}
			for k, v := range rw.Header() {
				if _, ok := logHttp.Response.Headers[k]; !ok {
					logHttp.Response.Headers[k] = strings.Join(v, ",")
				}
			}
		}()
	}

	status, res, err := op.getFirstSuccessResponse()
	if err != nil {
		writeError(rw, r, err, h.config.Info.Name)
		return
	}
	contentType, mediaType, err := ContentTypeFromRequest(r, res)
	if err != nil {
		writeError(rw, r, err, h.config.Info.Name)
		return
	}

	request := EventRequestFrom(r)
	response := NewEventResponse(status)

	if op.RequestBody != nil {
		body, err := BodyFromRequest(r, op)
		if err != nil {
			writeError(rw, r, err, h.config.Info.Name)
			return
		} else if body != nil {
			request.Body = body.Value
			logHttp.Request.Body = body.Raw
		}
	}

	if !contentType.IsEmpty() {
		response.Headers["Content-Type"] = contentType.String()
	}

	if mediaType != nil {
		if len(mediaType.Examples) > 0 {
			keys := reflect.ValueOf(mediaType.Examples).MapKeys()
			v := keys[rand.Intn(len(keys))].Interface().(*ExampleRef)
			response.Data = v.Value.Value
		} else if mediaType.Example != nil {
			response.Data = mediaType.Example
		} else {
			if data, err := h.g.New(mediaType.Schema); err != nil {
				log.Infof("unable to generate response data for %v: %v", r.URL.String(), err)
			} else {
				response.Data = data
			}
		}
	}

	for k, v := range res.Headers {
		if v.Value == nil {
			log.Warnf("header ref not resovled: %v", v.Ref)
			continue
		}
		if data, err := h.g.New(v.Value.Schema); err != nil {
			log.Infof("unable to generate response header %v for %v: %v", k, r.URL.String(), err)
		} else {
			response.Headers[k] = fmt.Sprintf("%v", data)
		}
	}

	actions := h.eventEmitter.Emit("http", request, response)

	res = op.getResponse(response.StatusCode)
	if res == nil {
		writeError(rw, r,
			fmt.Errorf(
				"no configuration was found for HTTP status code %v, https://swagger.io/docs/specification/describing-responses",
				response.StatusCode),
			h.config.Info.Name)
		return
	}

	for k, v := range response.Headers {
		rw.Header().Add(k, v)
		if logHttp != nil {
			logHttp.Response.Headers[k] = v
		}
	}

	if len(res.Content) == 0 {
		// no response content is defined which means body is empty
		if response.StatusCode > 0 {
			rw.WriteHeader(response.StatusCode)
			if logHttp != nil {
				logHttp.Response.StatusCode = response.StatusCode
			}
		}
		return
	}

	contentType = media.ParseContentType(response.Headers["Content-Type"])
	mediaType = res.GetContent(contentType)
	if mediaType == nil {
		writeError(rw, r, fmt.Errorf("response has no definition for content type: %v", contentType), h.config.Info.Name)
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
				writeError(rw, r, err, h.config.Info.Name)
				return
			}
		}
	}

	if response.StatusCode > 0 {
		rw.WriteHeader(response.StatusCode)
		if logHttp != nil {
			logHttp.Response.StatusCode = response.StatusCode
		}
	}

	_, err = rw.Write(body)
	if err != nil {
		log.Errorf("unable to write body: %v", err)
	}
	if logHttp != nil {
		logHttp.Response.Body = string(body)
		logHttp.Response.Size = len(body)
		logHttp.Actions = actions
	}
}

func (h *operationHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	regex := regexp.MustCompile(`\{(?P<name>.+)\}`) // parameter format "/{param}/"
	reqSeg := strings.Split(r.URL.Path, "/")
	var lastError error

endpointLoop:
	for path, ref := range h.config.Paths.Value {
		if ref.Value == nil {
			continue
		}
		endpoint := ref.Value
		op := endpoint.getOperation(r.Method)
		if op == nil {
			continue
		}

		routePath := ""
		if path != "/" {
			routePath = path
		}
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
			lastError = newHttpError(http.StatusBadRequest, err.Error())
			continue
		}

		r = r.WithContext(parameter.NewContext(r.Context(), rp))
		r = r.WithContext(NewOperationContext(r.Context(), op))
		r = r.WithContext(context.WithValue(r.Context(), "endpointPath", path))

		if m, ok := monitor.HttpFromContext(r.Context()); ok {
			m.LastRequest.WithLabel(h.config.Info.Name, path).Set(float64(time.Now().Unix()))
			m.RequestCounter.WithLabel(h.config.Info.Name, path).Add(1)
		}

		if ctx, err := NewLogEventContext(r, op.Deprecated, events.NewTraits().WithName(h.config.Info.Name).With("path", path).With("method", r.Method)); err != nil {
			log.Errorf("unable to log http event: %v", err)
		} else {
			r = r.WithContext(ctx)
		}

		h.next.ServeHTTP(rw, r)
		return
	}

	if ctx, err := NewLogEventContext(r, false, events.NewTraits().WithName(h.config.Info.Name)); err != nil {
		log.Errorf("unable to log http event: %v", err)
	} else {
		r = r.WithContext(ctx)
	}

	if lastError == nil {
		writeError(rw, r, newHttpErrorf(http.StatusNotFound, "no matching endpoint found: %v %v", strings.ToUpper(r.Method), r.URL), h.config.Info.Name)
	} else {
		writeError(rw, r, lastError, h.config.Info.Name)
	}
}

func writeError(rw http.ResponseWriter, r *http.Request, err error, serviceName string) {
	message := err.Error()
	var status int

	if hErr, ok := err.(*httpError); ok {
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
	if m, ok := monitor.HttpFromContext(r.Context()); ok {
		endpointPath := r.Context().Value("endpointPath")
		if endpointPath != nil {
			endpointPathString := endpointPath.(string)
			m.RequestErrorCounter.WithLabel(serviceName, endpointPathString).Add(1)
			m.LastRequest.WithLabel(serviceName, endpointPathString).Set(float64(time.Now().Unix()))
		} else {
			m.RequestErrorCounter.WithLabel(serviceName, "").Add(1)
			m.LastRequest.WithLabel(serviceName, "").Set(float64(time.Now().Unix()))
		}
	}
	rw.Header().Add("Content-Type", "text/plain")
	http.Error(rw, message, status)
	logHttp, ok := LogEventFromContext(r.Context())
	if ok {
		logHttp.Response.StatusCode = status
		logHttp.Response.Body = message
		logHttp.Response.Size = len([]byte(message))
		for k, v := range rw.Header() {
			logHttp.Response.Headers[k] = strings.Join(v, ",")
		}
	}
}
