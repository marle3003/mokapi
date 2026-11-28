package mail_test

import (
	"io"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/mail"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type deliverFixture struct {
	handler *mail.Handler
	store   *mail.Store
	events  *events.StoreManager
}

func newDeliverFixture(cfg *mail.Config, emitter common.MailEventEmitter) *deliverFixture {
	logrus.SetOutput(io.Discard)
	cfg.Info.Name = "Mail A"
	sm := events.NewStoreManager(nil)
	sm.SetStore(10, events.NewTraits().WithNamespace("mail"))
	store := mail.NewStore(cfg)
	if emitter == nil {
		emitter = enginetest.NewEngine()
	}
	return &deliverFixture{handler: mail.NewHandler(cfg, store, emitter, sm), store: store, events: sm}
}

func (f *deliverFixture) logs() []*mail.Log {
	var logs []*mail.Log
	for _, e := range f.events.GetEvents(events.NewTraits().WithNamespace("mail")) {
		logs = append(logs, e.Data.(*mail.Log))
	}
	return logs
}

func (f *deliverFixture) inbox(t *testing.T, name string) []*mail.Mail {
	box, ok := f.store.Mailbox(name)
	require.True(t, ok, "mailbox %v", name)
	inbox, ok := box.Folders["INBOX"]
	if !ok {
		return nil
	}
	return inbox.Messages
}

func aMail(subject string) *smtp.Message {
	return &smtp.Message{
		From:    []smtp.Address{{Address: "sender@foo.bar"}},
		To:      []smtp.Address{{Address: "bob@foo.bar"}},
		Subject: subject,
		Date:    time.Now(),
	}
}

func TestDeliver_AppendsToInbox(t *testing.T) {
	f := newDeliverFixture(&mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}}, nil)

	err := f.handler.Deliver("sender@foo.bar", aMail("hello"), []string{"bob@foo.bar", "carol@foo.bar"}, nil)

	require.NoError(t, err)
	require.Len(t, f.inbox(t, "bob@foo.bar"), 1)
	require.Equal(t, "hello", f.inbox(t, "carol@foo.bar")[0].Subject)
}

func TestDeliver_SetsSize(t *testing.T) {
	f := newDeliverFixture(&mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}}, nil)
	msg := aMail("hello")
	msg.Body = "some body"

	require.NoError(t, f.handler.Deliver("sender@foo.bar", msg, []string{"bob@foo.bar"}, nil))

	require.Greater(t, f.inbox(t, "bob@foo.bar")[0].Size, len(msg.Body))
}

func TestDeliver_RuleRejects(t *testing.T) {
	cfg := &mail.Config{
		Settings: &mail.Settings{AutoCreateMailbox: true},
		Rules: map[string]*mail.Rule{
			"subject": {
				Subject:        mail.NewRuleExpr("spam"),
				Action:         mail.Deny,
				RejectResponse: &mail.RejectResponse{StatusCode: 550, Message: "no spam please"},
			},
			"rcpt": {Recipient: mail.NewRuleExpr("^blocked"), Action: mail.Deny},
		},
	}
	f := newDeliverFixture(cfg, nil)

	err := f.handler.Deliver("sender@foo.bar", aMail("spam"), []string{"bob@foo.bar"}, nil)
	require.EqualError(t, err, "no spam please")

	err = f.handler.Deliver("sender@foo.bar", aMail("hi"), []string{"blocked@foo.bar"}, nil)
	require.EqualError(t, err, "Recipient blocked@foo.bar does match deny rule: ^blocked")

	require.Empty(t, f.store.Mailboxes())
}

func TestDeliver_UnknownSender(t *testing.T) {
	cfg := &mail.Config{
		Mailboxes: map[string]*mail.MailboxConfig{"bob@foo.bar": {}},
		Settings:  &mail.Settings{},
	}
	f := newDeliverFixture(cfg, nil)

	err := f.handler.Deliver("nobody@foo.bar", aMail("hi"), []string{"bob@foo.bar"}, nil)

	require.EqualError(t, err, "Unknown mailbox nobody@foo.bar")
	require.Empty(t, f.inbox(t, "bob@foo.bar"))
}

