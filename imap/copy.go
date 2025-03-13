package imap

type Copy struct {
	UIDValidity uint32
	SourceUIDs  IdSet
	DestUIDs    IdSet
}
