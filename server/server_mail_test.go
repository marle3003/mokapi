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
	"sync"
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
					Config: dynamictest.NewConfig(&mail.Config{
						Info: mail.Info{Name: "foo"},
						Servers: map[string]*mail.Server{
							"foo": {
								Host:     fmt.Sprintf("localhost:%d", port),
								Protocol: "smtp",
							},
						},
						Settings: &mail.Settings{AutoCreateMailbox: true},
					}, dynamictest.WithInfo(dynamictest.NewConfigInfo())),
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
					Config: dynamictest.NewConfig(&mail.Config{
						Info: mail.Info{Name: "foo"},
						Servers: map[string]*mail.Server{
							"foo": {
								Host:     fmt.Sprintf(":%d", port),
								Protocol: "smtp",
							},
						},
						Settings: &mail.Settings{AutoCreateMailbox: true},
					}, dynamictest.WithInfo(dynamictest.NewConfigInfo())),
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
					Config: dynamictest.NewConfig(&mail.Config{
						Info: mail.Info{Name: "foo"},
						Servers: map[string]*mail.Server{
							"foo": {
								Host:     fmt.Sprintf("localhost:%d", port),
								Protocol: "smtp",
							},
						},
						Settings: &mail.Settings{AutoCreateMailbox: true},
					}, dynamictest.WithInfo(dynamictest.NewConfigInfo())),
				})

				port2 := try.GetFreePort()
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: dynamictest.NewConfig(&mail.Config{
						Info: mail.Info{Name: "foo"},
						Servers: map[string]*mail.Server{
							"foo": {
								Host:     fmt.Sprintf("localhost:%d", port2),
								Protocol: "smtp",
							},
						},
						Settings: &mail.Settings{AutoCreateMailbox: true},
					}, dynamictest.WithInfo(dynamictest.NewConfigInfo())),
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
					Config: dynamictest.NewConfig(&mail.Config{
						Info: mail.Info{Name: "foo"},
						Servers: map[string]*mail.Server{
							"foo": {
								Host:     fmt.Sprintf("localhost:%d", port),
								Protocol: "smtp",
							},
						},
					}, dynamictest.WithInfo(dynamictest.NewConfigInfo())),
				})
				m.UpdateConfig(dynamic.ConfigEvent{
					Event: dynamic.Delete,
					Config: dynamictest.NewConfig(&mail.Config{
						Info: mail.Info{Name: "foo"},
					}, dynamictest.WithInfo(dynamictest.NewConfigInfo())),
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
					Config: dynamictest.NewConfig(&mail.Config{
						Info: mail.Info{Name: "foo"},
						Servers: map[string]*mail.Server{
							"foo": {
								Host:     fmt.Sprintf("localhost:%d", port),
								Protocol: "smtp",
							},
						},
					}, dynamictest.WithInfo(dynamictest.NewConfigInfo())),
				})
				m.UpdateConfig(dynamic.ConfigEvent{
					Event: dynamic.Delete,
					Config: dynamictest.NewConfig(&mail.Config{
						Info: mail.Info{Name: "foo"},
					}, dynamictest.WithInfo(dynamictest.NewConfigInfo())),
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
		Config: dynamictest.NewConfig(&mail.Config{
			Info: mail.Info{Name: "foo"},
			Servers: map[string]*mail.Server{
				"foo": {Host: fmt.Sprintf("localhost:%d", port), Protocol: protocol},
			},
			Settings: &mail.Settings{AutoCreateMailbox: true},
		}, dynamictest.WithInfo(dynamictest.NewConfigInfo())),
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

func TestMailManager_DeleteOfUnknownConfigDoesNotPanic(t *testing.T) {
	for _, protocol := range mailProtocols {
		t.Run(protocol, func(t *testing.T) {
			m := newMailManager(t)
			e := mailConfigEvent(protocol, try.GetFreePort())
			e.Event = dynamic.Delete

			require.NotPanics(t, func() { m.UpdateConfig(e) })
		})
	}
}

func mailFileEvent(file, title string, servers map[string]*mail.Server, rules mail.Rules) dynamic.ConfigEvent {
	return dynamic.ConfigEvent{Config: dynamictest.NewConfig(&mail.Config{
		Info:     mail.Info{Name: title},
		Servers:  servers,
		Settings: &mail.Settings{AutoCreateMailbox: true},
		Rules:    rules,
	}, dynamictest.WithInfo(dynamic.ConfigInfo{Url: try.MustUrl(file)}))}
}

func smtpServer(port int) *mail.Server {
	return &mail.Server{Host: fmt.Sprintf("localhost:%d", port), Protocol: "smtp"}
}

var denySpam = mail.Rules{"denySpam": {Subject: mail.NewRuleExpr("spam"), Action: mail.Deny}}

func sendTestMail(port int, subject string) error {
	return smtptest.SendMail("from@foo.bar", "rcipient@foo.bar", fmt.Sprintf("smtp://localhost:%d", port), smtptest.WithSubject(subject))
}

// A reload of the file changes the rules of the SMTP server that is already running.
func TestMailManager_ChangedRuleAppliesToRunningServer(t *testing.T) {
	m := newMailManager(t)
	port := try.GetFreePort()
	servers := map[string]*mail.Server{"smtp": smtpServer(port)}

	m.UpdateConfig(mailFileEvent("a.yml", "foo", servers, nil))
	waitForPort(t, port)
	require.NoError(t, sendTestMail(port, "spam"))

	m.UpdateConfig(mailFileEvent("a.yml", "foo", servers, denySpam))
	require.Error(t, sendTestMail(port, "spam"))
	require.NoError(t, sendTestMail(port, "hello"))

	m.UpdateConfig(mailFileEvent("a.yml", "foo", servers, nil))
	require.NoError(t, sendTestMail(port, "spam"))
}

// The servers of all files of a service run, and the servers of a file stop when the file is
// deleted.
func TestMailManager_ServerOfMergedFileStartsAndStopsWithFile(t *testing.T) {
	m := newMailManager(t)
	portA, portB := try.GetFreePort(), try.GetFreePort()
	a := mailFileEvent("a.yml", "foo", map[string]*mail.Server{"a": smtpServer(portA)}, nil)
	b := mailFileEvent("b.yml", "foo", map[string]*mail.Server{"b": smtpServer(portB)}, nil)

	m.UpdateConfig(a)
	m.UpdateConfig(b)
	waitForPort(t, portA)
	waitForPort(t, portB)

	deleted := b
	deleted.Event = dynamic.Delete
	m.UpdateConfig(deleted)

	requirePortIsFree(t, portB)
	requirePortOpen(t, portA)
}

// Requests keep being served while reloads swap the handler. Run with -race.
func TestMailManager_HandlerSwapWhileServingRequests(t *testing.T) {
	m := newMailManager(t)
	port := try.GetFreePort()
	servers := map[string]*mail.Server{"smtp": smtpServer(port)}
	m.UpdateConfig(mailFileEvent("a.yml", "foo", servers, nil))
	waitForPort(t, port)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 30; i++ {
			rules := mail.Rules(nil)
			if i%2 == 0 {
				rules = denySpam
			}
			m.UpdateConfig(mailFileEvent("a.yml", "foo", servers, rules))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 30; i++ {
			_ = sendTestMail(port, "hello")
		}
	}()
	wg.Wait()

	require.NoError(t, sendTestMail(port, "spam"), "last reload removed the rule")
}
