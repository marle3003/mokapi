package mokapi_test

import (
	"fmt"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/smtp"
	"sync"
	"testing"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

type mailHandler struct {
	Filter  common.MailFilter
	Args    common.EventArgs
	Execute common.EventHandlerFunc
}

type mailRecorder struct {
	mu       sync.Mutex
	handlers []mailHandler
	sends    []*common.MailSendArgs
	sendErr  error
	onSend   func(args *common.MailSendArgs)
}

func (rec *mailRecorder) attach(host *enginetest.Host) {
	host.OnMailFunc = func(filter common.MailFilter, do common.EventHandlerFunc, args common.EventArgs) {
		rec.mu.Lock()
		defer rec.mu.Unlock()
		rec.handlers = append(rec.handlers, mailHandler{Filter: filter, Args: args, Execute: do})
	}
	host.MailClientTest = &enginetest.MailClient{SendFunc: func(args *common.MailSendArgs) error {
		rec.mu.Lock()
		rec.sends = append(rec.sends, args)
		err, onSend := rec.sendErr, rec.onSend
		rec.mu.Unlock()
		if onSend != nil {
			onSend(args)
		}
		return err
	}}
}

func (rec *mailRecorder) sent() []*common.MailSendArgs {
	rec.mu.Lock()
	defer rec.mu.Unlock()
	return append([]*common.MailSendArgs(nil), rec.sends...)
}

func TestModule_AppMail_Handlers(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		test   func(t *testing.T, handlers []mailHandler)
	}{
		{
			name:   "message",
			script: `m.app.mail().message((mail) => {})`,
			test: func(t *testing.T, handlers []mailHandler) {
				r.Len(t, handlers, 1)
				r.Equal(t, common.MailFilter{}, handlers[0].Filter)
			},
		},
		{
			name:   "api and message",
			script: `m.app.api('Mail A').mail().message((mail) => {})`,
			test: func(t *testing.T, handlers []mailHandler) {
				r.Len(t, handlers, 1)
				r.Equal(t, common.MailFilter{Api: "Mail A"}, handlers[0].Filter)
			},
		},
		{
			name:   "mailbox",
			script: `m.app.mail().mailbox('alice@example.com').message((mail) => {})`,
			test: func(t *testing.T, handlers []mailHandler) {
				r.Len(t, handlers, 1)
				r.Equal(t, common.MailFilter{Mailbox: "alice@example.com"}, handlers[0].Filter)
			},
		},
		{
			name:   "api and mailbox",
			script: `m.app.api('Mail A').mail().mailbox('alice@example.com').message((mail) => {})`,
			test: func(t *testing.T, handlers []mailHandler) {
				r.Equal(t, common.MailFilter{Api: "Mail A", Mailbox: "alice@example.com"}, handlers[0].Filter)
			},
		},
		{
			name: "chained routers keep their own filter",
			script: `
const mail = m.app.mail()
mail.mailbox('alice@example.com').message((x) => {}).message((x) => {})
mail.message((x) => {})`,
			test: func(t *testing.T, handlers []mailHandler) {
				r.Len(t, handlers, 3)
				r.Equal(t, "alice@example.com", handlers[1].Filter.Mailbox)
				r.Equal(t, "", handlers[2].Filter.Mailbox)
			},
		},
		{
			name:   "priority and tags",
			script: `m.app.mail().message((mail) => {}, { priority: 5, tags: { name: 'foo' } })`,
			test: func(t *testing.T, handlers []mailHandler) {
				r.Equal(t, common.EventArgs{Priority: 5, Tags: map[string]string{"name": "foo"}}, handlers[0].Args)
			},
		},
		{
			name:   "mailbox handler gets options",
			script: `m.app.mail().mailbox('a@b.io').message((mail) => {}, { tags: { name: 'alice' } })`,
			test: func(t *testing.T, handlers []mailHandler) {
				r.Equal(t, map[string]string{"name": "alice"}, handlers[0].Args.Tags)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			loop, host := newModuleLoop(t)
			rec := &mailRecorder{}
			rec.attach(host)

			_, err := evalOnLoop(loop, "const m = require('mokapi')\n"+tc.script)

			r.NoError(t, err)
			rec.mu.Lock()
			defer rec.mu.Unlock()
			tc.test(t, rec.handlers)
		})
	}
}

