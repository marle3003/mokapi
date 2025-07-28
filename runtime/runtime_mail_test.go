package runtime_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"mokapi/smtp/smtptest"
	"net/url"
	"testing"
)

func TestApp_AddSmtp(t *testing.T) {
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

				require.Equal(t, float64(1), m.Mails.Sum())
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
			app := runtime.New(cfg)
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
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("smtp").WithName("foo"))
				require.NoError(t, err)

				app.Mail.Add(newSmtpConfig("https://mokapi.io/a", &mail.Config{Info: mail.Info{Name: "foo", Description: "bar"}}))

				e := app.Events.GetEvents(events.NewTraits().WithNamespace("smtp").WithName("foo"))
				require.Len(t, e, 1)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg)
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
