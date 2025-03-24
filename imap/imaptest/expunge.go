package imaptest

type ExpungeRecorder struct {
	Ids []uint32
}

func (r *ExpungeRecorder) Write(id uint32) error {
	r.Ids = append(r.Ids, id)
	return nil
}
