package imap

import (
	"fmt"
	"mokapi/media"
	"strings"
	"time"
)

type FetchResponse interface {
	NewMessage(sequenceNumber uint32) MessageWriter
}

type MessageWriter interface {
	WriteUID(uid uint32)
	WriteInternalDate(date time.Time)
	WriteRFC822Size(size uint32)
	WriteFlags(flags ...Flag)
	WriteEnvelope(env *Envelope)
	WriteBody(section FetchBodySection) BodyWriter
	WriteBodyStructure(body *BodyStructure)
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
	Type        string
	Subtype     string
	Params      map[string]string
	Encoding    string
	Size        uint32
	MD5         *string
	Disposition string
	ContentId   *string
	Language    *string
	Location    *string
	Parts       []BodyStructure
}

type FetchData struct {
	Def  FetchBodySection
	Data string
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
	m.sb.WriteString(fmt.Sprintf(" INTERNALDATE "))
	m.sb.WriteByte('"')
	m.sb.WriteString(date.Format(DateTimeLayout))
	m.sb.WriteByte('"')
}

func (m *message) WriteRFC822Size(size uint32) {
	m.sb.WriteString(fmt.Sprintf(" RFC822.SIZE %v", size))
}

func (m *message) WriteFlags(flags ...Flag) {
	m.sb.WriteString(fmt.Sprintf(" FLAGS (%v)", joinFlags(flags)))
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

	if len(env.InReplyTo) == 0 {
		m.sb.WriteString(" NIL")
	} else {
		m.sb.WriteString(fmt.Sprintf(" \"%v\"", env.InReplyTo))
	}

	m.sb.WriteString(")")
}

func (m *message) WriteBody(section FetchBodySection) BodyWriter {
	section.Peek = false
	w := &body{section: section, m: m}
	return w
}

type BodyWriter interface {
	WriteHeader(name, value string)
	WriteBody(s string)
	Close()
}

type body struct {
	section FetchBodySection
	header  strings.Builder
	body    strings.Builder
	m       *message
}

func (w *body) WriteHeader(name, value string) {
	w.header.WriteString(fmt.Sprintf("%s: %s\r\n", name, value))
}

func (w *body) WriteBody(s string) {
	w.body.WriteString(s)
	w.body.WriteString("\r\n\r\n")
}

func (w *body) Close() {
	w.m.sb.WriteString(" " + w.section.encode())
	w.m.sb.WriteString(fmt.Sprintf(" {%v}\r\n", w.header.Len()+w.body.Len()+2))

	// header must end with a blank line
	data := w.header.String() + "\r\n" + w.body.String()
	if w.section.Partially != nil {
		offset := w.section.Partially.Offset
		if offset > uint32(len(data)) {
			return
		}
		limit := w.section.Partially.Limit
		end := min(offset+limit, uint32(len(data)))
		data = data[offset:end]
	}

	w.m.sb.WriteString(data)
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

func (m *message) WriteBodyStructure(b *BodyStructure) {
	m.sb.WriteString(" BODYSTRUCTURE ")

	writeStructure(m, b)
}

func writeStructure(m *message, b *BodyStructure) {
	// Multipart
	if strings.ToLower(b.Type) == "multipart" {
		m.sb.WriteString("(")
		for _, part := range b.Parts {
			writeStructure(m, &part)
		}
		m.sb.WriteString(fmt.Sprintf("\"%s\") ", b.Subtype))
		return
	}

	// Single part
	m.sb.WriteString("(")
	m.sb.WriteString(fmt.Sprintf("\"%v\" ", b.Type))
	m.sb.WriteString(fmt.Sprintf("\"%v\" ", b.Subtype))

	// Params
	var params []string
	for k, v := range b.Params {
		params = append(params,
			fmt.Sprintf("\"%v\"", k),
			fmt.Sprintf("\"%v\"", v),
		)
	}
	if len(params) > 0 {
		m.sb.WriteString("(" + strings.Join(params, " ") + ") ")
	} else {
		m.sb.WriteString("NIL ")
	}

	// ID and Description
	m.sb.WriteString(toNilString(b.ContentId) + " ")

	m.sb.WriteString("NIL ") // body description

	// Encoding and size
	m.sb.WriteString(fmt.Sprintf("\"%v\" ", b.Encoding))
	m.sb.WriteString(fmt.Sprintf("%v ", b.Size))

	// MD5
	m.sb.WriteString(toNilString(b.MD5) + " ")

	// Disposition
	if len(b.Disposition) > 0 {
		dispo := media.ParseContentType(b.Disposition)

		list := ""
		for k, v := range dispo.Parameters {
			if len(list) > 0 {
				list += " "
			}
			list += fmt.Sprintf("\"%v\" \"%v\"", k, v)
		}

		if len(list) == 0 {
			m.sb.WriteString(fmt.Sprintf("(\"%s\") ", dispo.Type))
		} else {
			m.sb.WriteString(fmt.Sprintf("(\"%s\" (%s)) ", dispo.Type, list))
		}
	} else {
		m.sb.WriteString("NIL ")
	}

	// Language and Location
	m.sb.WriteString(toNilString(b.Language) + " ")
	m.sb.WriteString(toNilString(b.Location))

	m.sb.WriteString(") ")
}

func toNilString(s *string) string {
	if s == nil {
		return "NIL"
	}
	return fmt.Sprintf("\"%v\"", s)
}
