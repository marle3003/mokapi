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
	WriteInternalDate(date time.Time)
	WriteRFC822Size(size int64)
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

func (m *message) WriteInternalDate(date time.Time) {
	m.sb.WriteString(fmt.Sprintf(" INTERNALDATE \"%v\"", date.Format(dateTimeLayout)))
}

func (m *message) WriteRFC822Size(size int64) {
	m.sb.WriteString(fmt.Sprintf(" RFC822.SIZE %v", size))
}