func TestDeliver_MaxRecipients(t *testing.T) {
	f := newDeliverFixture(&mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true, MaxRecipients: 2}}, nil)

	err := f.handler.Deliver("sender@foo.bar", aMail("hi"), []string{"a@foo.bar", "b@foo.bar", "c@foo.bar"}, nil)

	require.EqualError(t, err, "Too many recipients of 2 reached")
	require.Empty(t, f.store.Mailboxes())
}

func TestDeliver_UnknownRecipientWithoutAutoCreate(t *testing.T) {
	cfg := &mail.Config{
		Mailboxes: map[string]*mail.MailboxConfig{"bob@foo.bar": {}},
		Settings:  &mail.Settings{AllowUnknownSenders: true},
	}
	f := newDeliverFixture(cfg, nil)

	err := f.handler.Deliver("sender@foo.bar", aMail("hi"), []string{"nobody@foo.bar"}, nil)

	require.EqualError(t, err, "Unknown mailbox nobody@foo.bar")
}

func TestDeliver_NoMailboxCreatedOnReject(t *testing.T) {
	cfg := &mail.Config{
		Settings: &mail.Settings{AutoCreateMailbox: true},
		Rules:    map[string]*mail.Rule{"rcpt": {Recipient: mail.NewRuleExpr("^blocked"), Action: mail.Deny}},
	}
	f := newDeliverFixture(cfg, nil)

	require.Error(t, f.handler.Deliver("sender@foo.bar", aMail("hi"), []string{"new@foo.bar", "blocked@foo.bar"}, nil))

	_, ok := f.store.Mailbox("new@foo.bar")
	require.False(t, ok, "first recipient is valid, but the mail is rejected as a whole")
}

func TestDeliver_AllOrNothing(t *testing.T) {
	cfg := &mail.Config{
		Mailboxes: map[string]*mail.MailboxConfig{"bob@foo.bar": {}},
		Settings:  &mail.Settings{AllowUnknownSenders: true},
	}
	f := newDeliverFixture(cfg, nil)

	err := f.handler.Deliver("sender@foo.bar", aMail("hi"), []string{"bob@foo.bar", "nobody@foo.bar"}, nil)

	require.Error(t, err)
	require.Empty(t, f.inbox(t, "bob@foo.bar"), "bob is valid, but nobody is not")
}

func TestDeliver_EmitsEventAndWritesBack(t *testing.T) {
	var event *common.MailEvent
	emitter := enginetest.MailEventHandlerFunc(func(e *common.MailEvent) []*common.Action {
		event = e
		e.Message.Subject = "changed by handler"
		return []*common.Action{{Duration: 1}}
	})
	f := newDeliverFixture(&mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}}, emitter)

	err := f.handler.Deliver("sender@foo.bar", aMail("original"), []string{"bob@foo.bar"}, nil)

	require.NoError(t, err)
	require.Equal(t, "Mail A", event.Api)
	require.Equal(t, []string{"bob@foo.bar"}, event.Recipients)
	require.Equal(t, "changed by handler", f.inbox(t, "bob@foo.bar")[0].Subject, "handlers run before the mail is stored")
	require.Len(t, f.logs()[0].GetActions(), 1)
}

func TestDeliver_LogShowsChangedSubject(t *testing.T) {
	emitter := enginetest.MailEventHandlerFunc(func(e *common.MailEvent) []*common.Action {
		e.Message.Subject = "changed by handler"
		return []*common.Action{{Duration: 1}}
	})
	f := newDeliverFixture(&mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}}, emitter)

	err := f.handler.Deliver("sender@foo.bar", aMail("original"), []string{"bob@foo.bar"}, nil)

	require.NoError(t, err)
	require.Equal(t, "changed by handler", f.logs()[0].Subject)
}

