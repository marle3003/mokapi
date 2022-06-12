package ldap

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
	"mokapi/runtime/monitor"
)

func (d *Directory) bind(rw ResponseWriter, r *Request) error {
	ldapVersion := r.Body.Children[0].Value.(int64)
	// we do only support ldap v3
	if ldapVersion != 3 {
		log.Debugf("unsupported LDAP version: %d", ldapVersion)
		return rw.Write(newBindResponse(r.MessageId, ProtocolError, "server supports only ldap version 3"))
	}

	// auth types
	name, ok := r.Body.Children[1].Value.(string)
	if !ok {
		return fmt.Errorf("unable to parse bindDN: expected string")
	}

	switch r.Body.Children[2].Tag {
	case 0: // simple
		log.Debugf("received bind request with messageId %v, version %v. auth: %v", r.MessageId, ldapVersion, name)
		if m, ok := monitor.LdapFromContext(r.Context); ok {
			m.Bind.WithLabel(d.config.Info.Name).Add(1)
		}
		return rw.Write(newBindResponse(r.MessageId, ResultSuccess, ""))
	default:
		return rw.Write(newBindResponse(r.MessageId, AuthMethodNotSupported, "server supports only simple auth method"))
	}

}

func newBindResponse(messageID int64, result int64, msg string) *ber.Packet {
	responsePacket := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	responsePacket.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageID, "Message ID"))

	bindResponse := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ApplicationBindResponse, nil, "Bind Response")
	bindResponse.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, result, "resultCode: "))
	bindResponse.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "matchedDN: "))
	bindResponse.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, msg, "errorMessage: "))

	responsePacket.AppendChild(bindResponse)

	return responsePacket
}
