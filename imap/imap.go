package imap

type ConnState uint8

const (
	NotAuthenticated ConnState = iota
	Authenticated
	Selected
	Logout
)

var crnl = []byte{'\r', '\n'}
