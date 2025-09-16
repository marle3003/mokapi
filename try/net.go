package try

import (
	"net"
	"net/url"
)

func GetFreePort() int {
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:")

	l, _ := net.ListenTCP("tcp", addr)
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

func MustUrl(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
