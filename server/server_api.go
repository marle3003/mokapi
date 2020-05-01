package server

import (
	"context"
	"fmt"
	"mokapi/providers/data"
	"mokapi/server/handlers"
	"mokapi/service"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

type ServiceItem struct {
	service      *service.Service
	handler      *handlers.ServiceHandler
	dataProvider data.Provider
}

type ApiServer struct {
	router      *mux.Router
	entryPoints map[string]*handlers.EntryPointHandler
	servers     map[string]*HttpServer

	services map[string]*ServiceItem
}

type HttpServer struct {
	server *http.Server
	router *mux.Router
}

func NewHttpServer(address string) *HttpServer {
	router := mux.NewRouter()
	server := &http.Server{Addr: address, Handler: router}

	return &HttpServer{server: server, router: router}
}

func NewApiServer() *ApiServer {
	apiServer := &ApiServer{
		router:      mux.NewRouter(),
		entryPoints: make(map[string]*handlers.EntryPointHandler),
		servers:     make(map[string]*HttpServer),
		services:    make(map[string]*ServiceItem),
	}
	return apiServer
}

func (s *ApiServer) startServer(address string) {
	log.Infof("Starting server on %v", address)

	server := NewHttpServer(address)
	s.servers[address] = server

	go func() {
		server.server.ListenAndServe()
	}()
}

func (s *ApiServer) stopServer(server *http.Server) {
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

func (a *ApiServer) AddOrUpdate(s *service.Service) {
	var item *ServiceItem

	if old, ok := a.services[s.Name]; ok {
		for _, v := range old.service.Servers {
			serverAddress := fmt.Sprintf(":%v", v.Port)
			if entrypoint, ok := a.entryPoints[serverAddress]; ok {
				entrypoint.RemoveHandler(v.Path)
			}
		}

		old.dataProvider.Close()

		item = old
	} else {
		item = &ServiceItem{}
		a.services[s.Name] = item
	}

	item.service = s
	item.handler, item.dataProvider = build(s)

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

func build(service *service.Service) (*handlers.ServiceHandler, data.Provider) {
	serviceHandler := handlers.NewServiceHandler()

	dataProvider := getDataProvider(service)

	for p, endpoint := range service.Endpoint {
		endpointHandler := getEndPointHandler(endpoint, dataProvider)
		serviceHandler.AddHandler(p, endpointHandler)
	}

	return serviceHandler, dataProvider
}

func getDataProvider(service *service.Service) data.Provider {
	if service.DataProviders.File != nil {
		return data.NewStaticDataProvider(service.DataProviders.File.Path)
	}

	return data.NewRandomDataProvider()
}

func getEndPointHandler(endpoint *service.Endpoint, dataProvider data.Provider) *handlers.EndpointHandler {
	handler := handlers.NewEndpointHandler()

	if endpoint.Get != nil {
		log.WithFields(log.Fields{"method": "GET"}).Info("Adding operation handler")
		operationHandler := getOperationHandler(endpoint.Get, dataProvider)
		handler.AddHandler("GET", operationHandler)
	}

	return handler
}

func getOperationHandler(operation *service.Operation, dataProvider data.Provider) *handlers.OperationHandler {
	operationHandler := handlers.NewOperationHandler()

	// todo error handling
	response := selectSuccessResponse(operation)
	if response != nil {
		for contentType, content := range response.ContentTypes {
			log.WithFields(log.Fields{"contentType": contentType}).Info("Adding data handler")
			dataHandler := handlers.NewContentHandler(operation.Parameters, contentType, content.Schema, dataProvider)
			operationHandler.AddHandler(contentType, dataHandler)
		}
	}

	return operationHandler
}

func selectSuccessResponse(operation *service.Operation) *service.Response {
	keys := make([]service.HttpStatus, 0, len(operation.Responses))
	for k := range operation.Responses {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, key := range keys {
		if key >= 200 && key < 300 {
			return operation.Responses[key]
		}
	}

	return nil
}
