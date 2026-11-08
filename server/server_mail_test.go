package server_test

import (
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/imap"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/server"
	"mokapi/server/cert"
	"mokapi/smtp/smtptest"
	"mokapi/try"
	"net"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSmtp(t *testing.T) {
	logrus.SetOutput(io.Discard)
	port := try.GetFreePort()

	testcases := []struct {
		name string
		test func(t *testing.T, m *server.MailManager)
	}{
		{
			name: "add smtp server",
			test: func(t *testing.T, m *server.MailManager) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info: mail.Info{Name: "foo"},
							Servers: map[string]*mail.Server{
								"foo": {
									Host:     fmt.Sprintf("localhost:%d", port),
									Protocol: "smtp",
								},
							},
							Settings: &mail.Settings{AutoCreateMailbox: true},
						},
					},
				})
				waitForPort(t, port)

				err := smtptest.SendMail("from@foo.bar",
					"rcipient@foo.bar",
					fmt.Sprintf("smtp://localhost:%d", port),
					smtptest.WithSubject("Test Mail"),
				)
				require.NoError(t, err)
			},
		},
		{
			name: "add smtp server only port defined",
			test: func(t *testing.T, m *server.MailManager) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info: mail.Info{Name: "foo"},
							Servers: map[string]*mail.Server{
								"foo": {
									Host:     fmt.Sprintf(":%d", port),
									Protocol: "smtp",
								},
							},
							Settings: &mail.Settings{AutoCreateMailbox: true},
						},
					},
				})
				waitForPort(t, port)

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
			test: func(t *testing.T, m *server.MailManager) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info: mail.Info{Name: "foo"},
							Servers: map[string]*mail.Server{
								"foo": {
									Host:     fmt.Sprintf("localhost:%d", port),
									Protocol: "smtp",
								},
							},
							Settings: &mail.Settings{AutoCreateMailbox: true},
						},
					},
				})

				port2 := try.GetFreePort()
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info: mail.Info{Name: "foo"},
							Servers: map[string]*mail.Server{
								"foo": {
									Host:     fmt.Sprintf("localhost:%d", port2),
									Protocol: "smtp",
								},
							},
							Settings: &mail.Settings{AutoCreateMailbox: true},
						},
					},
				})
				waitForPort(t, port2)

				err := smtptest.SendMail("from@foo.bar",
					"rcipient@foo.bar",
					fmt.Sprintf("smtp://localhost:%d", port2),
					smtptest.WithSubject("Test Mail"),
				)
				require.NoError(t, err)

				err = smtptest.SendMail("from@foo.bar",
					"rcipient@foo.bar",
					fmt.Sprintf("smtp://localhost:%d", port),
					smtptest.WithSubject("Test Mail"),
				)
				require.Error(t, err)
			},
		},
		{
			name: "delete event",
			test: func(t *testing.T, m *server.MailManager) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info: mail.Info{Name: "foo"},
							Servers: map[string]*mail.Server{
								"foo": {
									Host:     fmt.Sprintf("localhost:%d", port),
									Protocol: "smtp",
								},
							},
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
			test: func(t *testing.T, m *server.MailManager) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: &mail.Config{
							Info: mail.Info{Name: "foo"},
							Servers: map[string]*mail.Server{
								"foo": {
									Host:     fmt.Sprintf("localhost:%d", port),
									Protocol: "smtp",
								},
							},
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

				c := imap.NewClient(fmt.Sprintf("smtp://localhost:%d", port))
				_, err := c.Dial()
				require.Error(t, err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			certStore, err := cert.NewStore(&static.Config{})
			require.NoError(t, err)

			cfg := &static.Config{}
			m := server.NewMailManager(runtime.New(cfg, &dynamictest.Reader{}), enginetest.NewEngine(), certStore)
			defer m.Stop()

			tc.test(t, m)
		})
	}
}

func waitForPort(t *testing.T, port int) {
	t.Helper()
	try.WaitFor(t, func() bool {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	})
}

