package server_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/dynamic/mail"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/imap/imaptest"
	"mokapi/runtime"
	"mokapi/server"
	"mokapi/server/cert"
	"mokapi/smtp/smtptest"
	"mokapi/try"
	"testing"
)

func TestSmtp(t *testing.T) {
	port := try.GetFreePort()

	testcases := []struct {
		name string
		test func(t *testing.T, m *server.SmtpManager)
	}{
		{
			name: "add smtp server",
			test: func(t *testing.T, m *server.SmtpManager) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info:   mail.Info{Name: "foo"},
							Server: fmt.Sprintf("smtp://localhost:%d", port),
						},
					},
				})

				err := smtptest.SendMail("from@foo.bar",
					"rcipient@foo.bar",
					fmt.Sprintf("smtp://localhost:%d", port),
					smtptest.WithSubject("Test Mail"),
				)
				require.NoError(t, err)
			},
		},
		{
			name: "update smtp server",
			test: func(t *testing.T, m *server.SmtpManager) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info:   mail.Info{Name: "foo"},
							Server: fmt.Sprintf("smtp://localhost:%d", port),
						},
					},
				})
				port2 := try.GetFreePort()
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info:   mail.Info{Name: "foo"},
							Server: fmt.Sprintf("smtp://localhost:%d", port2),
						},
					},
				})

				err := smtptest.SendMail("from@foo.bar",
					"rcipient@foo.bar",
					fmt.Sprintf("smtp://localhost:%d", port2),
					smtptest.WithSubject("Test Mail"),
				)
				require.NoError(t, err)
			},
		},
		{
			name: "delete event",
			test: func(t *testing.T, m *server.SmtpManager) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info:   mail.Info{Name: "foo"},
							Server: fmt.Sprintf("smtp://localhost:%d", port),
						},
					},
				})
				m.UpdateConfig(dynamic.ConfigEvent{
					Event: dynamic.Delete,
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info: mail.Info{Name: "foo"},
						},
					},
				})

				err := smtptest.SendMail("from@foo.bar",
					"rcipient@foo.bar",
					fmt.Sprintf("smtp://localhost:%d", port),
					smtptest.WithSubject("Test Mail"),
				)
				require.Error(t, err)
			},
		},
		{
			name: "delete event imap",
			test: func(t *testing.T, m *server.SmtpManager) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info:   mail.Info{Name: "foo"},
							Server: fmt.Sprintf("imap://localhost:%d", port),
						},
					},
				})
				m.UpdateConfig(dynamic.ConfigEvent{
					Event: dynamic.Delete,
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info: mail.Info{Name: "foo"},
						},
					},
				})

				c := imaptest.NewClient(fmt.Sprintf("smtp://localhost:%d", port))
				_, err := c.Dial()
				require.Error(t, err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			certStore, err := cert.NewStore(&static.Config{})
			require.NoError(t, err)

			m := server.NewSmtpManager(runtime.New(), enginetest.NewEngine(), certStore)
			defer m.Stop()

			tc.test(t, m)
		})
	}
}
