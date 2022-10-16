package smtp_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/smtp"
	"mokapi/smtp/smtptest"
	"testing"
)

func TestServer_Serve(t *testing.T) {
	testcases := []struct {
		name string
		h    func(rw smtp.ResponseWriter, r *smtp.Request)
		f    func(client *smtptest.Client)
	}{
		{
			"connect",
			func(rw smtp.ResponseWriter, r *smtp.Request) {

			},
			func(client *smtptest.Client) {
				r, err := client.Connect()
				require.NoError(t, err)
				require.Equal(t, "220 localhost ESMTP Service Ready", r.Message)
			},
		},
		{
			"say hello",
			func(rw smtp.ResponseWriter, r *smtp.Request) {
				if r.Cmd == smtp.Hello {
					rw.Write(smtp.StatusOk, smtp.Undefined, "foobar")
				} else {
					rw.Write(smtp.StatusSyntaxError, smtp.Undefined)
				}
			},
			func(client *smtptest.Client) {
				r, err := client.Connect()
				require.NoError(t, err)
				r, err = client.Send(&smtp.Request{Cmd: smtp.Hello})
				require.NoError(t, err)
				require.Equal(t, "250 foobar", r.Message)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			server := smtptest.NewServer(tc.h)
			defer server.Close()
			server.Start()

			client := smtptest.NewClient(server.Listener.Addr().String())
			defer client.Close()

			tc.f(client)
		})
	}
}
