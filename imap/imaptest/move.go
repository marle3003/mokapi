package imaptest

import "mokapi/imap"

type MoveRecorder struct {
	Copy *imap.Copy
	Ids  []uint32
}

func (r *MoveRecorder) WriteCopy(copy *imap.Copy) error {
	r.Copy = copy
	return nil
}

func (r *MoveRecorder) WriteExpunge(id uint32) error {
	r.Ids = append(r.Ids, id)
	return nil
}
