package imap

import (
	"fmt"
	"strings"
	"time"
)

type FetchResponse interface {
	NewMessage(sequenceNumber uint32) MessageWriter
}

type MessageWriter interface {
	WriteUID(uid uint32)
	WriteInternalDate(date time.Time)
	WriteRFC822Size(size int64)
	WriteEnvelope(env *Envelope)
}

type Envelope struct {
	Date      time.Time
	Subject   string
	From      []Address
	Sender    []Address
	ReplyTo   []Address
	To        []Address
	Cc        []Address
	Bcc       []Address
	InReplyTo string
	MessageId string
}

type Header struct {
	Date        time.Time
	Subject     string
	From        []Address
	ReplyTo     []Address
	To          []Address
	Cc          []Address
	Bcc         []Address
	InReplyTo   string
	MessageId   string
	ContentType string
}

type Address struct {
	Name    string
	Mailbox string
	Host    string
}

type BodyStructure struct {
}

type fetchResponse struct {
	messages []*message
}

type message struct {
	sequenceNumber uint32
	sb             strings.Builder
}

func (r *fetchResponse) NewMessage(sequenceNumber uint32) MessageWriter {
	w := &message{sequenceNumber: sequenceNumber}
	r.messages = append(r.messages, w)
	return w
}

func (m *message) WriteUID(uid uint32) {
	m.sb.WriteString(fmt.Sprintf(" UID %v", uid))
}

func (m *message) WriteInternalDate(date time.Time) {
	m.sb.WriteString(fmt.Sprintf(" INTERNALDATE \"%v\"", date.Format(DateTimeLayout)))
}

func (m *message) WriteRFC822Size(size int64) {
	m.sb.WriteString(fmt.Sprintf(" RFC822.SIZE %v", size))
}

func (m *message) WriteEnvelope(env *Envelope) {
	m.sb.WriteString(" ENVELOPE (")
	m.sb.WriteString(fmt.Sprintf("\"%v\"", env.Date.Format(DateTimeLayout)))
	m.sb.WriteString(fmt.Sprintf(" \"%v\"", env.Subject))
	m.writeAddress(env.From)
	/*
		If the Sender or Reply-To lines are absent in the [RFC-2822]
		header, or are present but empty, the server sets the
		corresponding member of the envelope to be the same value as
		the from member
	*/
	if len(env.Sender) == 0 {
		env.Sender = env.From
	}
	m.writeAddress(env.Sender)
	if len(env.ReplyTo) == 0 {
		env.ReplyTo = env.From
	}
	m.writeAddress(env.ReplyTo)
	m.writeAddress(env.To)
	m.writeAddress(env.Cc)
	m.writeAddress(env.Bcc)
	if len(env.InReplyTo) == 0 {
		m.sb.WriteString(" NIL")
	} else {
		m.sb.WriteString(fmt.Sprintf(" \"%v\"", env.InReplyTo))
	}
	m.sb.WriteString(fmt.Sprintf(" \"%v\"", env.MessageId))
	m.sb.WriteString(")")
}

func (m *message) WriteBody(body *FetchBody, values []string) {
	m.sb.WriteString(" BODY[HEADER.FIELDS (")
	var sb strings.Builder
	for i, field := range body.HeaderFields {
		if i > 0 {
			m.sb.WriteString(" ")
		}
		m.sb.WriteString(field)

		sb.WriteString(fmt.Sprintf("%s=%s\r\n", field, values[i]))
	}
	m.sb.WriteString(")]")
	m.sb.WriteString(fmt.Sprintf(" {%v}\r\n", sb.Len()))
	m.sb.WriteString(sb.String())
}

func (m *message) writeAddress(addrList []Address) {
	if len(addrList) == 0 {
		m.sb.WriteString(" NIL")
		return
	}
	m.sb.WriteString(" (")
	for _, addr := range addrList {
		s := fmt.Sprintf("(\"%v\" NIL \"%v\" \"%v\")", addr.Name, addr.Mailbox, addr.Host)
		m.sb.WriteString(s)
	}
	m.sb.WriteString(")")
}
