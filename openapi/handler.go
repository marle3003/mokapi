package openapi

import (
	"mokapi/config/dynamic/openapi"
	"net/http"
)

type handler struct {
	config *openapi.Config
}

func New(config *openapi.Config) (http.Handler, error) {
	return &handler{config: config}, nil
}

func (o *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

}
