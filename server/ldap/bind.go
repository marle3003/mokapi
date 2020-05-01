package ldap

import (
	"net"

	log "github.com/sirupsen/logrus"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

func (s *Server) handleBindRequest(conn net.Conn, messageId int64, req *ber.Packet) {
	ldapVersion, ok := req.Children[0].Value.(int64)
	if !ok {
		log.Errorf("Error handling ldapVersion")
	}

	// we only support ldapv3
	if ldapVersion != 3 {
		log.Errorf("Unsupported LDAP version: %d", ldapVersion)
	}

	// auth types
	bindDN, ok := req.Children[1].Value.(string)
	if !ok {
		log.Errorf("Error handling bindDN\n")
	}
	log.Infof("Received bind request with messageId %v, version %v. auth: %v", messageId, ldapVersion, bindDN)

	sendResponse(conn, encodeBindResponse(messageId))
}
