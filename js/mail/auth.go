package mail

import (
	"fmt"
	"github.com/pkg/errors"
	"net/smtp"
)

type Auth struct {
	Plain *PlainAuth `json:"plain"`
	Login *LoginAuth `json:"login"`
}

func (a *Auth) getAuth() smtp.Auth {
	if a == nil {
		return nil
	}
	if a.Plain != nil {
		return a.Plain
	}
	if a.Login != nil {
		return a.Login
	}
	return nil
}

type LoginAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *LoginAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *LoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.Username), nil
		case "Password:":
			return []byte(a.Password), nil
		default:
			return nil, fmt.Errorf("unknown command from server: %v", fromServer)
		}
	}
	return nil, nil
}

type PlainAuth struct {
	Identity string `json:"identity"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *PlainAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	resp := []byte(a.Identity + "\x00" + a.Username + "\x00" + a.Password)
	return "PLAIN", resp, nil
}

func (a *PlainAuth) Next(_ []byte, more bool) ([]byte, error) {
	if more {
		// We've already sent everything.
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
}
