package mail_test

import (
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	mod "mokapi/js/mail"
	"mokapi/js/require"
	"mokapi/providers/mail"
	"mokapi/runtime/events/eventstest"
	"mokapi/smtp"
	"mokapi/smtp/smtptest"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/sirupsen/logrus"
	r "github.com/stretchr/testify/require"
)

func TestMailModule(t *testing.T) {
	logrus.SetOutput(io.Discard)

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
				bob, _ := store.Mailbox("bob@mokapi.io")
				msg := bob.Folders["INBOX"].Messages[0]
				r.Equal(t, "A test mail", msg.Subject)
				r.Equal(t, "Hello Bob", msg.Body)
				r.Equal(t, "Alice", msg.From[0].Name)
				r.Equal(t, "alice@mokapi.io", msg.From[0].Address)
			},
		},
		{
			name: "Send mail with auth login",
			cfg: &mail.Config{Mailboxes: map[string]*mail.MailboxConfig{
				"alice@mokapi.io": {
					Username: "alice",
					Password: "secret",
				},
				"bob@mokapi.io": {},
			}},
			run: func(t *testing.T, vm *goja.Runtime, addr string) {
				_, err := vm.RunString(fmt.Sprintf(`const m = require('mokapi/smtp')
m.send('smtp://%v', {from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'},
{ login: { username: 'alice', password: 'secret' } })
`, addr))
				r.NoError(t, err)
			},
			test: func(t *testing.T, store *mail.Store) {
				bob, _ := store.Mailbox("bob@mokapi.io")
				msg := bob.Folders["INBOX"].Messages[0]
				r.Equal(t, "A test mail", msg.Subject)
				r.Equal(t, "Hello Bob", msg.Body)
				r.Equal(t, "Alice", msg.From[0].Name)
				r.Equal(t, "alice@mokapi.io", msg.From[0].Address)
			},
		},
		{
			name: "Send mail with auth plain",
			cfg: &mail.Config{Mailboxes: map[string]*mail.MailboxConfig{
				"alice@mokapi.io": {
					Username: "alice",
					Password: "secret",
				},
				"bob@mokapi.io": {},
			}},
			run: func(t *testing.T, vm *goja.Runtime, addr string) {
				_, err := vm.RunString(fmt.Sprintf(`const m = require('mokapi/smtp')
m.send('smtp://%v', {from: {name: 'Alice', address: 'alice@mokapi.io'}, to: ['bob@mokapi.io'], subject: 'A test mail', body: 'Hello Bob'},
{ plain: { username: 'alice', password: 'secret' } })
`, addr))
				r.NoError(t, err)
			},
			test: func(t *testing.T, store *mail.Store) {
				bob, _ := store.Mailbox("bob@mokapi.io")
				msg := bob.Folders["INBOX"].Messages[0]
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
				cfg = &mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}}
			}
			s := mail.NewStore(cfg)
			h := mail.NewHandler(cfg, s, enginetest.NewEngine(), &eventstest.Handler{})
			server, _, err := smtptest.NewServer(h.ServeSMTP)
			r.NoError(t, err)
			defer server.Close()

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			js.EnableInternal(vm, host, &eventloop.EventLoop{}, &dynamic.Config{})
			req, err := require.NewRegistry(nil)
			r.NoError(t, err)
			req.Enable(vm)
			req.RegisterNativeModule("mokapi/smtp", mod.Require)

			tc.run(t, vm, server.Addr)
			tc.test(t, s)
		})
	}
}

