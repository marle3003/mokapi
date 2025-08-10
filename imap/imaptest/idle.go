package imaptest

import "mokapi/imap"

type UpdateRecorder struct {
	Messages [][]any
}

func (r *UpdateRecorder) WriteNumMessages(n uint32) error {
	r.Messages = append(r.Messages, []any{n})
	return nil
}

func (r *UpdateRecorder) WriteMessageFlags(msn uint32, flags []imap.Flag) error {
	r.Messages = append(r.Messages, []any{msn, flags})
	return nil
}

func (r *UpdateRecorder) WriteExpunge(msn uint32) error {
	r.Messages = append(r.Messages, []any{msn})
	return nil
}
