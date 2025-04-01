package openapi

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/engine/common"
	"mokapi/media"
	"mokapi/providers/openapi/parameter"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"net/http"
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

	request, ctx := NewEventRequest(r)
	r = r.WithContext(ctx)

	if op.RequestBody != nil {
		body, err := BodyFromRequest(r, op)
		if body != nil {
			if logHttp != nil {
				logHttp.Request.Body = body.Raw
			}
			request.Body = body.Value
		}
		if err != nil {
			writeError(rw, r, err, h.config.Info.Name)
			return
		}
	}

	if len(op.Security) > 0 {
		if err = h.serveSecurity(r, op.Security); err != nil {
			writeError(rw, r, err, h.config.Info.Name)
			return
		}
	} else if len(h.config.Security) > 0 {
		if err = h.serveSecurity(r, h.config.Security); err != nil {
			writeError(rw, r, err, h.config.Info.Name)
			return
		}
	}

	response := NewEventResponse(status, contentType)

	err = setResponseData(response, mediaType, request)
	if err != nil {
		writeError(rw, r, err, h.config.Info.Name)
		return
	}

	err = setResponseHeader(response, res.Headers)
	if err != nil {
		writeError(rw, r, err, h.config.Info.Name)
		return
	}

	actions := h.eventEmitter.Emit("http", request, response)

	if status != response.StatusCode {
		res = op.getResponse(response.StatusCode)
		if res == nil {
			// try to get "default" response
			res = op.getResponse(0)
			if res == nil {
				writeError(rw, r,
					fmt.Errorf(
						"no configuration was found for HTTP status code %v, https://swagger.io/docs/specification/describing-responses",
						response.StatusCode),
					h.config.Info.Name)
				return
			}
		}
	}

	// todo: only specified headers should be written
	for k, v := range response.Headers {
		rw.Header().Add(k, v)
		if logHttp != nil {
			logHttp.Response.Headers[k] = v
		}
	}

	if len(res.Content) == 0 {
		rw.Header().Del("Content-Type")
		rw.Header().Set("Content-Length", "0")
		// no response content is defined which means body is empty
		if response.StatusCode > 0 {
			rw.WriteHeader(response.StatusCode)
			if logHttp != nil {
				logHttp.Response.StatusCode = response.StatusCode
			}
		}
		return
	}

	if ct, ok := response.Headers["Content-Type"]; ok {
		contentType = media.ParseContentType(ct)
	} else {
		contentType, _, err = ContentTypeFromRequest(r, res)
		if err != nil {
			writeError(rw, r, err, h.config.Info.Name)
			return
		}
	}

	mediaType = res.GetContent(contentType)
	if mediaType == nil {
		writeError(rw, r, fmt.Errorf("response has no definition for content type: %v", contentType), h.config.Info.Name)
		return
	}
	var body []byte

	if len(response.Body) > 0 {
		// we do not validate body against schema
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
		log.Errorf("write HTTP body failed for %v: %v", r.URL.String(), err)
	}
	if logHttp != nil {
		logHttp.Response.Body = string(body)
		logHttp.Response.Size = len(body)
		logHttp.Actions = actions
	}
}

func (h *operationHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	regex := regexp.MustCompile(`\{(?P<name>.+)\}`) // parameter format "/{param}/"
	requestPath := r.URL.Path
	if len(requestPath) > 1 {
		requestPath = strings.TrimRight(requestPath, "/")
	}
	reqSeg := strings.Split(requestPath, "/")
	var lastError error

endpointLoop:
	for path, ref := range h.config.Paths {
		if ref.Value == nil {
			continue
		}
		endpoint := ref.Value
		op := endpoint.Operation(r.Method)
		if op == nil {
			continue
		}

		servicePath, ok := r.Context().Value("servicePath").(string)

		routePath := path
		if ok && servicePath != "/" {
			routePath = servicePath + routePath
		}
		if len(routePath) > 1 {
			routePath = strings.TrimRight(routePath, "/")
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

	if ctx, err := NewLogEventContext(r, false, events.NewTraits().WithName(h.config.Info.Name).With("path", r.URL.Path).With("method", r.Method)); err != nil {
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

	logMessage := fmt.Sprintf("HTTP request failed with status code %d: %s %s: %s", status, r.Method, r.URL.String(), err.Error())
	if status == http.StatusInternalServerError {
		log.Errorf(logMessage)
	} else {
		log.Infof(logMessage)
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

func (h *responseHandler) serveSecurity(r *http.Request, requirements []SecurityRequirement) error {
	var errs error
	for _, req := range requirements {
		var reqError error
		for name := range req {
			scheme, ok := h.config.Components.SecuritySchemes[name]
			if !ok {
				return newHttpError(http.StatusInternalServerError, fmt.Sprintf("no security scheme found for %v", name))
			}
			err := scheme.Serve(r)
			var notSupported *NotSupportedSecuritySchemeError
			if errors.As(err, &notSupported) {
				log.Warnf(notSupported.Error())
			} else if err != nil {
				reqError = errors.Join(reqError, fmt.Errorf("security scheme name '%s' type '%s': %w", name, getSecuritySchemeType(scheme), err))
			}
		}
		if reqError == nil {
			return nil
		}
		errs = errors.Join(errs, reqError)
	}
	if errs == nil {
		return nil
	}
	return newHttpError(http.StatusForbidden, errs.Error())
}
