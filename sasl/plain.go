package sasl

import (
	"bytes"
	"github.com/pkg/errors"
)

type PlainAuthenticator func(identity, username, password string) error

type plainClient struct {
	Identity string
	Username string
	Password string
}

func (c *plainClient) Next(_ []byte) (response []byte, err error) {
	response = []byte(c.Identity + "\x00" + c.Username + "\x00" + c.Password)
	return
}

func (c *plainClient) HasNext() bool {
	return false
}

func NewPlainClient(identity, username, password string) Client {
	return &plainClient{
		Identity: identity,
		Username: username,
		Password: password,
	}
}

type plainServer struct {
	auth    PlainAuthenticator
	hasNext bool
}

func (s *plainServer) Next(response []byte) (challenge []byte, err error) {
	if len(response) == 0 {
		return
	}

	s.hasNext = false

	parts := bytes.Split(response, []byte("\x00"))
	if len(parts) != 3 {
		err = errors.New("Invalid response")
		return
	}

	identity := string(parts[0])
	username := string(parts[1])
	password := string(parts[2])

	err = s.auth(identity, username, password)
	return
}

func (s *plainServer) HasNext() bool {
	return s.hasNext
}

func NewPlainServer(auth PlainAuthenticator) Server {
	s := &plainServer{auth: auth, hasNext: true}
	return s
}
