package runtime_test

import (
	"context"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"mokapi/smtp/smtptest"
	"net/url"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestApp_AddSmtp(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "event store available",
			test: func(t *testing.T, app *runtime.App) {
				app.Mail.Add(newSmtpConfig("https://mokapi.io", &mail.Config{Info: mail.Info{Name: "foo"}}))

				require.NotNil(t, app.Mail.Get("foo"))
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("mail").WithName("foo"))
				require.NoError(t, err, "event store should be available")
			},
		},
		{
			name: "send mail request is counted in monitor",
			test: func(t *testing.T, app *runtime.App) {
				info := app.Mail.Add(newSmtpConfig("https://mokapi.io", &mail.Config{Info: mail.Info{Name: "foo"}}))
				m := monitor.NewMail()
				h := info.Handler(m, enginetest.NewEngine(), app.Events)

				ctx := smtp.NewClientContext(context.Background(), "")
				rr := smtptest.NewRecorder()
				h.ServeSMTP(rr, smtp.NewDataRequest(&smtp.Message{}, ctx))

				require.Equal(t, float64(1), m.Mails.Sum(metrics.NewQuery()))
			},
		},
		{
			name: "retrieve configs",
			test: func(t *testing.T, app *runtime.App) {
				info := app.Mail.Add(newSmtpConfig("https://mokapi.io", &mail.Config{Info: mail.Info{Name: "foo"}}))

				configs := filesOf(info)
				require.Len(t, configs, 1)
				require.Equal(t, "https://mokapi.io", configs[0].Info.Url.String())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg, &dynamictest.Reader{})
			tc.test(t, app)
		})
	}
}

func TestApp_AddSmtp_Patching(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*dynamic.Config
		test    func(t *testing.T, app *runtime.App)
	}{
		{
			name: "overwrite value",
			configs: []*dynamic.Config{
				newSmtpConfig("https://mokapi.io/a", &mail.Config{Info: mail.Info{Name: "foo", Description: "foo"}}),
				newSmtpConfig("https://mokapi.io/b", &mail.Config{Info: mail.Info{Name: "foo", Description: "bar"}}),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Mail.Get("foo")
				require.Equal(t, "bar", info.Config().Info.Description)
				configs := filesOf(info)
				require.Len(t, configs, 2)
			},
		},
		{
			name: "a is patched with b",
			configs: []*dynamic.Config{
				newSmtpConfig("https://mokapi.io/b", &mail.Config{Info: mail.Info{Name: "foo", Description: "foo"}}),
				newSmtpConfig("https://mokapi.io/a", &mail.Config{Info: mail.Info{Name: "foo", Description: "bar"}}),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Mail.Get("foo")
				require.Equal(t, "foo", info.Config().Info.Description)
			},
		},
		{
			name: "order only by filename",
			configs: []*dynamic.Config{
				newSmtpConfig("https://a.io/b", &mail.Config{Info: mail.Info{Name: "foo", Description: "foo"}}),
				newSmtpConfig("https://mokapi.io/a", &mail.Config{Info: mail.Info{Name: "foo", Description: "bar"}}),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Mail.Get("foo")
				require.Equal(t, "foo", info.Config().Info.Description)
			},
		},
		{
			name: "patch does not reset events",
			configs: []*dynamic.Config{
				newSmtpConfig("https://a.io/b", &mail.Config{Info: mail.Info{Name: "foo", Description: "foo"}}),
			},
			test: func(t *testing.T, app *runtime.App) {
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("mail").WithName("foo"))
				require.NoError(t, err)

				app.Mail.Add(newSmtpConfig("https://mokapi.io/a", &mail.Config{Info: mail.Info{Name: "foo", Description: "bar"}}))

				e := app.Events.GetEvents(events.NewTraits().WithNamespace("mail").WithName("foo"))
				require.Len(t, e, 1)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg, &dynamictest.Reader{})
			for _, c := range tc.configs {
				app.Mail.Add(c)
			}
			tc.test(t, app)
		})
	}
}

func TestIsSmtpConfig(t *testing.T) {
	_, ok := runtime.IsSmtpConfig(dynamictest.NewConfig(&mail.Config{}))
	require.True(t, ok)
	_, ok = runtime.IsSmtpConfig(dynamictest.NewConfig("foo"))
	require.False(t, ok)
}

func newSmtpConfig(name string, config *mail.Config) *dynamic.Config {
	c := dynamictest.NewConfig(config)
	u, _ := url.Parse(name)
	c.Info.Url = u
	return c
}

