package ldaptest

import (
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
	"mokapi/server/ldap"
)

func NewSimpleBindRequest(messageId, version int64, name, password string) *ldap.Request {
	body := ber.Encode(
		ber.ClassApplication,
		ber.TypeConstructed,
		ldap.ApplicationBindRequest,
		nil,
		"Bind Request")

	body.AppendChild(ber.NewInteger(
		ber.ClassUniversal,
		ber.TypePrimitive,
		ber.TagInteger,
		version,
		"Version",
	))

	body.AppendChild(ber.NewString(
		ber.ClassUniversal,
		ber.TypePrimitive,
		ber.TagOctetString,
		name,
		"Username",
	))

	body.AppendChild(ber.NewString(
		ber.ClassContext,
		ber.TypePrimitive,
		0,
		password,
		"Password",
	))

	return &ldap.Request{
		MessageId: messageId,
		Body:      body,
	}
}

func NewUnbindRequest(messageId int64) *ldap.Request {
	body := ber.Encode(
		ber.ClassApplication,
		ber.TypeConstructed,
		ldap.ApplicationUnbindRequest,
		nil,
		"Bind Request")

	return &ldap.Request{
		MessageId: messageId,
		Body:      body,
	}
}

func NewAbandonRequest(messageId int64) *ldap.Request {
	body := ber.Encode(
		ber.ClassApplication,
		ber.TypeConstructed,
		ldap.ApplicationAbandonRequest,
		nil,
		"Abandon Request")

	return &ldap.Request{
		MessageId: messageId,
		Body:      body,
	}
}

func NewSearchRequest(messageId int64, r *ldap.SearchRequest) *ldap.Request {
	body := ber.Encode(
		ber.ClassUniversal,
		ber.TypeConstructed,
		ldap.ApplicationSearchRequest,
		nil,
		"Search Request")

	body.AppendChild(ber.NewString(
		ber.ClassUniversal,
		ber.TypePrimitive,
		ber.TagOctetString,
		r.BaseDN,
		"Base DN",
	))

	body.AppendChild(ber.NewInteger(
		ber.ClassUniversal,
		ber.TypePrimitive,
		ber.TagInteger,
		r.Scope,
		"Scope",
	))

	body.AppendChild(ber.NewInteger(
		ber.ClassContext,
		ber.TypePrimitive,
		ber.TagInteger,
		r.DereferencePolicy,
		"Dereference Policy",
	))

	body.AppendChild(ber.NewInteger(
		ber.ClassContext,
		ber.TypePrimitive,
		ber.TagInteger,
		r.SizeLimit,
		"SizeLimit",
	))

	body.AppendChild(ber.NewInteger(
		ber.ClassContext,
		ber.TypePrimitive,
		ber.TagInteger,
		r.TimeLimit,
		"TimeLimit",
	))

	body.AppendChild(ber.NewBoolean(
		ber.ClassContext,
		ber.TypePrimitive,
		ber.TagBoolean,
		r.TypesOnly,
		"TypesOnly",
	))

	body.AppendChild(r.Filter)

	attributes := ber.NewSequence("Attributes")
	for _, attr := range r.Attributes {
		attributes.AppendChild(
			ber.NewString(ber.ClassContext,
				ber.TypePrimitive,
				ber.TagOctetString,
				attr,
				""))
	}
	body.AppendChild(attributes)

	return &ldap.Request{
		MessageId: messageId,
		Body:      body,
	}
}
