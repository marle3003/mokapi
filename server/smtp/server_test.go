package smtp

import (
	"crypto/tls"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/mail"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/server/cert"
	"mokapi/server/smtp/smtptest"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	testcases := []struct {
		name   string
		config *mail.Config
		store  *cert.Store
		fn     func(t *testing.T)
	}{
		{
			name: "fixed ip:port",
			fn: func(t *testing.T) {
				server, err := New(&mail.Config{Server: "smtp://127.0.0.1:12345"}, nil, &eventEmitter{})
				require.NoError(t, err)
				err = server.Start()
				t.Cleanup(server.Stop)
				require.NoError(t, err)

				err = smtptest.SendMail("foo@foo.bar", "bar@foo.bar", "smtp://127.0.0.1:12345")
				require.NoError(t, err)
			},
		},
		{
			name: "simple config",
			fn: func(t *testing.T) {
				server, err := New(&mail.Config{}, nil, &eventEmitter{})
				require.NoError(t, err)
				l, err := net.Listen("tcp", "127.0.0.1:")
				require.NoError(t, err)
				server.StartWith(l)
				t.Cleanup(server.Stop)

				err = smtptest.SendMail("foo@foo.bar", "bar@foo.bar", fmt.Sprintf("smtp://%v", l.Addr().String()))
				require.NoError(t, err)
			},
		},
		{
			name: "with tls",
			fn: func(t *testing.T) {
				server, err := New(&mail.Config{}, nil, &eventEmitter{})
				require.NoError(t, err)
				store, err := cert.NewStore(&static.Config{})
				require.NoError(t, err)
				tlsConfig := &tls.Config{GetCertificate: store.GetCertificate}
				l, err := tls.Listen("tcp", "127.0.0.1:", tlsConfig)
				require.NoError(t, err)
				server.StartWith(l)
				t.Cleanup(server.Stop)

				err = smtptest.SendMail(
					"foo@foo.bar",
					"bar@foo.bar",
					fmt.Sprintf("smtps://%v", l.Addr().String()),
					smtptest.InsecureSkipVerfiy(),
				)
				require.NoError(t, err)
			},
		},
	}
	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			test.fn(t)
		})
	}
}

type eventEmitter struct {
}

func (e *eventEmitter) Emit(_ string, _ ...interface{}) []*common.Action {
	return nil
}
