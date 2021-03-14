package web

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/openapi"
	"net/http"
	"regexp"
	"strings"
)

type ServiceHandler struct {
	config *openapi.Config
}

func NewWebServiceHandler(config *openapi.Config) *ServiceHandler {
	return &ServiceHandler{config: config}
}

func (handler *ServiceHandler) ServeHTTP(ctx *HttpContext) {
	err := handler.resolveEndpoint(ctx)
	if err != nil {
		message := fmt.Sprintf("No endpoint found in service %v. Request %v %v",
			handler.config.Info.Name,
			ctx.Request.Method,
			ctx.Request.URL.String())
		http.Error(ctx.Response, message, http.StatusNotFound)
		log.Infof(message)
		return
	}

	if handler.config.Info.Mokapi != nil {
		ctx.Mokapi = handler.config.Info.Mokapi.Value
	}
	if err := ctx.Init(); err != nil {
		msg := err.Error()
		http.Error(ctx.Response, msg, http.StatusBadRequest)
		log.Infof(msg)
		return
	}

	operationHandler := NewOperationHandler()
	operationHandler.ProcessRequest(ctx)
}

func (handler *ServiceHandler) resolveEndpoint(ctx *HttpContext) error {
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
		if ctx.ServicPath != "/" {
			routePath = ctx.ServicPath + routePath
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
			continue
		}
		ctx.Parameters = p
		ctx.Operation = op
		return nil
	}

	return errors.Errorf("no matching endpoint found")
}

// Gets the operation for the given method name
func getOperation(method string, e *openapi.Endpoint) *openapi.Operation {
	switch strings.ToUpper(method) {
	case "GET":
		return e.Get
	case "POST":
		return e.Post
	case "Put":
		return e.Put
	case "Patch":
		return e.Patch
	case "Delete":
		return e.Delete
	case "Head":
		return e.Head
	case "Options":
		return e.Options
	case "Trace":
		return e.Trace
	}

	return nil
}