var mailProtocols = []string{"smtp", "smtps", "imap", "imaps"}

func TestMailManager_PortIsBoundWhenUpdateConfigReturns(t *testing.T) {
	for _, protocol := range mailProtocols {
		t.Run(protocol, func(t *testing.T) {
			m := newMailManager(t)
			requireBoundWhenUpdateReturns(t, func(port int) { m.UpdateConfig(mailConfigEvent(protocol, port)) })
		})
	}
}

func TestMailManager_BindErrorIsLoggedAndServerStaysRegistered(t *testing.T) {
	for _, protocol := range mailProtocols {
		t.Run(protocol, func(t *testing.T) {
			m := newMailManager(t)
			requireBindErrorIsLoggedAndServerStaysRegistered(t,
				func(port int) { m.UpdateConfig(mailConfigEvent(protocol, port)) },
				func(port int) string { return fmt.Sprintf("listen tcp :%d: bind: address already in use", port) },
			)
		})
	}
}

func newMailManager(t *testing.T) *server.MailManager {
	certStore, err := cert.NewStore(&static.Config{})
	require.NoError(t, err)
	m := server.NewMailManager(runtime.New(&static.Config{}, &dynamictest.Reader{}), enginetest.NewEngine(), certStore)
	t.Cleanup(m.Stop)
	return m
}

func mailConfigEvent(protocol string, port int) dynamic.ConfigEvent {
	return dynamic.ConfigEvent{
		Config: &dynamic.Config{
			Info: dynamictest.NewConfigInfo(),
			Data: &mail.Config{
				Info: mail.Info{Name: "foo"},
				Servers: map[string]*mail.Server{
					"foo": {Host: fmt.Sprintf("localhost:%d", port), Protocol: protocol},
				},
				Settings: &mail.Settings{AutoCreateMailbox: true},
			},
		},
	}
}

func TestMailManager_DuplicateUpdateDoesNotOrphanRunningServer(t *testing.T) {
	for _, protocol := range mailProtocols {
		t.Run(protocol, func(t *testing.T) {
			m := newMailManager(t)
			port := try.GetFreePort()

			// e.g. the file provider reloads an unchanged mail config
			m.UpdateConfig(mailConfigEvent(protocol, port))
			m.UpdateConfig(mailConfigEvent(protocol, port))
			waitForPort(t, port)

			deleted := mailConfigEvent(protocol, port)
			deleted.Event = dynamic.Delete
			m.UpdateConfig(deleted)

			requirePortIsFree(t, port)
		})
	}
}

func TestMailManager_ServerCanBeReAddedAfterItWasRemovedByConfigUpdate(t *testing.T) {
	for _, protocol := range mailProtocols {
		t.Run(protocol, func(t *testing.T) {
			m := newMailManager(t)
			port := try.GetFreePort()
			otherPort := try.GetFreePort()

			m.UpdateConfig(mailConfigEvent(protocol, port))
			m.UpdateConfig(mailConfigEvent(protocol, otherPort))
			requirePortIsFree(t, port)

			m.UpdateConfig(mailConfigEvent(protocol, port))
			waitForPort(t, port)
		})
	}
}

func TestMailManager_ServerCanBeReAddedAfterConfigWasDeleted(t *testing.T) {
	for _, protocol := range mailProtocols {
		t.Run(protocol, func(t *testing.T) {
			m := newMailManager(t)
			port := try.GetFreePort()

			m.UpdateConfig(mailConfigEvent(protocol, port))
			deleted := mailConfigEvent(protocol, port)
			deleted.Event = dynamic.Delete
			m.UpdateConfig(deleted)
			requirePortIsFree(t, port)

			m.UpdateConfig(mailConfigEvent(protocol, port))
			waitForPort(t, port)
		})
	}
}

func requirePortIsFree(t *testing.T, port int) {
	t.Helper()
	try.WaitFor(t, func() bool {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return false
		}
		_ = l.Close()
		return true
	})
}