func TestApp_AddMail_EventStore(t *testing.T) {
	mailTraits := func(name string) events.Traits {
		return events.NewTraits().WithNamespace("mail").WithName(name)
	}

	t.Run("event store size of the mail service is applied", func(t *testing.T) {
		cfg := &static.Config{}
		cfg.Event.Store = map[string]static.Store{"default": {Size: 100}, "foo": {Size: 2}}
		app := runtime.New(cfg, &dynamictest.Reader{})
		app.Mail.Add(newSmtpConfig("https://mokapi.io", &mail.Config{Info: mail.Info{Name: "foo"}}))

		for i := 0; i < 5; i++ {
			require.NoError(t, app.Events.Push(&eventstest.Event{Name: "bar"}, mailTraits("foo")))
		}

		require.Len(t, app.Events.GetEvents(mailTraits("foo")), 2)
	})

	t.Run("removing the mail service resets its events", func(t *testing.T) {
		cfg := &static.Config{}
		app := runtime.New(cfg, &dynamictest.Reader{})
		c := newSmtpConfig("https://mokapi.io", &mail.Config{Info: mail.Info{Name: "foo"}})
		app.Mail.Add(c)
		require.NoError(t, app.Events.Push(&eventstest.Event{Name: "bar"}, mailTraits("foo")))

		app.Mail.Remove(c)

		require.Empty(t, app.Events.GetEvents(mailTraits("foo")))
	})
}

func TestMailStore_Remove(t *testing.T) {
	logrus.SetOutput(io.Discard)

	newFile := func(url, name string) *dynamic.Config {
		return newSmtpConfig(url, &mail.Config{Info: mail.Info{Name: name}})
	}

	t.Run("unknown config does not panic and keeps the store usable", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})

		requireStoreUsableAfterRemoveOfUnknownConfig(t,
			func() { app.Mail.Remove(newFile("https://mokapi.io/unknown.yml", "unknown")) },
			func() { app.Mail.Add(newFile("https://mokapi.io/a.yml", "foo")) },
		)
	})

	t.Run("two files removed at the same time", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a := newFile("https://mokapi.io/a.yml", "foo")
		b := newFile("https://mokapi.io/b.yml", "foo")
		app.Mail.Add(a)
		app.Mail.Add(b)

		removeConcurrently(app.Mail.Remove, a, b)

		require.Equal(t, 0, app.Mail.Len())
	})
}

// mailFile returns a file of the service "foo" with one configured mailbox and a description.
func mailFile(url, description string, mailboxes ...string) *dynamic.Config {
	cfg := &mail.Config{Info: mail.Info{Name: "foo", Description: description}, Mailboxes: map[string]*mail.MailboxConfig{}}
	for _, name := range mailboxes {
		cfg.Mailboxes[name] = &mail.MailboxConfig{Folders: map[string]*mail.FolderConfig{"INBOX": {}}}
	}
	return newSmtpConfig(url, cfg)
}

func TestMailStore_ConcurrentAccess(t *testing.T) {
	logrus.SetOutput(io.Discard)

	t.Run("reader does not race with reload of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		app.Mail.Add(mailFile("https://mokapi.io/a.yml", "first", "alice@foo.bar"))

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				app.Mail.Add(mailFile("https://mokapi.io/a.yml", "second", "alice@foo.bar", "bob@foo.bar"))
			}
		}()
		go func() {
			defer wg.Done()
			// what the dashboard API handler does: get the info, then read the config and the store
			for i := 0; i < 200; i++ {
				info := app.Mail.Get("foo")
				cfg := info.Config()
				_ = cfg.Info.Name
				_ = cfg.Mailboxes
				for _, mb := range info.Mailboxes() {
					_ = mb.NumMessages()
				}
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})

	t.Run("reader does not race with remove", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a := mailFile("https://mokapi.io/a.yml", "", "alice@foo.bar")
		b := mailFile("https://mokapi.io/b.yml", "", "bob@foo.bar")
		info := app.Mail.Add(a)

		done := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			defer close(done)
			for i := 0; i < 200; i++ {
				app.Mail.Add(b)
				app.Mail.Remove(b)
			}
		}()
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				if cfg := info.Config(); cfg != nil {
					_ = cfg.Info.Name
				}
				_ = info.Mailboxes()
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})
}

