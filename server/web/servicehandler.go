package web

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/config/dynamic/openapi"
	"net/http"
	"regexp"
	"strings"
)

type serviceHandler struct {
	config *openapi.Config
}

func newServiceHandler(config *openapi.Config) *serviceHandler {
	return &serviceHandler{config: config}
}

func (handler *serviceHandler) ServeHTTP(ctx *HttpContext) {
	ctx.metric.Service = handler.config.Info.Name
	err := handler.resolveEndpoint(ctx)
	if err != nil {
		message := fmt.Sprintf("unable to serve http request of API %v: %v",
			handler.config.Info.Name,
			err.Error())
		if hErr, ok := err.(*httpError); ok {
			writeError(message, hErr.StatusCode, ctx)
		} else {
			writeError(message, http.StatusNotFound, ctx)
		}
		return
	}

	if err := ctx.setResponse(); err != nil {
		message := err.Error()
		if hErr, ok := err.(*httpError); ok {
			writeError(message, hErr.StatusCode, ctx)
		} else {
			writeError(message, http.StatusInternalServerError, ctx)
		}
		return
	}

	operationHandler := NewOperationHandler()
	operationHandler.ProcessRequest(ctx)
}

func (handler *serviceHandler) resolveEndpoint(ctx *HttpContext) (lastError error) {
	regex := regexp.MustCompile(`\{(?P<name>.+)\}`) // parameter format "/{param}/"
	reqSeg := strings.Split(ctx.Request.URL.Path, "/")

endpointLoop:
	for path, ref := range handler.config.EndPoints {
		if ref.Value == nil {
			continue
		}
		endpoint := ref.Value
		op := getOperation(ctx.Request.Method, endpoint)
		if op == nil {
			continue
		}

		routePath := path
		if ctx.ServicePath != "/" {
			routePath = ctx.ServicePath + routePath
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
		p, err := parseParams(params, routePath, ctx.Request)
		if err != nil {
			lastError = newHttpError(400, err.Error())
			continue
		}

		ctx.ServiceName = handler.config.Info.Name
		ctx.Parameters = p
		ctx.Operation = op
		ctx.EndpointPath = path
		return nil
	}

	if lastError == nil {
		lastError = errors.Errorf("no matching endpoint found")
	}
	return
}

// Gets the operation for the given method name
func getOperation(method string, e *openapi.Endpoint) *openapi.Operation {
	switch strings.ToUpper(method) {
	case "GET":
		return e.Get
	case "POST":
		return e.Post
	case "PUT":
		return e.Put
	case "PATCH":
		return e.Patch
	case "DELETE":
		return e.Delete
	case "HEAD":
		return e.Head
	case "OPTIONS":
		return e.Options
	case "TRACE":
		return e.Trace
	}

	return nil
}
