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
	Msn   uint32
	Uid   uint32
	Flags []imap.Flag
}

func (r *MessageRecorder) WriteUID(uid uint32) {
	r.Uid = uid
}

func (r *MessageRecorder) WriteInternalDate(date time.Time) {

}

func (r *MessageRecorder) WriteRFC822Size(size uint32) {

}

func (r *MessageRecorder) WriteFlags(flags ...imap.Flag) {
	r.Flags = append(r.Flags, flags...)

}
func (r *MessageRecorder) WriteEnvelope(env *imap.Envelope) {

}
func (r *MessageRecorder) WriteBody(section imap.FetchBodySection) *imap.BodyWriter {
	return nil
}

func (r *MessageRecorder) WriteBodyStructure(body imap.BodyStructure) {

}
