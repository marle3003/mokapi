package engine_test

import (
	"io"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/smtp"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func newMailApi(title string, autoCreate bool, mailboxes ...string) *mail.Config {
	cfg := &mail.Config{
		Info:     mail.Info{Name: title},
		Settings: &mail.Settings{AutoCreateMailbox: autoCreate, AllowUnknownSenders: true},
	}
	if len(mailboxes) > 0 {
		cfg.Mailboxes = map[string]*mail.MailboxConfig{}
	}
	for _, m := range mailboxes {
		cfg.Mailboxes[m] = &mail.MailboxConfig{}
	}
	return cfg
}

func newMessage(rcpt ...string) *smtp.Message {
	msg := &smtp.Message{
		From:    []smtp.Address{{Address: "alice@example.com"}},
		Subject: "Hello",
		Body:    "Hi",
	}
	for _, r := range rcpt {
		msg.To = append(msg.To, smtp.Address{Address: r})
	}
	return msg
}

func inboxSize(t *testing.T, app *runtime.App, api, mailbox string) int {
	box, ok := app.Mail.Get(api).Store.Mailbox(mailbox)
	if !ok {
		return -1
	}
	inbox, ok := box.Folders["INBOX"]
	if !ok {
		return 0
	}
	return len(inbox.Messages)
}

func TestMailClient_Send(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		apis []*mail.Config
		args *common.MailSendArgs
		test func(t *testing.T, app *runtime.App, err error)
	}{
		{
			name: "scoped delivers into the api",
			apis: []*mail.Config{newMailApi("A", true), newMailApi("B", true)},
			args: &common.MailSendArgs{Api: "B", Message: newMessage("bob@example.com")},
			test: func(t *testing.T, app *runtime.App, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, inboxSize(t, app, "B", "bob@example.com"))
				require.Equal(t, -1, inboxSize(t, app, "A", "bob@example.com"))
			},
		},
		{
			name: "scoped unknown title",
			apis: []*mail.Config{newMailApi("A", true)},
			args: &common.MailSendArgs{Api: "X", Message: newMessage("bob@example.com")},
			test: func(t *testing.T, app *runtime.App, err error) {
				require.EqualError(t, err, "mail API 'X' not found")
			},
		},
		{
			name: "unscoped existing mailbox wins over the autoCreate api",
			apis: []*mail.Config{newMailApi("A", true), newMailApi("B", false, "bob@example.com")},
			args: &common.MailSendArgs{Message: newMessage("bob@example.com")},
			test: func(t *testing.T, app *runtime.App, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, inboxSize(t, app, "B", "bob@example.com"))
				require.Equal(t, -1, inboxSize(t, app, "A", "bob@example.com"))
			},
		},
		{
			name: "unscoped without a mailbox falls back to the autoCreate api",
			apis: []*mail.Config{newMailApi("A", true), newMailApi("B", false, "bob@example.com")},
			args: &common.MailSendArgs{Message: newMessage("new@example.com")},
			test: func(t *testing.T, app *runtime.App, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, inboxSize(t, app, "A", "new@example.com"))
			},
		},
		{
			name: "unscoped api with a mailbox in cc and bcc is chosen over the autoCreate api",
			apis: []*mail.Config{newMailApi("A", true), newMailApi("B", false, "bob@example.com", "carol@example.com")},
			args: func() *common.MailSendArgs {
				msg := newMessage("new@example.com")
				msg.Cc = []smtp.Address{{Address: "bob@example.com"}}
				msg.Bcc = []smtp.Address{{Address: "carol@example.com"}}
				return &common.MailSendArgs{Message: msg}
			}(),
			test: func(t *testing.T, app *runtime.App, err error) {
				// new@ is rejected by B (no autoCreate), but B is chosen over A
				require.Error(t, err)
				require.Contains(t, err.Error(), "Unknown mailbox new@example.com")
			},
		},
		{
			name: "unscoped two apis create mailboxes",
			apis: []*mail.Config{newMailApi("A", true), newMailApi("D", true)},
			args: &common.MailSendArgs{Message: newMessage("bob@example.com")},
			test: func(t *testing.T, app *runtime.App, err error) {
				require.EqualError(t, err, "ambiguous mail API: 2 create mailboxes automatically ('A', 'D'). Use app.api('...').mail() to choose the mail API")
				require.Equal(t, -1, inboxSize(t, app, "A", "bob@example.com"))
			},
		},
		{
			name: "unscoped recipients with mailboxes in two apis",
			apis: []*mail.Config{newMailApi("A", false, "bob@example.com"), newMailApi("B", false, "carol@example.com")},
			args: &common.MailSendArgs{Message: newMessage("bob@example.com", "carol@example.com")},
			test: func(t *testing.T, app *runtime.App, err error) {
				require.EqualError(t, err, "ambiguous mail API: recipients have mailboxes in 'A', 'B'. Use app.api('...').mail() to choose the mail API")
			},
		},
		{
			name: "unscoped no mailbox and no api creates mailboxes",
			apis: []*mail.Config{newMailApi("A", false, "bob@example.com"), newMailApi("B", false, "carol@example.com")},
			args: &common.MailSendArgs{Message: newMessage("new@example.com")},
			test: func(t *testing.T, app *runtime.App, err error) {
				require.EqualError(t, err, "no mail API has a mailbox for new@example.com and none creates mailboxes automatically ('A', 'B'). Use app.api('...').mail() to choose the mail API")
			},
		},
		{
			name: "unscoped without any mail api",
			args: &common.MailSendArgs{Message: newMessage("bob@example.com")},
			test: func(t *testing.T, app *runtime.App, err error) {
				require.EqualError(t, err, "no mail API found")
			},
		},
		{
			name: "settings missing count as autoCreateMailbox",
			apis: []*mail.Config{{Info: mail.Info{Name: "A"}}},
			args: &common.MailSendArgs{Message: newMessage("bob@example.com")},
			test: func(t *testing.T, app *runtime.App, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, inboxSize(t, app, "A", "bob@example.com"))
			},
		},
		{
			name: "recipient in to and cc gets one copy",
			apis: []*mail.Config{newMailApi("A", true)},
			args: func() *common.MailSendArgs {
				msg := newMessage("bob@example.com")
				msg.Cc = []smtp.Address{{Address: "bob@example.com"}}
				return &common.MailSendArgs{Api: "A", Message: msg}
			}(),
			test: func(t *testing.T, app *runtime.App, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, inboxSize(t, app, "A", "bob@example.com"))
			},
		},
		{
			name: "sender header is the envelope sender",
			apis: []*mail.Config{newMailApi("A", true)},
			args: func() *common.MailSendArgs {
				msg := newMessage("bob@example.com")
				msg.From = append(msg.From, smtp.Address{Address: "carol@example.com"})
				msg.Sender = &smtp.Address{Address: "carol@example.com"}
				return &common.MailSendArgs{Api: "A", Message: msg}
			}(),
			test: func(t *testing.T, app *runtime.App, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "several from without sender",
			apis: []*mail.Config{newMailApi("A", true)},
			args: func() *common.MailSendArgs {
				msg := newMessage("bob@example.com")
				msg.From = append(msg.From, smtp.Address{Address: "carol@example.com"})
				return &common.MailSendArgs{Api: "A", Message: msg}
			}(),
			test: func(t *testing.T, app *runtime.App, err error) {
				require.EqualError(t, err, "sender required if using multiple from addresses")
			},
		},
		{
			name: "no from and no sender",
			apis: []*mail.Config{newMailApi("A", true)},
			args: func() *common.MailSendArgs {
				msg := newMessage("bob@example.com")
				msg.From = nil
				return &common.MailSendArgs{Api: "A", Message: msg}
			}(),
			test: func(t *testing.T, app *runtime.App, err error) {
				require.EqualError(t, err, "a sender or from address must be specified")
			},
		},
		{
			name: "no recipient without api scope",
			apis: []*mail.Config{newMailApi("A", true), newMailApi("B", true)},
			args: func() *common.MailSendArgs {
				msg := newMessage()
				return &common.MailSendArgs{Message: msg}
			}(),
			test: func(t *testing.T, app *runtime.App, err error) {
				require.EqualError(t, err, "at least one recipient is required")
			},
		},
		{
			name: "delivery rejection is returned",
			apis: []*mail.Config{{
				Info:     mail.Info{Name: "A"},
				Settings: &mail.Settings{AutoCreateMailbox: true, MaxRecipients: 1},
			}},
			args: &common.MailSendArgs{Api: "A", Message: newMessage("bob@example.com", "carol@example.com")},
			test: func(t *testing.T, app *runtime.App, err error) {
				require.EqualError(t, err, "Too many recipients of 1 reached")
				require.Equal(t, -1, inboxSize(t, app, "A", "bob@example.com"))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app := runtimetest.NewApp(runtimetest.WithMail(tc.apis...))
			client := engine.NewMailClient(app, enginetest.NewEngine())

			err := client.Send(tc.args)

			tc.test(t, app, err)
		})
	}
}