func TestMailStore_MergedConfigs(t *testing.T) {
	logrus.SetOutput(io.Discard)

	newFirst := func() *mail.Config {
		return &mail.Config{
			Info:      mail.Info{Name: "foo", Description: "first"},
			Servers:   map[string]*mail.Server{"smtp": {Host: ":25", Protocol: "smtp"}},
			Settings:  &mail.Settings{MaxRecipients: 1},
			Mailboxes: map[string]*mail.MailboxConfig{"alice@foo.bar": {Folders: map[string]*mail.FolderConfig{"INBOX": {Flags: []string{"a"}}}}},
			Rules:     map[string]*mail.Rule{"r": {Sender: mail.NewRuleExpr("a")}},
		}
	}
	second := &mail.Config{
		Info:      mail.Info{Name: "foo", Description: "second"},
		Servers:   map[string]*mail.Server{"smtp": {Host: ":2525"}, "imap": {Host: ":143", Protocol: "imap"}},
		Settings:  &mail.Settings{MaxRecipients: 2},
		Mailboxes: map[string]*mail.MailboxConfig{"alice@foo.bar": {Password: "secret", Folders: map[string]*mail.FolderConfig{"INBOX": {Flags: []string{"b"}}}}, "bob@foo.bar": {}},
		Rules:     map[string]*mail.Rule{"r": {Sender: mail.NewRuleExpr("b")}, "s": {}},
	}

	t.Run("merge does not change the parsed file config", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		first := newFirst()
		app.Mail.Add(newSmtpConfig("https://mokapi.io/a.yml", first))

		app.Mail.Add(newSmtpConfig("https://mokapi.io/b.yml", second))

		require.Equal(t, newFirst(), first)
	})

	t.Run("servers of both files are published", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		app.Mail.Add(newSmtpConfig("https://mokapi.io/a.yml", newFirst()))

		info := app.Mail.Add(newSmtpConfig("https://mokapi.io/b.yml", second))

		cfg := info.Config()
		require.Equal(t, ":2525", cfg.Servers["smtp"].Host)
		require.Equal(t, ":143", cfg.Servers["imap"].Host)
		require.Equal(t, 2, cfg.Settings.MaxRecipients)
	})
}

func TestMailInfo_Snapshot(t *testing.T) {
	logrus.SetOutput(io.Discard)

	t.Run("snapshot is not changed by a reload of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		info := app.Mail.Add(mailFile("https://mokapi.io/a.yml", "first", "alice@foo.bar"))
		snapshot := info.Config()

		app.Mail.Add(mailFile("https://mokapi.io/a.yml", "second", "bob@foo.bar"))

		require.Equal(t, "first", snapshot.Info.Description)
		require.Contains(t, snapshot.Mailboxes, "alice@foo.bar")
		require.NotContains(t, snapshot.Mailboxes, "bob@foo.bar")
		require.NotSame(t, snapshot, info.Config())
		require.Equal(t, "second", info.Config().Info.Description)
	})

	t.Run("published config is not the config of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := mailFile("https://mokapi.io/a.yml", "first", "alice@foo.bar")

		info := app.Mail.Add(file)

		fileConfig := file.Data().(*mail.Config)
		require.NotSame(t, fileConfig, info.Config())
		info.Config().Mailboxes["alice@foo.bar"].Password = "changed"
		require.Empty(t, fileConfig.Mailboxes["alice@foo.bar"].Password)
	})

	t.Run("config is nil after the last file is removed", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := mailFile("https://mokapi.io/a.yml", "first")
		info := app.Mail.Add(file)

		app.Mail.Remove(file)

		require.Nil(t, info.Config())
		require.Nil(t, app.Mail.Get("foo"))
	})

	t.Run("one merged file removed keeps the service", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a := mailFile("https://mokapi.io/a.yml", "first", "alice@foo.bar")
		b := mailFile("https://mokapi.io/b.yml", "second", "bob@foo.bar")
		info := app.Mail.Add(a)
		app.Mail.Add(b)
		require.Len(t, info.Config().Mailboxes, 2)

		app.Mail.Remove(b)

		require.Same(t, info, app.Mail.Get("foo"))
		require.Equal(t, "first", info.Config().Info.Description)
		require.Len(t, info.Config().Mailboxes, 1)
		require.Len(t, info.Mailboxes(), 1)
	})

	t.Run("store and its mails survive a reload", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		info := app.Mail.Add(mailFile("https://mokapi.io/a.yml", "first", "alice@foo.bar"))
		store := info.Store
		mb, _ := info.Mailbox("alice@foo.bar")
		mb.Append(&smtp.Message{MessageId: "m1"})

		app.Mail.Add(mailFile("https://mokapi.io/a.yml", "second", "alice@foo.bar"))

		require.Same(t, store, info.Store)
		again, _ := info.Mailbox("alice@foo.bar")
		require.Same(t, mb, again)
		require.NotNil(t, info.GetMail("m1"))
	})

	t.Run("store stays after the last file is removed", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := mailFile("https://mokapi.io/a.yml", "first", "alice@foo.bar")
		info := app.Mail.Add(file)

		app.Mail.Remove(file)

		require.NotNil(t, info.Store)
	})

	t.Run("Configs returns a copy", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		app.Mail.Add(mailFile("https://mokapi.io/a.yml", ""))
		info := app.Mail.Add(mailFile("https://mokapi.io/b.yml", ""))

		configs := filesOf(info)
		require.Len(t, configs, 2)
		configs[0] = nil

		require.NotContains(t, filesOf(info), (*dynamic.Config)(nil))
		require.Len(t, filesOf(info), 2)
	})

	t.Run("Configs does not race with Add", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		info := app.Mail.Add(mailFile("https://mokapi.io/a.yml", ""))
		b := mailFile("https://mokapi.io/b.yml", "")

		done := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			defer close(done)
			for i := 0; i < 200; i++ {
				app.Mail.Add(b)
			}
		}()
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})
}
