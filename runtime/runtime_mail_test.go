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

				configs := info.Configs()
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
				require.Equal(t, "bar", info.Info.Description)
				configs := info.Configs()
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
				require.Equal(t, "foo", info.Info.Description)
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
				require.Equal(t, "foo", info.Info.Description)
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
	_, ok := runtime.IsSmtpConfig(&dynamic.Config{Data: &mail.Config{}})
	require.True(t, ok)
	_, ok = runtime.IsSmtpConfig(&dynamic.Config{Data: "foo"})
	require.False(t, ok)
}

func newSmtpConfig(name string, config *mail.Config) *dynamic.Config {
	c := &dynamic.Config{Data: config}
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