func TestMailClient_Send_EmitsEventWithApiAndEnvelope(t *testing.T) {
	logrus.SetOutput(io.Discard)
	app := runtimetest.NewApp(runtimetest.WithMail(newMailApi("A", true)))
	var got *common.MailEvent
	emitter := enginetest.MailEventHandlerFunc(func(event *common.MailEvent) []*common.Action {
		got = event
		return nil
	})

	msg := newMessage("bob@example.com")
	msg.Bcc = []smtp.Address{{Address: "carol@example.com"}}
	err := engine.NewMailClient(app, emitter).Send(&common.MailSendArgs{Message: msg})

	require.NoError(t, err)
	require.Equal(t, "A", got.Api)
	require.Equal(t, []string{"bob@example.com", "carol@example.com"}, got.Recipients)
}

func TestMailClient_Send_ConfigRemovedWhileScriptIsRunning(t *testing.T) {
	logrus.SetOutput(io.Discard)

	// a mail API whose last file was removed keeps its store but has no config snapshot (nil)
	newApp := func() *runtime.App {
		removed := &runtime.MailInfo{Store: mail.NewStore(&mail.Config{Info: mail.Info{Name: "A"}})}
		return runtimetest.NewApp(runtimetest.WithMailInfo("A", removed))
	}

	t.Run("scoped returns not found", func(t *testing.T) {
		require.NotPanics(t, func() {
			err := engine.NewMailClient(newApp(), enginetest.NewEngine()).
				Send(&common.MailSendArgs{Api: "A", Message: newMessage("bob@example.com")})

			require.EqualError(t, err, "mail API 'A' not found")
		})
	})

	t.Run("unscoped skips it", func(t *testing.T) {
		require.NotPanics(t, func() {
			err := engine.NewMailClient(newApp(), enginetest.NewEngine()).
				Send(&common.MailSendArgs{Message: newMessage("bob@example.com")})

			require.EqualError(t, err, "no mail API found")
		})
	})
}
