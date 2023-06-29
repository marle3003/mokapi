package sasl

type Client interface {
	Next(challenge []byte) (response []byte, err error)
	HasNext() bool
}

type Server interface {
	Next(response []byte) (challenge []byte, err error)
	HasNext() bool
}

type Handler func()

type ClientOptions func(c Client)

type ServerOptions func(s Server)
