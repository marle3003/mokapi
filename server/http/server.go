package http

import (
	"context"
	"fmt"
	"mokapi/config/static"
	"mokapi/models"
	"mokapi/providers/data"
	"mokapi/server/api"
	"mokapi/server/http/handlers"
	"net/http"
	h "net/http"
	"time"

	assetfs "github.com/elazarl/go-bindata-assetfs"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

// TODO Refactoring Classes

type ServiceItem struct {
	service *models.Service
	handler *handlers.ServiceHandler
}

type Server struct {
	entryPoints map[string]*handlers.EntryPointHandler
	servers     map[string]*HttpServer

	services map[string]*ServiceItem
	api      *api.Handler

	requestChannel chan *models.RequestMetric
}

type HttpServer struct {
	server *h.Server
	router *mux.Router
}

func newHttpServer(address string) *HttpServer {
	router := mux.NewRouter()
	server := &h.Server{Addr: address, Handler: router}

	return &HttpServer{server: server, router: router}
}

func NewServer(api *api.Handler, apiConfig static.Api, requestChannel chan *models.RequestMetric) *Server {
	apiServer := &Server{
		entryPoints:    make(map[string]*handlers.EntryPointHandler),
		servers:        make(map[string]*HttpServer),
		services:       make(map[string]*ServiceItem),
		requestChannel: requestChannel,
	}

	apiServer.startApi(api, apiConfig)

	return apiServer
}

func (s *Server) startApi(api *api.Handler, config static.Api) {
	server := s.startServer(":" + config.Port)
	api.CreateRoutes(server.router)

	if config.Dashboard {
		server.router.PathPrefix("/").Handler(http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir}))
	}
}

func (s *Server) startServer(address string) *HttpServer {
	log.Infof("Starting server on %v", address)

	server := newHttpServer(address)
	s.servers[address] = server

	go func() {
		server.server.ListenAndServe()
	}()

	return server
}

func (s *Server) stopServer(server *h.Server) {
	go func() {
		log.Infof("Stopping server on %v", server)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if error := server.Shutdown(ctx); error != nil {
			log.Errorf("Could not gracefully shutdown server %v", server)
		}
	}()
}

func (s *Server) Stop() {
	for _, server := range s.servers {
		s.stopServer(server.server)
	}
}

func (a *Server) AddOrUpdate(s *models.Service) {
	var item *ServiceItem

	if old, ok := a.services[s.Name]; ok {
		for _, v := range old.service.Servers {
			if entrypoint, ok := a.entryPoints[v.Host]; ok {
				entrypoint.RemoveHandler(v.Path)
			}
		}

		old.handler.Close()

		item = old
	} else {
		item = &ServiceItem{}
		a.services[s.Name] = item
	}

	item.service = s
	item.handler = build(s, a.requestChannel)

	for _, v := range s.Servers {
		serverAddress := fmt.Sprintf(":%v", v.Port)
		if entrypoint, ok := a.entryPoints[v.Host]; ok {
			log.Infof("Adding service %v at %v on %v", item.service.Name, v.Path, serverAddress)
			entrypoint.AddHandler(v.Path, item.handler)
		} else {
			entrypoint = handlers.NewEntryPointHandler(v.Host, v.Port, a.requestChannel)

			if _, ok := a.servers[serverAddress]; !ok {
				a.startServer(serverAddress)
			}

			a.servers[serverAddress].router.Host(v.Host).Subrouter().NewRoute().Handler(entrypoint)

			log.Infof("Adding service %v at %v on %v", item.service.Name, v.Path, serverAddress)
			entrypoint.AddHandler(v.Path, item.handler)

			a.entryPoints[v.Host] = entrypoint
		}
	}
}

func build(service *models.Service, requestChannel chan *models.RequestMetric) *handlers.ServiceHandler {
	dataProvider := getDataProvider(service)
	serviceHandler := handlers.NewServiceHandler(service, dataProvider, requestChannel)

	return serviceHandler
}

func getDataProvider(service *models.Service) data.Provider {
	if service.DataProviders.File != nil {
		return data.NewStaticDataProvider(service.DataProviders.File.Path, true)
	}

	return data.NewRandomDataProvider()
}
