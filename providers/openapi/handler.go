package openapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"mokapi/engine/common"
	"mokapi/lib"
	"mokapi/media"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) *HttpError
}

type operationHandler struct {
	config *Config
	next   http.Handler
	eh     events.Handler
}

type responseHandler struct {
	config       *Config
	eventEmitter common.EventEmitter
}

func NewHandler(config *Config, eventEmitter common.EventEmitter, eh events.Handler) Handler {
	return &operationHandler{
		config: config,
		next: &responseHandler{
			config:       config,
			eventEmitter: eventEmitter,
		},
		eh: eh,
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
		writeError(
			rw,
			r,
			fmt.Errorf("resolve success response failed for '%s %s': %w", r.Method, op.Path.Path, err),
			h.config.Info.Name,
		)
		return
	} else if res == nil {
		writeError(rw, r, fmt.Errorf("response not defined for HTTP status %v", status), h.config.Info.Name)
		return
	}

	contentType, mediaType, err := ContentTypeFromRequest(r, res)
	if err != nil {
		writeError(rw, r, err, h.config.Info.Name)
		return
	}

	request, ctx := NewEventRequest(r, contentType, h.config.Info.Name)
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
	} else {
		// Not reading the request body can cause a couple of problems.
		// Go’s HTTP server uses connection pooling by default.
		// If you don’t read and fully consume (or close) the request body, the remaining unread bytes will stay in the TCP buffer.
		drainRequestBody(r)
	}

	if len(op.Security) > 0 {
		h.serveSecurity(r, op.Security)
	} else if len(h.config.Security) > 0 {
		h.serveSecurity(r, h.config.Security)
	}

	response := NewEventResponse(status, contentType)
	setResponseRebuild(response, request, op)

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
	if logHttp != nil {
		logHttp.Actions = actions
	}

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
	err = setHeaders(response.Headers, res.Headers, rw)
	if err != nil {
		writeError(rw, r, err, h.config.Info.Name)
		return
	}

	if len(res.Content) == 0 {
		if response.Body == "" {
			// no response content and no response body is defined which means body is empty
			rw.Header().Del("Content-Type")
			rw.Header().Set("Content-Length", "0")
		}
		if response.StatusCode > 0 {
			rw.WriteHeader(response.StatusCode)
			if logHttp != nil {
				logHttp.Response.StatusCode = response.StatusCode
			}
		}
		if response.Body != "" {
			// Write body even specification defines no content.
			// Useful for development, UI testing, or when working with incomplete specs.
			_, err = rw.Write([]byte(response.Body))
			if err != nil {
				log.Errorf("write HTTP body failed for %v: %v", r.URL.String(), err)
			}
			if logHttp != nil {
				logHttp.Response.Body = response.Body
				logHttp.Response.Size = len(response.Body)
			}
		}
		return
	}

	if ct, err := getContentType(response.Headers); ct != "" && err == nil {
		contentType = media.ParseContentType(ct)
	} else if err != nil {
		writeError(rw, r, err, h.config.Info.Name)
		return
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
	}
}

func (h *operationHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) *HttpError {
	requestPath := r.URL.Path
	if len(requestPath) > 1 {
		requestPath = strings.TrimRight(requestPath, "/")
	}

	servicePath, ok := r.Context().Value("servicePath").(string)
	if ok && servicePath != "/" {
		servicePath = strings.TrimRight(servicePath, "/")
		requestPath = strings.Replace(requestPath, servicePath, "", 1)
		if requestPath == "" {
			requestPath = "/"
		}
	}
	op, errOpResolve := findOperation(r.Method, requestPath, h.config.Paths)
	if op != nil {
		params := append(op.Path.Parameters, op.Parameters...)
		route := op.Path.Path
		if len(route) > 1 {
			// there is no official specification for trailing slash. For ease of use, mokapi considers it equivalent
			route = strings.TrimRight(route, "/")
		}

		var rp *RequestParameters
		rp, errOpResolve = FromRequest(params, route, r)

		if errOpResolve == nil {
			r = r.WithContext(NewContext(r.Context(), rp))
			r = r.WithContext(NewOperationContext(r.Context(), op))
			r = r.WithContext(context.WithValue(r.Context(), "endpointPath", op.Path.Path))

			if m, ok := monitor.HttpFromContext(r.Context()); ok {
				m.LastRequest.WithLabel(h.config.Info.Name, op.Path.Path, r.Method).Set(float64(time.Now().Unix()))
				m.RequestCounter.WithLabel(h.config.Info.Name, op.Path.Path, r.Method).Add(1)
			}

			if ctx, err := NewLogEventContext(
				r,
				op.Deprecated,
				h.eh,
				events.NewTraits().WithName(h.config.Info.Name).With("path", op.Path.Path).With("method", r.Method),
			); err != nil {
				log.Errorf("unable to log http event: %v", err)
			} else {
				r = r.WithContext(ctx)
			}

			h.next.ServeHTTP(rw, r)
			return nil
		} else {
			errOpResolve = newHttpError(http.StatusBadRequest, errOpResolve.Error())
		}
	}

	/*if ctx, err := NewLogEventContext(
		r,
		false,
		h.eh,
		events.NewTraits().WithName(h.config.Info.Name).With("path", r.URL.Path).With("method", r.Method),
	); err != nil {
		log.Errorf("unable to log http event: %v", err)
	} else {
		r = r.WithContext(ctx)
	}*/

	// Not reading the request body can cause a couple of problems.
	// Go’s HTTP server uses connection pooling by default.
	// If you don’t read and fully consume (or close) the request body, the remaining unread bytes will stay in the TCP buffer.
	//_, _ = io.Copy(io.Discard, r.Body)

	if errOpResolve == nil {
		return newHttpErrorf(http.StatusNotFound, "no matching endpoint found: %v %v", strings.ToUpper(r.Method), lib.GetUrl(r))
	} else {
		var httpError *HttpError
		if errors.As(errOpResolve, &httpError) {
			return httpError
		}
		return newHttpErrorf(http.StatusInternalServerError, "%s", errOpResolve.Error())
	}
}

