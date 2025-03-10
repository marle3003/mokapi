package mail_test

import (
	"fmt"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/mail"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	mod "mokapi/js/mail"
	"mokapi/js/require"
	"mokapi/smtp/smtptest"
	"testing"
)

func TestMailModule(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *mail.Config
		run  func(t *testing.T, vm *goja.Runtime, addr string)
		test func(t *testing.T, store *mail.Store)
	}{
		{
			name: "Send anonymous mail",
			run: func(t *testing.T, vm *goja.Runtime, addr string) {
				_, err := vm.RunString(fmt.Sprintf(`const m = require('mokapi/smtp')
m.send('smtp://%v', {from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'})
`, addr))
				r.NoError(t, err)
			},
			test: func(t *testing.T, store *mail.Store) {
				msg := store.Mailboxes["bob@mokapi.io"].Folders["INBOX"].Messages[0]
				r.Equal(t, "A test mail", msg.Subject)
				r.Equal(t, "Hello Bob", msg.Body)
				r.Equal(t, "Alice", msg.From[0].Name)
				r.Equal(t, "alice@mokapi.io", msg.From[0].Address)
			},
		},
		{
			name: "Send mail with auth login",
			cfg: &mail.Config{Mailboxes: []mail.MailboxConfig{
				{
					Name:     "alice@mokapi.io",
					Username: "alice",
					Password: "secret",
				},
				{
					Name: "bob@mokapi.io",
				},
			}},
			run: func(t *testing.T, vm *goja.Runtime, addr string) {
				_, err := vm.RunString(fmt.Sprintf(`const m = require('mokapi/smtp')
m.send('smtp://%v', {from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'},
{ login: { username: 'alice', password: 'secret' } })
`, addr))
				r.NoError(t, err)
			},
			test: func(t *testing.T, store *mail.Store) {
				msg := store.Mailboxes["bob@mokapi.io"].Folders["INBOX"].Messages[0]
				r.Equal(t, "A test mail", msg.Subject)
				r.Equal(t, "Hello Bob", msg.Body)
				r.Equal(t, "Alice", msg.From[0].Name)
				r.Equal(t, "alice@mokapi.io", msg.From[0].Address)
			},
		},
		{
			name: "Send mail with auth plain",
			cfg: &mail.Config{Mailboxes: []mail.MailboxConfig{
				{
					Name:     "alice@mokapi.io",
					Username: "alice",
					Password: "secret",
				},
				{
					Name: "bob@mokapi.io",
				},
			}},
			run: func(t *testing.T, vm *goja.Runtime, addr string) {
				_, err := vm.RunString(fmt.Sprintf(`const m = require('mokapi/smtp')
m.send('smtp://%v', {from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'},
{ plain: { username: 'alice', password: 'secret' } })
`, addr))
				r.NoError(t, err)
			},
			test: func(t *testing.T, store *mail.Store) {
				msg := store.Mailboxes["bob@mokapi.io"].Folders["INBOX"].Messages[0]
				r.Equal(t, "A test mail", msg.Subject)
				r.Equal(t, "Hello Bob", msg.Body)
				r.Equal(t, "Alice", msg.From[0].Name)
				r.Equal(t, "alice@mokapi.io", msg.From[0].Address)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg
			if cfg == nil {
				cfg = &mail.Config{}
			}
			s := mail.NewStore(cfg)
			h := mail.NewHandler(cfg, s, enginetest.NewEngine())
			server, _, err := smtptest.NewServer(h.ServeSMTP)
			r.NoError(t, err)
			defer server.Close()

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			js.EnableInternal(vm, host, &eventloop.EventLoop{}, &dynamic.Config{})
			req, err := require.NewRegistry()
			r.NoError(t, err)
			req.Enable(vm)
			req.RegisterNativeModule("mokapi/smtp", mod.Require)

			tc.run(t, vm, server.Addr)
			tc.test(t, s)
		})
	}
}
