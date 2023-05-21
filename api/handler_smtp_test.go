package api

import (
	"fmt"
	"mokapi/config/dynamic/mail"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/smtp"
	"mokapi/try"
	"net/http"
	"regexp"
	"testing"
	"time"
)

func TestHandler_Smtp(t *testing.T) {
	now := time.Now()

	mustCompile := func(s string) *mail.RuleExpr {
		r, _ := regexp.Compile(s)
		return mail.NewRuleExpr(r)
	}

	testcases := []struct {
		name         string
		app          *runtime.App
		requestUrl   string
		contentType  string
		responseBody string
	}{
		{
			name: "get smtp services",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{Info: mail.Info{Name: "foo", Description: "bar", Version: "1.0"}},
						&mail.Store{},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services",
			contentType:  "application/json",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"smtp"}]`,
		},
		{
			name: "/api/services/smtp",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{Info: mail.Info{Name: "foo"}},
						&mail.Store{},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo","server":""}`,
		},
		{
			name: "get smtp service with mailbox",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{
							Info:      mail.Info{Name: "foo"},
							Mailboxes: []mail.MailboxConfig{{Name: "alice@foo.bar", Username: "alice", Password: "foo"}},
						},
						&mail.Store{
							Mailboxes: map[string]*mail.Mailbox{
								"alice@foo.bar": {
									Name:     "alice@foo.bar",
									Username: "alice",
									Password: "foo",
								},
							},
						},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo","server":"","mailboxes":[{"name":"alice@foo.bar","username":"alice","password":"foo"}]}`,
		},
		{
			name: "get smtp service with rules",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{
							Info: mail.Info{Name: "foo"},
							Rules: []mail.Rule{{
								Sender:    mustCompile("alice@foo.bar"),
								Recipient: mustCompile("alice@foo.bar"),
								Subject:   mustCompile("foo"),
								Body:      mustCompile("bar"),
								Action:    "deny",
							}},
						},
						&mail.Store{},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo","server":"","rules":[{"name":"","sender":"alice@foo.bar","recipient":"alice@foo.bar","subject":"foo","body":"bar","action":"deny"}]}`,
		},
		{
			name: "get smtp mailbox",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{
							Info: mail.Info{Name: "foo"},
						},
						&mail.Store{
							Mailboxes: map[string]*mail.Mailbox{
								"alice@foo.bar": {
									Name: "alice@foo.bar",
									Messages: []*smtp.Message{
										{
											Sender:      nil,
											From:        []smtp.Address{{Address: "bob@foo.bar"}},
											To:          []smtp.Address{{Address: "alice@foo.bar"}},
											MessageId:   "foo-1@mokapi.io",
											Date:        now,
											Subject:     "Hello Alice",
											ContentType: "text/plain",
											Body:        "foobar",
										},
									},
								},
							},
						},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo/mailboxes/alice@foo.bar",
			contentType:  "application/json",
			responseBody: fmt.Sprintf(`{"name":"alice@foo.bar","mails":[{"from":[{"address":"bob@foo.bar"}],"to":[{"address":"alice@foo.bar"}],"messageId":"foo-1@mokapi.io","time":"%v","subject":"Hello Alice","contentType":"text/plain","body":"foobar"}]}`, now.Format(time.RFC3339Nano)),
		},
		{
			name: "get smtp mail",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{
							Info: mail.Info{Name: "foo"},
						},
						&mail.Store{
							Mailboxes: map[string]*mail.Mailbox{
								"alice@foo.bar": {
									Name: "alice@foo.bar",
									Messages: []*smtp.Message{
										{
											Sender:      nil,
											From:        []smtp.Address{{Address: "bob@foo.bar"}},
											To:          []smtp.Address{{Address: "alice@foo.bar"}},
											MessageId:   "foo-1@mokapi.io",
											Date:        now,
											Subject:     "Hello Alice",
											ContentType: "text/plain",
											Body:        "foobar",
										},
									},
								},
							},
						},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/smtp/mails/foo-1@mokapi.io",
			contentType:  "application/json",
			responseBody: fmt.Sprintf(`{"from":[{"address":"bob@foo.bar"}],"to":[{"address":"alice@foo.bar"}],"messageId":"foo-1@mokapi.io","time":"%v","subject":"Hello Alice","contentType":"text/plain","body":"foobar"}`, now.Format(time.RFC3339Nano)),
		},
		{
			name: "get smtp mail attachment content",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{
							Info: mail.Info{Name: "foo"},
						},
						&mail.Store{
							Mailboxes: map[string]*mail.Mailbox{
								"alice@foo.bar": {
									Name: "alice@foo.bar",
									Messages: []*smtp.Message{
										{
											Sender:    nil,
											From:      []smtp.Address{{Address: "bob@foo.bar"}},
											To:        []smtp.Address{{Address: "alice@foo.bar"}},
											MessageId: "foo-1@mokapi.io",
											Date:      now,
											Attachments: []smtp.Attachment{
												{
													Name:        "foo",
													ContentType: "text/plain",
													Data:        []byte("foobar"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/smtp/mails/foo-1@mokapi.io/attachments/foo",
			contentType:  "text/plain",
			responseBody: "foobar",
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app, static.Api{})

			try.Handler(t,
				http.MethodGet,
				tc.requestUrl,
				nil,
				"",
				h,
				try.HasStatusCode(200),
				try.HasHeader("Content-Type", tc.contentType),
				try.HasBody(tc.responseBody))
		})
	}
}
