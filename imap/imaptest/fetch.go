package imaptest

import (
	"mokapi/imap"
	"time"
)

type FetchRecorder struct {
	Messages []*MessageRecorder
}

func (r *FetchRecorder) NewMessage(sequenceNumber uint32) imap.MessageWriter {
	m := &MessageRecorder{Msn: sequenceNumber}
	r.Messages = append(r.Messages, m)
	return m
}

type MessageRecorder struct {
	Msn           uint32
	Uid           uint32
	Flags         []imap.Flag
	InternalDate  time.Time
	Size          uint32
	Body          []*BodyRecorder
	Envelope      *imap.Envelope
	BodyStructure *imap.BodyStructure
}

func (r *MessageRecorder) WriteUID(uid uint32) {
	r.Uid = uid
}

func (r *MessageRecorder) WriteInternalDate(date time.Time) {
	r.InternalDate = date
}

func (r *MessageRecorder) WriteRFC822Size(size uint32) {
	r.Size = size
}

func (r *MessageRecorder) WriteFlags(flags ...imap.Flag) {
	r.Flags = append(r.Flags, flags...)

}
func (r *MessageRecorder) WriteEnvelope(env *imap.Envelope) {
	r.Envelope = env
}

func (r *MessageRecorder) WriteBody(section imap.FetchBodySection) imap.BodyWriter {
	w := &BodyRecorder{Section: section}
	r.Body = append(r.Body, w)
	return w
}

func (r *MessageRecorder) WriteBodyStructure(body *imap.BodyStructure) {
	r.BodyStructure = body
}

type BodyRecorder struct {
	Section imap.FetchBodySection
	Headers map[string]string
	Body    string
}

func (r *BodyRecorder) WriteHeader(name, value string) {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers[name] = value
}

func (r *BodyRecorder) WriteBody(s string) {
	r.Body = s
}

func (r *BodyRecorder) Close() {

}
