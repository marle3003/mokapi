package mail_test

import (
	"context"
	"encoding/base64"
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/providers/mail"
	"mokapi/runtime/events/eventstest"
	"mokapi/smtp"
	"mokapi/smtp/smtptest"
	"testing"
	"time"
)

func TestHandler_ServeSMTP(t *testing.T) {
	testcases := []struct {
		name   string
		config *mail.Config
		test   func(t *testing.T, h *mail.Handler, s *mail.Store, eh *eventstest.Handler)
	}{
		{
			name:   "no auth required",
			config: &mail.Config{},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendMail(t, h, ctx)
				require.Equal(t, smtp.Ok, r.Result)
			},
		},
		{
			name: "auth required without login",
			config: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@foo.bar": {
						Username: "foo",
						Password: "bar",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendMail(t, h, ctx)
				require.Equal(t, &smtp.AuthRequired, r.Result)
			},
		},
		{
			name: "auth invalid credentials",
			config: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@foo.bar": {Username: "alice", Password: "foo"},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendLogin(t, h, ctx, "foo", "foo")
				require.Equal(t, &smtp.InvalidAuthCredentials, r.Result)
			},
		},
		{
			name: "auth valid",
			config: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@foo.bar": {Username: "alice", Password: "foo"},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendLogin(t, h, ctx, "alice", "foo")
				require.Equal(t, &smtp.InvalidAuthCredentials, r.Result)
			},
		},
		{
			name: "mail invalid mailbox",
			config: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"bob@foo.bar": {},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendMail(t, h, ctx)
				exp := smtp.AddressRejected
				exp.Message = "Unknown mailbox alice@foo.bar"
				require.Equal(t, &exp, r.Result)
			},
		},
		{
			name: "mail valid mailbox",
			config: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@foo.bar": {},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendMail(t, h, ctx)
				require.Equal(t, smtp.Ok, r.Result)
			},
		},
		{
			name:   "mail any is valid",
			config: &mail.Config{},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendMail(t, h, ctx)
				require.Equal(t, smtp.Ok, r.Result)
			},
		},
		{
			name: "rcpt invalid mailbox",
			config: &mail.Config{
				Settings: &mail.Settings{AutoCreateMailbox: false},
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@foo.bar": {},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendRcpt(t, h, ctx)
				exp := smtp.AddressRejected
				exp.Message = "Unknown mailbox bob@foo.bar"
				require.Equal(t, &exp, r.Result)
			},
		},
		{
			name: "rcpt valid mailbox",
			config: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@foo.bar": {},
					"bob@foo.bar":   {},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendRcpt(t, h, ctx)
				require.Equal(t, smtp.Ok, r.Result)
			},
		},
		{
			name:   "rcpt any is valid",
			config: &mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendRcpt(t, h, ctx)
				require.Equal(t, smtp.Ok, r.Result)
			},
		},
		{
			name:   "max recipients valid",
			config: &mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true, MaxRecipients: 5}},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendRcpt(t, h, ctx)
				require.Equal(t, smtp.Ok, r.Result)
				r = sendThisRcpt(t, h, ctx, "carol@foo.bar")
				require.Equal(t, smtp.Ok, r.Result)
				r = sendThisRcpt(t, h, ctx, "charlie@foo.bar")
				require.Equal(t, smtp.Ok, r.Result)
			},
		},
		{
			name:   "max recipients not valid",
			config: &mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true, MaxRecipients: 2}},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendRcpt(t, h, ctx)
				require.Equal(t, smtp.Ok, r.Result)
				r = sendThisRcpt(t, h, ctx, "carol@foo.bar")
				require.Equal(t, smtp.Ok, r.Result)
				r = sendThisRcpt(t, h, ctx, "charlie@foo.bar")
				exp := smtp.TooManyRecipients
				exp.Message = "Too many recipients of 2 reached"
				require.Equal(t, &exp, r.Result)
			},
		},
		{
			name:   "data",
			config: &mail.Config{Info: mail.Info{Name: "Testing Mail Server"}, Settings: &mail.Settings{AutoCreateMailbox: true}},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, eh *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendData(t, h, ctx)
				require.Equal(t, smtp.Ok, r.Result)
				require.Len(t, eh.Events, 1)
				require.Equal(t, "namespace=mail, name=Testing Mail Server", eh.Events[0].Traits.String())
			},
		},
		{
			name:   "server should add message into mailbox",
			config: &mail.Config{Settings: &mail.Settings{AutoCreateMailbox: true}},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendData(t, h, ctx)
				require.Equal(t, smtp.Ok, r.Result)
				require.Contains(t, s.Mailboxes, "bob@foo.bar")
				box := s.Mailboxes["bob@foo.bar"]
				require.Len(t, box.Folders["INBOX"].Messages, 1)
			},
		},
		{
			name: "data with allow rule not match sender",
			config: &mail.Config{Rules: map[string]*mail.Rule{
				"foo": {
					Sender: mail.NewRuleExpr(".*@mokapi.io"),
					Action: mail.Allow,
				}}},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendMail(t, h, ctx)
				require.Equal(t, "sender alice@foo.bar does not match allow rule: .*@mokapi.io", r.Result.Message)
			},
		},
		{
			name: "data with deny rule",
			config: &mail.Config{Rules: map[string]*mail.Rule{
				"foo": {
					Sender: mail.NewRuleExpr("@foo.bar"),
					Action: mail.Deny,
				}}},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendMail(t, h, ctx)
				require.Equal(t, "sender alice@foo.bar does match deny rule: @foo.bar", r.Result.Message)
				require.Equal(t, smtp.StatusCode(550), r.Result.StatusCode)
				require.Equal(t, smtp.EnhancedStatusCode{5, 1, 0}, r.Result.EnhancedStatusCode)
			},
		},
		{
			name: "data with deny rule custom response",
			config: &mail.Config{Rules: map[string]*mail.Rule{
				"foo": {
					Sender: mail.NewRuleExpr(".*@foo.bar"),
					Action: mail.Deny,
					RejectResponse: &mail.RejectResponse{
						StatusCode:         500,
						EnhancedStatusCode: smtp.EnhancedStatusCode{5, 1, 2},
						Message:            "custom error message",
					},
				}}},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, _ *eventstest.Handler) {
				ctx := smtp.NewClientContext(context.Background(), "")
				r := sendMail(t, h, ctx)
				require.Equal(t, "custom error message", r.Result.Message)
				require.Equal(t, smtp.StatusCode(500), r.Result.StatusCode)
				require.Equal(t, smtp.EnhancedStatusCode{5, 1, 2}, r.Result.EnhancedStatusCode)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := mail.NewStore(tc.config)
			eh := &eventstest.Handler{}
			h := mail.NewHandler(tc.config, s, enginetest.NewEngine(), eh)
			tc.test(t, h, s, eh)
		})
	}
}

