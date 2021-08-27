package kafka

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type broker struct {
	name string
	id   int
	host string
	port int

	stopCh chan bool
}

func newBroker(name string, id int, host string, port int) *broker {
	return &broker{name: name, id: id, host: host, port: port, stopCh: make(chan bool)}
}

type client struct {
	id            string
	group         *group
	lastHeartbeat time.Time
}

func (b *broker) start(handler func(net.Conn)) {
	listen := fmt.Sprintf(":%v", b.port)
	l, err := net.Listen("tcp", listen)
	if err != nil {
		log.Errorf("Error listening: %v", err.Error())
		return
	}

	log.Infof("Started kafka broker %q with id %v on %v", b.name, b.id, listen)

	// Close the listener when the application closes.
	connChannl := make(chan net.Conn)
	closeListener := make(chan bool)
	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-closeListener:
					return
				default:
					log.Errorf("Error accepting: %v", err.Error())
				}
			}
			// Handle connections in a new goroutine.
			connChannl <- conn
		}
	}()

	go func() {
		for {
			select {
			case conn := <-connChannl:
				go handler(conn)
			case <-b.stopCh:
				log.Infof("Stopping kafka broker %q with id %v on %v", b.name, b.id, listen)
				closeListener <- true
				err := l.Close()
				if err != nil {
					log.Errorf("unable to stop kafka broker %q: %v", b.name, err.Error())
				}
			}
		}
	}()
}

func (b *broker) stop() {
	b.stopCh <- true
}
