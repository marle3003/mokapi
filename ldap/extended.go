package ldap

import (
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
)

const (
	extendedResponse = 24

	// noticeOfDisconnectionOID is the responseName of the unsolicited notification
	// a server sends before it closes the connection (RFC 4511 4.4.1).
	noticeOfDisconnectionOID = "1.3.6.1.4.1.1466.20036"
	// responseNameTag is the context-specific tag of ExtendedResponse.responseName.
	responseNameTag = 10
)

// noticeOfDisconnection is an unsolicited ExtendedResponse that tells the client
// why the server closes the connection.
type noticeOfDisconnection struct {
	ResultCode uint8
	Message    string
}

func (n *noticeOfDisconnection) toPacket() *ber.Packet {
	body := ber.Encode(ber.ClassApplication, ber.TypeConstructed, extendedResponse, nil, "Notice of Disconnection")
	body.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(n.ResultCode), "ResultCode"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "MatchedDN"))
	body.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, n.Message, "DiagnosticMessage"))
	body.AppendChild(ber.NewString(ber.ClassContext, ber.TypePrimitive, responseNameTag, noticeOfDisconnectionOID, "ResponseName"))
	return body
}
