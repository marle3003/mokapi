package kafka_test

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"mokapi/kafka/kafkatest"
	"net"
	"testing"
	"time"
)

func TestBroker_Disconnect(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	hook := test.NewGlobal()
	b := kafkatest.NewBroker(
		kafkatest.WithHandler(
			kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {
				rw.Write(&apiVersion.Response{})
			})))

	var err error
	var conn net.Conn

	for i := 0; i < 10; i++ {
		d := net.Dialer{Timeout: time.Second * 10}
		conn, err = d.Dial("tcp", b.Addr)
		if err != nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
	}
	require.NoError(t, err)

	r := &kafka.Request{
		Header: &kafka.Header{
			ApiKey:     kafka.ApiVersions,
			ApiVersion: 0,
		},
		Message: &apiVersion.Request{},
	}

	r.Write(conn)
	conn.Close()
	time.Sleep(1000 * time.Millisecond)
	// should not log any panic message
	require.Nil(t, hook.LastEntry(), "there should be no log message")
}
