package imaptest

import "mokapi/imap"

type CopyRecorder struct {
	Copy *imap.Copy
}

func (r *CopyRecorder) WriteCopy(copy *imap.Copy) error {
	r.Copy = copy
	return nil
}
