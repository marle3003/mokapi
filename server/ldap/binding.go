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
	err := s.Apply(config)
	if err != nil {
		log.Errorf("unable to start ldap server: %v", err.Error())
	}
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

	schema, err := s.getSchema()
	if err != nil {
		log.Errorf("error in parsing schema: %v", err.Error())
	}
	s.schema = schema

	l, err := net.Listen("tcp", s.listen)
	if err != nil {
		log.Errorf("error listening: %v", err.Error())
		return
	}

	log.Infof("started ldap server on %v", s.listen)

	// Close the listener when the application closes.
	connChannl := make(chan net.Conn)
	c := make(chan bool)
	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-c:
					return
				default:
					log.Errorf("error accepting: %v", err.Error())
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
				log.Infof("stopping ldap server on %v", s.listen)
				c <- true
				err := l.Close()
				if err != nil {
					log.Infof("error while stopping ldap server on %v: %v", s.listen, err.Error())
				}
			}
		}
	}()
}

func (s *Binding) handle(conn net.Conn) {
	defer func() {
		log.Info("closing connection")
		err := conn.Close()
		if err != nil {
			log.Infof("error while closing ldap server on %v: %v", s.listen, err.Error())
		}
	}()

	for {
		packet, err := ber.ReadPacket(conn)
		if err == io.EOF { // Client closed connection
			return
		} else if err != nil {
			log.Errorf("handleConnection ber.ReadPacket ERROR: %v", err.Error())
			return
		}

		if len(packet.Children) < 2 {
			log.Errorf("invalid packat length %v expected at least 2", len(packet.Children))
			return
		}
		o := packet.Children[0].Value
		messageId, ok := packet.Children[0].Value.(int64)
		if !ok {
			log.Errorf("malformed messageId %v", reflect.TypeOf(o))
			return
		}
		req := packet.Children[1]
		if req.ClassType != ber.ClassApplication {
			log.Errorf("classType of packet is not ClassApplication was %v", req.ClassType)
			return
		}

		switch req.Tag {
		default:
			log.Errorf("unsupported tag %v", req.Tag)
		case ApplicationBindRequest:
			s.handleBindRequest(conn, messageId, req)
		case ApplicationUnbindRequest:
			log.Infof("received unbind request with messageId %v", messageId)
			// just close connection
			return
		case ApplicationSearchRequest:
			err := s.handleSearchRequest(conn, messageId, req)
			if err != nil {
				log.Errorf("error handling search request with messageId %v: %v", messageId, err.Error())
				return
			} else {
				sendResponse(conn, encodeSearchDone(messageId))
			}
		case ApplicationAbandonRequest:
			log.Infof("received abandon request with messageId %v", messageId)
			// todo stop any searches on this messageid
			// The abandon operation does not have a response
		}
	}
}

func (s *Binding) getEntry(dn string) (ldapConfig.Entry, error) {
	if entry, ok := s.entries[dn]; ok {
		return entry, nil
	}
	return ldapConfig.Entry{}, errors.Errorf("entry with dn %q not found", dn)
}

func sendResponse(conn net.Conn, packet *ber.Packet) {
	_, err := conn.Write(packet.Bytes())
	if err != nil {
		log.Errorf("error Sending Message: %s\n", err.Error())
	}
}