func TestDeliver_LogKeepsRawSubjectIfItCannotBeDecoded(t *testing.T) {
	f := newDeliverFixture(&mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}}, nil)

	err := f.handler.Deliver("sender@foo.bar", aMail("=?bogus-charset?q?abc?="), []string{"bob@foo.bar"}, nil)

	require.NoError(t, err)
	require.Equal(t, "=?bogus-charset?q?abc?=", f.logs()[0].Subject)
}

func TestDeliver_WritesLogEvent(t *testing.T) {
	f := newDeliverFixture(&mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}}, nil)

	require.NoError(t, f.handler.Deliver("sender@foo.bar", aMail("hello"), []string{"bob@foo.bar"}, nil))

	logs := f.logs()
	require.Len(t, logs, 1)
	require.Equal(t, "sender@foo.bar", logs[0].From)
	require.Equal(t, []string{"bob@foo.bar"}, logs[0].To)
	require.Equal(t, "hello", logs[0].Subject)
	require.Empty(t, logs[0].Error)
}

func TestDeliver_WritesLogEventOnReject(t *testing.T) {
	f := newDeliverFixture(&mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true, MaxRecipients: 1}}, nil)

	require.Error(t, f.handler.Deliver("sender@foo.bar", aMail("hello"), []string{"a@foo.bar", "b@foo.bar"}, nil))

	logs := f.logs()
	require.Len(t, logs, 1)
	require.Equal(t, "Too many recipients of 1 reached", logs[0].Error)
}

func TestDeliver_CountsMetrics(t *testing.T) {
	f := newDeliverFixture(&mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}}, nil)
	m := monitor.NewMail()

	require.NoError(t, f.handler.Deliver("sender@foo.bar", aMail("hello"), []string{"bob@foo.bar"}, m))

	require.Equal(t, 1.0, m.Mails.WithLabel("Mail A", "sender@foo.bar").Value())
}

// Only SMTP clients log in. A mailbox with credentials must not block a script.
func TestDeliver_NoAuthentication(t *testing.T) {
	cfg := &mail.Config{
		Mailboxes: map[string]*mail.MailboxConfig{"sender@foo.bar": {Username: "foo", Password: "bar"}},
		Settings:  &mail.Settings{AutoCreateMailbox: true},
	}
	f := newDeliverFixture(cfg, nil)

	err := f.handler.Deliver("sender@foo.bar", aMail("hi"), []string{"bob@foo.bar"}, nil)

	require.NoError(t, err)
	require.Len(t, f.inbox(t, "bob@foo.bar"), 1)
}

func TestDeliver_NoRecipientIsRejected(t *testing.T) {
	f := newDeliverFixture(&mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}}, nil)
	m := monitor.NewMail()

	err := f.handler.Deliver("sender@foo.bar", aMail("hi"), nil, m)

	require.EqualError(t, err, "at least one recipient is required")
	require.Empty(t, f.store.Mailboxes())
	require.Equal(t, 0.0, m.Mails.WithLabel("Mail A", "sender@foo.bar").Value())
	require.Equal(t, "at least one recipient is required", f.logs()[0].Error)
}

func TestDeliver_SizeIncludesChangesOfHandler(t *testing.T) {
	emitter := enginetest.MailEventHandlerFunc(func(e *common.MailEvent) []*common.Action {
		e.Message.Body = "a much longer body than the original one"
		return []*common.Action{{Duration: 1}}
	})
	f := newDeliverFixture(&mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}}, emitter)
	msg := aMail("hello")
	msg.Body = "short"
	expected := &smtp.Message{}
	*expected = *msg
	expected.Body = "a much longer body than the original one"
	expected.UpdateSize()

	require.NoError(t, f.handler.Deliver("sender@foo.bar", msg, []string{"bob@foo.bar"}, nil))

	require.Equal(t, expected.Size, f.inbox(t, "bob@foo.bar")[0].Size)
}