func writeError(rw http.ResponseWriter, r *http.Request, err error, serviceName string) {
	message := err.Error()
	status := http.StatusInternalServerError

	var hErr *HttpError
	if errors.As(err, &hErr) {
		status = hErr.StatusCode
		for k, values := range hErr.Header {
			for _, v := range values {
				rw.Header().Add(k, v)
			}
		}
	}

	logMessage := fmt.Sprintf("HTTP request failed with status code %d: %s %s: %s", status, r.Method, r.URL.String(), err.Error())
	if status == http.StatusInternalServerError {
		log.Error(logMessage)
	} else {
		log.Info(logMessage)
	}
	if m, ok := monitor.HttpFromContext(r.Context()); ok {
		endpointPath := r.Context().Value("endpointPath")
		if endpointPath != nil {
			endpointPathString := endpointPath.(string)
			m.RequestErrorCounter.WithLabel(serviceName, endpointPathString, r.Method).Add(1)
			m.LastRequest.WithLabel(serviceName, endpointPathString, r.Method).Set(float64(time.Now().Unix()))
		} else {
			m.RequestErrorCounter.WithLabel(serviceName, "", r.Method).Add(1)
			m.LastRequest.WithLabel(serviceName, "", r.Method).Set(float64(time.Now().Unix()))
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

func (h *responseHandler) serveSecurity(r *http.Request, requirements []SecurityRequirement) {
	var errs error
	for _, req := range requirements {
		var reqError error
		for name := range req {
			scheme, ok := h.config.Components.SecuritySchemes[name]
			if !ok {
				log.Warnf("%s: no security scheme found for %v", r.URL.String(), name)
			}
			err := scheme.Serve(r)
			var notSupported *NotSupportedSecuritySchemeError
			if errors.As(err, &notSupported) {
				log.Warn(notSupported.Error())
			} else if err != nil {
				reqError = errors.Join(reqError, fmt.Errorf("security scheme name '%s' type '%s': %w", name, getSecuritySchemeType(scheme), err))
			}
		}
		if reqError == nil {
			return
		}
		errs = errors.Join(errs, reqError)
	}
	if errs != nil {
		log.Infof("%s: security requirement skipped: %v", r.URL.String(), errs.Error())
	}
}

func findOperation(method, requestPath string, paths PathItems) (*Operation, error) {
	var lastError error
	var selectedOperation *Operation
	var numParams int

	for route, ref := range paths {
		if ref.Value == nil {
			continue
		}

		origRoute := route
		if len(route) > 1 {
			// there is no official specification for trailing slash. For ease of use, mokapi considers it equivalent
			route = strings.TrimRight(route, "/")
		}

		params, err := extractPathParams(route, requestPath)
		if err != nil {
			continue
		}
		op := ref.Value.Operation(method)
		if op == nil {
			allowed := slices.Collect(maps.Keys(ref.Value.Operations()))
			lastError = newMethodNotAllowedErrorf(allowed, "Method %s not defined for path %s", method, origRoute)
		}

		// Literal/static paths have higher priority than parameterized paths.
		if selectedOperation == nil || numParams > len(params) {
			selectedOperation = op
			numParams = len(params)
		}
	}

	return selectedOperation, lastError
}

func extractPathParams(route, path string) (map[string]string, error) {
	// Find all {param} names
	re := regexp.MustCompile(`\{([^}]+)\}`)
	names := re.FindAllStringSubmatch(route, -1)

	// Replace {param} with regex group
	pattern := "^" + re.ReplaceAllString(route, `([^/]+)`) + "$"
	re = regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(path)
	if match == nil {
		// path parameters are always required
		return nil, fmt.Errorf("url does not match route")
	}

	// Build result map
	params := map[string]string{}
	for i, name := range names {
		params[name[1]] = match[i+1]
	}

	return params, nil
}

func drainRequestBody(r *http.Request) {
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, r.Body)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		log.Warnf("timeout reading request body for %s %s", r.Method, lib.GetUrl(r))
	}
}

func setResponseRebuild(response *common.HttpEventResponse, request *common.HttpEventRequest, op *Operation) {
	response.Rebuild = func(statusCode int, contentType string) {
		res := op.Responses.GetResponse(statusCode)
		if res == nil {
			res = op.getResponse(0)
			if res == nil {
				panic(
					fmt.Sprintf(
						"no configuration was found for HTTP status code %v, https://swagger.io/docs/specification/describing-responses",
						statusCode,
					),
				)
			}
		}
		var mediaType *MediaType
		if contentType != "" {
			mediaType = res.Content[contentType]
			if mediaType == nil {
				panic(fmt.Sprintf("content type '%s' is not specified for HTTP status code %v", contentType, statusCode))
			}
		} else {
			for name, mt := range res.Content {
				contentType = name
				mediaType = mt
			}
		}

		response.StatusCode = statusCode
		response.Headers = map[string]any{}

		err := setResponseData(response, mediaType, request)
		if err != nil {
			panic(err)
		}

		err = setResponseHeader(response, res.Headers)
		if err != nil {
			panic(err)
		}
	}
}
