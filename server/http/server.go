package http

import (
	"context"
	"fmt"
	"mokapi/models"
	"mokapi/providers/data"
	"mokapi/server/api"
	"mokapi/server/http/handlers"
	h "net/http"
	"time"

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

func NewServer(api *api.Handler) *Server {
	apiServer := &Server{
		entryPoints: make(map[string]*handlers.EntryPointHandler),
		servers:     make(map[string]*HttpServer),
		services:    make(map[string]*ServiceItem),
	}

	apiServer.startApi(api)

	return apiServer
}

func (s *Server) startApi(api *api.Handler) {
	apiRoute := api.CreateRouter()

	apiServer := &HttpServer{router: apiRoute, server: &h.Server{Addr: ":8081", Handler: apiRoute}}
	s.servers[":8081"] = apiServer

	go func() {
		apiServer.server.ListenAndServe()
	}()
}

func (s *Server) startServer(address string) {
	log.Infof("Starting server on %v", address)

	server := newHttpServer(address)
	s.servers[address] = server

	go func() {
		server.server.ListenAndServe()
	}()
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
	item.handler = build(s)

	for _, v := range s.Servers {
		serverAddress := fmt.Sprintf(":%v", v.Port)
		if entrypoint, ok := a.entryPoints[v.Host]; ok {
			log.Infof("Adding service %v at %v on %v", item.service.Name, v.Path, serverAddress)
			entrypoint.AddHandler(v.Path, item.handler)
		} else {
			entrypoint = handlers.NewEntryPointHandler(v.Host, v.Port)

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

func build(service *models.Service) *handlers.ServiceHandler {
	dataProvider := getDataProvider(service)
	serviceHandler := handlers.NewServiceHandler(service, dataProvider)

	return serviceHandler
}

func getDataProvider(service *models.Service) data.Provider {
	if service.DataProviders.File != nil {
		return data.NewStaticDataProvider(service.DataProviders.File.Path, true)
	}

	return data.NewRandomDataProvider()
}
