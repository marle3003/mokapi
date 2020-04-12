package server

type Server struct {
	apiServer *ApiServer
}

func NewServer(apiServer *ApiServer) *Server {

	server := &Server{apiServer: apiServer}
	return server
}

func (server *Server) Start() {
	server.apiServer.Start()
}