func TestModule_AppMail_InvalidOptionsThrow(t *testing.T) {
	loop, host := newModuleLoop(t)
	(&mailRecorder{}).attach(host)

	_, err := evalOnLoop(loop, `require('mokapi').app.mail().message((m) => {}, 'foo')`)

	r.ErrorContains(t, err, "unexpected type for args")
}

func TestModule_AppMail_HandlerGetsMailWithApiAndDate(t *testing.T) {
	loop, host := newModuleLoop(t)
	rec := &mailRecorder{}
	rec.attach(host)
	_, err := evalOnLoop(loop, `
const m = require('mokapi')
let seen
m.app.mail().message((mail) => { seen = { api: mail.api, subject: mail.subject, date: mail.time instanceof Date } })
`)
	r.NoError(t, err)

	rec.mu.Lock()
	do := rec.handlers[0].Execute
	rec.mu.Unlock()
	_, err = do(&common.EventContext{Args: []any{common.NewMailEventMessage("Mail A", &smtp.Message{Subject: "Hello"})}})
	r.NoError(t, err)

	v, err := evalOnLoop(loop, "seen")
	r.NoError(t, err)
	r.Equal(t, map[string]any{"api": "Mail A", "subject": "Hello", "date": true}, v)
}

func TestModule_AppMail_Send(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		test   func(t *testing.T, sends []*common.MailSendArgs)
	}{
		{
			name: "scoped send passes api and converted message",
			script: `m.app.api('Mail A').mail().send({
  from: [{ name: 'Bob', address: 'bob@example.com' }],
  to: ['alice@example.com'],
  cc: [{ address: 'cc@example.com' }],
  subject: 'Hello',
  body: 'Hi Alice',
})`,
			test: func(t *testing.T, sends []*common.MailSendArgs) {
				r.Len(t, sends, 1)
				r.Equal(t, "Mail A", sends[0].Api)
				msg := sends[0].Message
				r.Equal(t, []smtp.Address{{Name: "Bob", Address: "bob@example.com"}}, msg.From)
				r.Equal(t, []smtp.Address{{Address: "alice@example.com"}}, msg.To)
				r.Equal(t, []smtp.Address{{Address: "cc@example.com"}}, msg.Cc)
				r.Equal(t, "Hello", msg.Subject)
				r.Equal(t, "Hi Alice", msg.Body)
				r.NotEmpty(t, msg.MessageId)
				r.False(t, msg.Date.IsZero())
			},
		},
		{
			name:   "unscoped send has no api",
			script: `m.app.mail().send({ from: 'bob@example.com', to: 'alice@example.com' })`,
			test: func(t *testing.T, sends []*common.MailSendArgs) {
				r.Equal(t, "", sends[0].Api)
			},
		},
		{
			name: "time and messageId are kept",
			script: `m.app.mail().send({ from: 'bob@example.com', to: 'alice@example.com',
  messageId: '<1@example.com>', time: new Date('2020-01-02T03:04:05Z') })`,
			test: func(t *testing.T, sends []*common.MailSendArgs) {
				r.Equal(t, "<1@example.com>", sends[0].Message.MessageId)
				r.Equal(t, int64(1577934245), sends[0].Message.Date.Unix())
			},
		},
		{
			name:   "send returns the router",
			script: `const router = m.app.mail(); if (router.send({ from: 'a@b.io', to: 'c@d.io' }) !== router) throw new Error('not the router')`,
			test: func(t *testing.T, sends []*common.MailSendArgs) {
				r.Len(t, sends, 1)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			loop, host := newModuleLoop(t)
			rec := &mailRecorder{}
			rec.attach(host)

			_, err := evalOnLoop(loop, "const m = require('mokapi')\n"+tc.script)

			r.NoError(t, err)
			tc.test(t, rec.sent())
		})
	}
}

func TestModule_AppMail_SendThrows(t *testing.T) {
	testcases := []struct {
		name    string
		script  string
		sendErr error
		thrown  string
	}{
		{
			name:    "client error",
			script:  `m.app.mail().send({ from: 'a@b.io', to: 'c@d.io' })`,
			sendErr: fmt.Errorf("mail API 'X' not found"),
			thrown:  "mail API 'X' not found",
		},
		{
			name:   "several from without sender",
			script: `m.app.mail().send({ from: ['a@b.io', 'e@f.io'], to: 'c@d.io' })`,
			thrown: "sender required if using multiple from addresses",
		},
		{
			name:   "invalid address",
			script: `m.app.mail().send({ from: 'a@b.io', to: [{ name: 'no address' }] })`,
			thrown: "expected address field",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			loop, host := newModuleLoop(t)
			rec := &mailRecorder{sendErr: tc.sendErr}
			rec.attach(host)

			_, err := evalOnLoop(loop, "const m = require('mokapi')\n"+tc.script)

			r.ErrorContains(t, err, tc.thrown)
			if tc.sendErr == nil {
				r.Empty(t, rec.sent(), "invalid mails never reach the client")
			}
		})
	}
}

