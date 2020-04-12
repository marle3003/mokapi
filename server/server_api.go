package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type ApiServer struct {
	router *mux.Router
}

func NewApiServer() *ApiServer {
	apiServer := &ApiServer{router: mux.NewRouter()}
	return apiServer
}

func (server *ApiServer) Start() {
	http.ListenAndServe(":8001", server.router)
}

func (server *ApiServer) SetRouters(entryPoints map[string]*EntryPoint) {
	for _, entryPoint := range entryPoints {
		server.router.Host(entryPoint.host).Subrouter().NewRoute().Handler(entryPoint.handler)
	}
}