func TestMail_ToMessage(t *testing.T) {
	testcases := []struct {
		name string
		mail *mod.Mail
		host *enginetest.Host
		test func(t *testing.T, msg *smtp.Message, err error)
	}{
		{
			name: "fields are converted",
			mail: &mod.Mail{
				From:        []interface{}{map[string]interface{}{"name": "Alice", "address": "alice@mokapi.io"}},
				To:          "bob@mokapi.io",
				ReplyTo:     []interface{}{"reply@mokapi.io"},
				InReplyTo:   "<1@mokapi.io>",
				Subject:     "Hello",
				Body:        "Hi",
				ContentType: "text/html",
				Encoding:    "7bit",
			},
			test: func(t *testing.T, msg *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, []smtp.Address{{Name: "Alice", Address: "alice@mokapi.io"}}, msg.From)
				r.Equal(t, []smtp.Address{{Address: "bob@mokapi.io"}}, msg.To)
				r.Equal(t, []smtp.Address{{Address: "reply@mokapi.io"}}, msg.ReplyTo)
				r.Equal(t, "<1@mokapi.io>", msg.InReplyTo)
				r.Equal(t, "Hello", msg.Subject)
				r.Equal(t, "Hi", msg.Body)
				r.Equal(t, "text/html", msg.ContentType)
				r.Equal(t, "7bit", msg.ContentTransferEncoding)
				r.Nil(t, msg.Sender)
			},
		},
		{
			name: "messageId and time are generated",
			mail: &mod.Mail{From: "alice@mokapi.io", To: "bob@mokapi.io"},
			test: func(t *testing.T, msg *smtp.Message, err error) {
				r.NoError(t, err)
				r.NotEmpty(t, msg.MessageId)
				r.WithinDuration(t, time.Now(), msg.Date, time.Minute)
			},
		},
		{
			name: "messageId and time are kept",
			mail: &mod.Mail{From: "alice@mokapi.io", To: "bob@mokapi.io", MessageId: "<id@mokapi.io>", Time: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)},
			test: func(t *testing.T, msg *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, "<id@mokapi.io>", msg.MessageId)
				r.Equal(t, time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC), msg.Date)
			},
		},
		{
			name: "cc and bcc",
			mail: &mod.Mail{From: "alice@mokapi.io", To: "bob@mokapi.io", Cc: []interface{}{"cc@mokapi.io"}, Bcc: []interface{}{"bcc@mokapi.io"}},
			test: func(t *testing.T, msg *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, []smtp.Address{{Address: "cc@mokapi.io"}}, msg.Cc)
				r.Equal(t, []smtp.Address{{Address: "bcc@mokapi.io"}}, msg.Bcc)
			},
		},
		{
			name: "sender as address object",
			mail: &mod.Mail{
				From:   []interface{}{"alice@mokapi.io", "carol@mokapi.io"},
				Sender: map[string]interface{}{"name": "Carol", "address": "carol@mokapi.io"},
				To:     "bob@mokapi.io",
			},
			test: func(t *testing.T, msg *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, &smtp.Address{Name: "Carol", Address: "carol@mokapi.io"}, msg.Sender)
			},
		},
		{
			name: "sender as string",
			mail: &mod.Mail{From: []interface{}{"alice@mokapi.io", "carol@mokapi.io"}, Sender: "carol@mokapi.io", To: "bob@mokapi.io"},
			test: func(t *testing.T, msg *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, "carol@mokapi.io", msg.Sender.Address)
			},
		},
		{
			name: "several from without sender",
			mail: &mod.Mail{From: []interface{}{"alice@mokapi.io", "carol@mokapi.io"}, To: "bob@mokapi.io"},
			test: func(t *testing.T, msg *smtp.Message, err error) {
				r.EqualError(t, err, "sender required if using multiple from addresses")
			},
		},
		{
			name: "invalid address",
			mail: &mod.Mail{From: "alice@mokapi.io", To: []interface{}{map[string]interface{}{"name": "Bob"}}},
			test: func(t *testing.T, msg *smtp.Message, err error) {
				r.Error(t, err)
			},
		},
		{
			name: "attachment with data",
			mail: &mod.Mail{From: "alice@mokapi.io", To: "bob@mokapi.io", Attachments: []*mod.Attachment{
				{Name: "a.txt", Data: "hello", ContentType: "text/plain", Disposition: "attachment"},
			}},
			test: func(t *testing.T, msg *smtp.Message, err error) {
				r.NoError(t, err)
				r.Len(t, msg.Attachments, 1)
				r.Equal(t, "a.txt", msg.Attachments[0].Name)
				r.Equal(t, "text/plain", msg.Attachments[0].ContentType)
				r.Equal(t, "attachment", msg.Attachments[0].Disposition)
				r.Equal(t, []byte("hello"), msg.Attachments[0].Data)
			},
		},
		{
			name: "attachment with path",
			host: &enginetest.Host{OpenFunc: func(file, hint string) (*dynamic.Config, error) {
				return &dynamic.Config{Raw: []byte("file content")}, nil
			}},
			mail: &mod.Mail{From: "alice@mokapi.io", To: "bob@mokapi.io", Attachments: []*mod.Attachment{{Path: "dir/report.txt"}}},
			test: func(t *testing.T, msg *smtp.Message, err error) {
				r.NoError(t, err)
				r.Equal(t, "report.txt", msg.Attachments[0].Name)
				r.Equal(t, []byte("file content"), msg.Attachments[0].Data)
				r.Contains(t, msg.Attachments[0].ContentType, "text/plain")
			},
		},
		{
			name: "attachment file not found",
			mail: &mod.Mail{From: "alice@mokapi.io", To: "bob@mokapi.io", Attachments: []*mod.Attachment{{Path: "missing.txt"}}},
			test: func(t *testing.T, msg *smtp.Message, err error) {
				r.Error(t, err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			host := tc.host
			if host == nil {
				host = &enginetest.Host{}
			}
			msg, err := tc.mail.ToMessage(host)
			tc.test(t, msg, err)
		})
	}
}

func TestMail_ExportedFromScript(t *testing.T) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	v, err := vm.RunString(`({from: 'a@b.io', to: ['c@d.io'], sender: {address: 'a@b.io'}, time: new Date(0), subject: 's'})`)
	r.NoError(t, err)
	var m mod.Mail
	r.NoError(t, vm.ExportTo(v, &m))

	msg, err := m.ToMessage(&enginetest.Host{})

	r.NoError(t, err)
	r.Equal(t, "a@b.io", msg.Sender.Address)
	r.True(t, msg.Date.Equal(time.Unix(0, 0)))
}
