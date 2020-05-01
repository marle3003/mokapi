package ldap

import (
	"io"
	"mokapi/config/dynamic"
	"net"
	"os"
	"reflect"

	log "github.com/sirupsen/logrus"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

type Server struct {
	stop      chan bool
	entries   []*dynamic.Entry
	listen    string
	isRunning bool
	root      *dynamic.Entry
}

func NewServer(config *dynamic.Ldap) *Server {
	s := &Server{stop: make(chan bool)}
	s.UpdateConfig(config)
	return s
}

func (s *Server) UpdateConfig(config *dynamic.Ldap) {
	shouldRestart := false
	if s.listen != "" && s.listen != config.Server.Listen {
		s.stop <- true
		shouldRestart = true
	}

	s.root = &dynamic.Entry{Dn: "", Attributes: make(map[string][]string)}
	for k, v := range config.Server.Config {
		s.root.Attributes[k] = v
	}
	s.root.Attributes["supportedLDAPVersion"] = []string{"3"}

	s.entries = config.Entries
	s.listen = config.Server.Listen

	if s.isRunning {
		log.Infof("Updated configuration of ldap server: %v", s.listen)

		if shouldRestart {
			go s.Start()
		}
	}
}

func (s *Server) Start() {
	s.isRunning = true

	l, err := net.Listen("tcp", s.listen)
	if err != nil {
		log.Errorf("Error listening: ", err.Error())
		os.Exit(1)
	}

	log.Infof("Started ldap server on %v", s.listen)

	// Close the listener when the application closes.
	connChannl := make(chan net.Conn)
	close := make(chan bool)
	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-close:
					return
				default:
					log.Errorf("Error accepting: ", err.Error())
				}
			}
			// Handle connections in a new goroutine.
			connChannl <- conn
		}
	}()

	for {
		select {
		case conn := <-connChannl:
			go s.handle(conn)
		case <-s.stop:
			log.Infof("Stopping ldap server on %v", s.listen)
			close <- true
			l.Close()
		}
	}
}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		log.Info("Closing connection")
		conn.Close()
	}()

	for {
		packet, err := ber.ReadPacket(conn)
		if err == io.EOF { // Client closed connection
			return
		} else if err != nil {
			log.Errorf("handleConnection ber.ReadPacket ERROR: %s", err.Error())
			return
		}

		if len(packet.Children) < 2 {
			log.Errorf("Invalid packat length %v expected at least 2", len(packet.Children))
			return
		}
		o := packet.Children[0].Value
		messageId, ok := packet.Children[0].Value.(int64)
		if !ok {
			log.Errorf("malformed messageId %v\n", reflect.TypeOf(o))
			return
		}
		req := packet.Children[1]
		if req.ClassType != ber.ClassApplication {
			log.Errorf("ClassType of packet is not ClassApplication was %v", req.ClassType)
			return
		}

		switch req.Tag {
		default:
			log.Errorf("Unsupported tag %v", req.Tag)
		case ApplicationBindRequest:
			s.handleBindRequest(conn, messageId, req)
		case ApplicationUnbindRequest:
			log.Infof("Received unbind request with messageId %v", messageId)
			// just close connection
			return
		case ApplicationSearchRequest:
			error := s.handleSearchRequest(conn, messageId, req)
			if error != nil {
				log.Errorf("Error handling search request with messageId %v: %s", messageId, error.Error())
				return
			} else {
				sendResponse(conn, encodeSearchDone(messageId))
			}
		case ApplicationAbandonRequest:
			log.Infof("Received abandon request with messageId %v", messageId)
			// todo stop any searches on this messageid
			// The abandon operation does not have a response
		}
	}
}

func sendResponse(conn net.Conn, packet *ber.Packet) {
	_, err := conn.Write(packet.Bytes())
	if err != nil {
		log.Errorf("Error Sending Message: %s\n", err.Error())
	}
}