func sendMail(t *testing.T, h smtp.Handler, ctx context.Context) *smtp.MailResponse {
	rr := smtptest.NewRecorder()
	h.ServeSMTP(rr, smtp.NewMailRequest("alice@foo.bar", ctx))
	return expectMailResponse(t, rr.Response)
}

func sendRcpt(t *testing.T, h smtp.Handler, ctx context.Context) *smtp.RcptResponse {
	r := sendMail(t, h, ctx)
	require.Equal(t, smtp.Ok, r.Result)
	rr := smtptest.NewRecorder()
	h.ServeSMTP(rr, smtp.NewRcptRequest("bob@foo.bar", ctx))
	return expectRcptResponse(t, rr.Response)
}

func sendThisRcpt(t *testing.T, h smtp.Handler, ctx context.Context, address string) *smtp.RcptResponse {
	r := sendMail(t, h, ctx)
	require.Equal(t, smtp.Ok, r.Result)
	rr := smtptest.NewRecorder()
	h.ServeSMTP(rr, smtp.NewRcptRequest(address, ctx))
	return expectRcptResponse(t, rr.Response)
}

func sendData(t *testing.T, h smtp.Handler, ctx context.Context) *smtp.DataResponse {
	r := sendMail(t, h, ctx)
	require.Equal(t, smtp.Ok, r.Result)
	rcpt := sendRcpt(t, h, ctx)
	require.Equal(t, smtp.Ok, rcpt.Result)
	rr := smtptest.NewRecorder()
	h.ServeSMTP(rr, smtp.NewDataRequest(&smtp.Message{
		From:    []smtp.Address{{Address: "alice@foo.bar"}},
		To:      []smtp.Address{{Address: "bob@foo.bar"}},
		Date:    time.Now(),
		Subject: "A mail message",
	}, ctx))
	return expectDataesponse(t, rr.Response)
}

func sendLogin(t *testing.T, h smtp.Handler, ctx context.Context, username, password string) *smtp.LoginResponse {
	rr := smtptest.NewRecorder()
	username = base64.StdEncoding.EncodeToString([]byte(username))
	password = base64.StdEncoding.EncodeToString([]byte(password))
	h.ServeSMTP(rr, smtp.NewLoginRequest(username, password, ctx))
	return expectLoginResponse(t, rr.Response)
}

func expectLoginResponse(t *testing.T, r smtp.Response) *smtp.LoginResponse {
	require.IsType(t, &smtp.LoginResponse{}, r)
	return r.(*smtp.LoginResponse)
}

func expectMailResponse(t *testing.T, r smtp.Response) *smtp.MailResponse {
	require.IsType(t, &smtp.MailResponse{}, r)
	return r.(*smtp.MailResponse)
}

func expectRcptResponse(t *testing.T, r smtp.Response) *smtp.RcptResponse {
	require.IsType(t, &smtp.RcptResponse{}, r)
	return r.(*smtp.RcptResponse)
}

func expectDataesponse(t *testing.T, r smtp.Response) *smtp.DataResponse {
	require.IsType(t, &smtp.DataResponse{}, r)
	return r.(*smtp.DataResponse)
}
