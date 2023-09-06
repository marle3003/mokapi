package try

import "net"

func GetFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}
