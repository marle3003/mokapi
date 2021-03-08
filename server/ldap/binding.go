package ldap

import (
	"github.com/pkg/errors"
	"io"
	ldapConfig "mokapi/config/dynamic/ldap"
	"net"
	"reflect"

	log "github.com/sirupsen/logrus"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

type Binding struct {
	stop      chan bool
	entries   map[string]ldapConfig.Entry
	listen    string
	isRunning bool
	root      ldapConfig.Entry
	schema    *Schema
}

func NewServer(config *ldapConfig.Config) *Binding {
	s := &Binding{stop: make(chan bool)}
	s.Apply(config)
	return s
}

func (s *Binding) Apply(data interface{}) error {
	config, _ := data.(*ldapConfig.Config)
	shouldRestart := false
	if s.listen != "" && s.listen != config.Address {
		s.stop <- true
		shouldRestart = true
	}

	s.root = config.Root
	s.root.Attributes["supportedLDAPVersion"] = []string{"3"}

	s.entries = config.Entries
	s.listen = config.Address

	if s.isRunning {
		log.Infof("Updated configuration of ldap server: %v", s.listen)

		if shouldRestart {
			go s.Start()
		}
	}
	return nil
}

func (s *Binding) Stop() {
	s.stop <- true
}

func (s *Binding) Start() {
	s.isRunning = true

	schema, error := s.getSchema()
	if error != nil {
		log.Errorf("Error in parsing schema: %v", error.Error())
	}
	s.schema = schema

	l, err := net.Listen("tcp", s.listen)
	if err != nil {
		log.Errorf("Error listening: %v", err.Error())
		return
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
					log.Errorf("Error accepting: %v", err.Error())
				}
			}
			// Handle connections in a new goroutine.
			connChannl <- conn
		}
	}()

	go func() {
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
	}()
}

func (s *Binding) handle(conn net.Conn) {
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

func (s *Binding) getEntry(dn string) (ldapConfig.Entry, error) {
	if entry, ok := s.entries[dn]; ok {
		return entry, nil
	}
	return ldapConfig.Entry{}, errors.Errorf("entry with dn %v not found", dn)
}

func sendResponse(conn net.Conn, packet *ber.Packet) {
	_, err := conn.Write(packet.Bytes())
	if err != nil {
		log.Errorf("Error Sending Message: %s\n", err.Error())
	}
}