func TestModule_AppMail_SendAsync(t *testing.T) {
	t.Run("resolves with the router", func(t *testing.T) {
		loop, host := newModuleLoop(t)
		rec := &mailRecorder{}
		rec.attach(host)

		_, err := evalOnLoop(loop, `
const m = require('mokapi')
const router = m.app.api('Mail A').mail()
const p = router.sendAsync({ from: 'a@b.io', to: 'c@d.io', subject: 'async' })
const same = p.then((x) => x === router)
`)
		r.NoError(t, err)

		state, result := awaitPromiseResult(t, loop, "same")
		r.Equal(t, goja.PromiseStateFulfilled, state)
		r.Equal(t, true, result)
		sends := rec.sent()
		r.Len(t, sends, 1)
		r.Equal(t, "Mail A", sends[0].Api)
		r.Equal(t, "async", sends[0].Message.Subject)
	})

	t.Run("rejects with the client error", func(t *testing.T) {
		loop, host := newModuleLoop(t)
		(&mailRecorder{sendErr: fmt.Errorf("rejected by rule")}).attach(host)
		host.MailClientTest.SendFunc = func(*common.MailSendArgs) error { return fmt.Errorf("rejected by rule") }

		_, err := evalOnLoop(loop, `
const m = require('mokapi')
const p = m.app.mail().sendAsync({ from: 'a@b.io', to: 'c@d.io' })
`)
		r.NoError(t, err)

		state, reason := awaitPromiseResult(t, loop, "p")
		r.Equal(t, goja.PromiseStateRejected, state)
		r.Equal(t, "rejected by rule", reason)
	})

	t.Run("invalid message throws synchronously", func(t *testing.T) {
		loop, host := newModuleLoop(t)
		(&mailRecorder{}).attach(host)

		_, err := evalOnLoop(loop, `require('mokapi').app.mail().sendAsync({ from: ['a@b.io', 'e@f.io'], to: 'c@d.io' })`)

		r.ErrorContains(t, err, "sender required")
	})

	t.Run("sends the mail as it was at call time", func(t *testing.T) {
		loop, host := newModuleLoop(t)
		rec := &mailRecorder{}
		rec.attach(host)

		_, err := evalOnLoop(loop, `
const m = require('mokapi')
const msg = { from: 'a@b.io', to: 'c@d.io', subject: 'before' }
const p = m.app.mail().sendAsync(msg)
msg.subject = 'after'
`)
		r.NoError(t, err)
		r.Equal(t, goja.PromiseStateFulfilled, awaitPromise(t, loop, "p"))

		r.Equal(t, "before", rec.sent()[0].Message.Subject)
	})
}

// A handler of the same script that sends a mail while that script is delivering one: the
// event loop runs the handler inline, so send() must not deadlock.
func TestModule_AppMail_HandlerCallsSend(t *testing.T) {
	loop, host := newModuleLoop(t)
	rec := &mailRecorder{}
	rec.attach(host)
	// the delivery of the host runs the handlers of the script
	rec.onSend = func(args *common.MailSendArgs) {
		rec.mu.Lock()
		handlers := append([]mailHandler(nil), rec.handlers...)
		rec.mu.Unlock()
		for _, h := range handlers {
			_, err := h.Execute(&common.EventContext{Args: []any{common.NewMailEventMessage("Mail A", args.Message)}})
			r.NoError(t, err)
		}
	}

	_, err := evalOnLoop(loop, `
const m = require('mokapi')
m.app.mail().message((mail) => {
  if (mail.subject === 'first') {
    m.app.mail().send({ from: 'a@b.io', to: 'c@d.io', subject: 'reply' })
  }
})
m.app.mail().send({ from: 'a@b.io', to: 'c@d.io', subject: 'first' })
`)

	r.NoError(t, err)
	sends := rec.sent()
	r.Len(t, sends, 2)
	r.Equal(t, "first", sends[0].Message.Subject)
	r.Equal(t, "reply", sends[1].Message.Subject)
}
