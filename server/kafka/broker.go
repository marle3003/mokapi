package kafka

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi/kafka"
	"net"
	"sync"
	"time"
)

type broker struct {
	name string
	id   int
	host string
	port int

	listener             net.Listener
	closeListener        chan bool
	stopped              chan bool
	stopRetentionChecker chan bool
	config               kafka.BrokerBindings
}

var nextBrokerId = 0

func newBroker(name string, host string, port int, config kafka.BrokerBindings) *broker {
	b := &broker{name: name, id: nextBrokerId, host: host, port: port, config: config, closeListener: make(chan bool, 1), stopRetentionChecker: make(chan bool, 1), stopped: make(chan bool)}
	nextBrokerId++
	return b
}

type client struct {
	id            string
	group         *group
	lastHeartbeat time.Time
}

func (b *broker) start(c controller) {
	var err error
	b.listener, err = net.Listen("tcp", fmt.Sprintf("%v:%v", b.host, b.port))
	if err != nil {
		log.Errorf("Error listening: %v", err.Error())
		return
	}
	var handlers sync.WaitGroup

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := b.listener.Accept()
			if err != nil {
				select {
				case <-b.closeListener:
					handlers.Wait()
					b.stopped <- true
					return
				default:
					log.Errorf("Error accepting: %v", err.Error())
				}
			}
			// Handle connections in a new goroutine.
			handlers.Add(1)
			go func() {
				go c.handle(conn)
				handlers.Done()
			}()
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Duration(b.config.LogRetentionCheckIntervalMs()) * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-b.stopRetentionChecker:
				return
			case <-ticker.C:
				c.checkRetention(b)
			}
		}
	}()

	log.Infof("Started kafka broker %q with id %v on %v:%v", b.name, b.id, b.host, b.port)
}

func (b *broker) stop() {
	log.Infof("Stopping kafka broker %q with id %v on %v:%v", b.name, b.id, b.host, b.port)
	b.closeListener <- true
	b.stopRetentionChecker <- true
	err := b.listener.Close()
	if err != nil {
		log.Errorf("unable to stop kafka broker %q: %v", b.name, err.Error())
	}
	<-b.stopped
}

func (b *broker) newGroup(name string) *group {
	return newGroup(name, b, b.config.GroupInitialRebalanceDelayMs())
}
