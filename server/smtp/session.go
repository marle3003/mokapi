package smtp

import (
	"github.com/emersion/go-smtp"
	"io"
	"io/ioutil"
)

type session struct {
	current  *Mail
	received chan *Mail
}

func newSession(received chan *Mail) *session {
	return &session{
		received: received,
	}
}

func (s *session) Reset() {
	s.received <- s.current
}

func (s *session) Logout() error {
	return nil
}

func (s *session) Mail(from string, opts smtp.MailOptions) error {
	s.current = &Mail{From: from}
	return nil
}

func (s *session) Rcpt(to string) error {
	s.current.To = to
	return nil
}

func (s *session) Data(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	s.current.Data = string(b)

	return nil
}
