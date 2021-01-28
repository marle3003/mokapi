package web

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/models"
	"net/http"
	"regexp"
	"strings"
)

type ServiceHandler struct {
	WebService *models.WebService
}

func NewWebServiceHandler(service *models.WebService) *ServiceHandler {
	return &ServiceHandler{WebService: service}
}

func (handler *ServiceHandler) ServeHTTP(ctx *HttpContext) {
	err := handler.resolveEndpoint(ctx)
	if err != nil {
		message := fmt.Sprintf("No endpoint found in service %v. Request %v %v",
			handler.WebService.Name,
			ctx.Request.Method,
			ctx.Request.URL.String())
		http.Error(ctx.Response, message, http.StatusNotFound)
		log.Infof(message)
		return
	}

	ctx.MokapiFile = handler.WebService.MokapiFile
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
	for _, endpoint := range handler.WebService.Endpoint {
		op := endpoint.GetOperation(ctx.Request.Method)
		if op == nil {
			continue
		}

		routePath := endpoint.Path
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
