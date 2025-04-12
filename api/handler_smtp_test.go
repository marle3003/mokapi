package api

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
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
	mustTime := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return t
	}

	mustCompile := func(s string) *mail.RuleExpr {
		r, _ := regexp.Compile(s)
		return mail.NewRuleExpr(r)
	}

	testcases := []struct {
		name         string
		app          func() *runtime.App
		requestUrl   string
		contentType  string
		responseBody string
	}{
		{
			name: "get smtp services",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				app.Mail.Set("foo", &runtime.MailInfo{
					Config: &mail.Config{Info: mail.Info{Name: "foo", Description: "bar", Version: "1.0"}},
					Store:  &mail.Store{},
				})
				return app
			},
			requestUrl:   "http://foo.api/api/services",
			contentType:  "application/json",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"smtp"}]`,
		},
		{
			name: "/api/services/smtp",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				app.Mail.Set("foo", &runtime.MailInfo{
					Config: &mail.Config{Info: mail.Info{Name: "foo"}},
					Store:  &mail.Store{},
				})
				return app
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo","server":""}`,
		},
		{
			name: "get smtp service",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				cfg := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(),
					Data: &mail.Config{
						Info:      mail.Info{Name: "foo"},
						Mailboxes: []mail.MailboxConfig{{Name: "alice@foo.bar", Username: "alice", Password: "foo"}},
					},
				}
				cfg.Info.Time = mustTime("2023-12-27T13:01:30+00:00")

				app.Mail.Add(cfg)
				return app
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo","server":"","mailboxes":[{"name":"alice@foo.bar","username":"alice","password":"foo"}],"configs":[{"id":"64613435-3062-6462-3033-316532633233","url":"file://foo.yml","provider":"test","time":"2023-12-27T13:01:30Z"}]}`,
		},
		{
			name: "get smtp service with mailbox",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				app.Mail.Set("foo", &runtime.MailInfo{
					Config: &mail.Config{
						Info:      mail.Info{Name: "foo"},
						Mailboxes: []mail.MailboxConfig{{Name: "alice@foo.bar", Username: "alice", Password: "foo"}},
					},
					Store: &mail.Store{
						Mailboxes: map[string]*mail.Mailbox{
							"alice@foo.bar": {
								Name:     "alice@foo.bar",
								Username: "alice",
								Password: "foo",
							},
						},
					},
				})
				return app
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo","server":"","mailboxes":[{"name":"alice@foo.bar","username":"alice","password":"foo"}]}`,
		},
		{
			name: "get smtp service with rules",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				app.Mail.Set("foo", &runtime.MailInfo{
					Config: &mail.Config{
						Info: mail.Info{Name: "foo"},
						Rules: []mail.Rule{{
							Sender:    mustCompile("alice@foo.bar"),
							Recipient: mustCompile("alice@foo.bar"),
							Subject:   mustCompile("foo"),
							Body:      mustCompile("bar"),
							Action:    "deny",
						}},
					},
					Store: &mail.Store{},
				})
				return app
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo","server":"","rules":[{"name":"","sender":"alice@foo.bar","recipient":"alice@foo.bar","subject":"foo","body":"bar","action":"deny"}]}`,
		},
		{
			name: "get smtp mailbox",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				app.Mail.Set("foo", &runtime.MailInfo{
					Config: &mail.Config{
						Info: mail.Info{Name: "foo"},
					},
					Store: &mail.Store{
						Mailboxes: map[string]*mail.Mailbox{
							"alice@foo.bar": {
								Name: "alice@foo.bar",
								Folders: map[string]*mail.Folder{
									"Inbox": {
										Messages: []*mail.Mail{
											{
												Message: &smtp.Message{Sender: nil,
													From:        []smtp.Address{{Address: "bob@foo.bar"}},
													To:          []smtp.Address{{Address: "alice@foo.bar"}},
													MessageId:   "foo-1@mokapi.io",
													Time:        now,
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
				})
				return app
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo/mailboxes/alice@foo.bar",
			contentType:  "application/json",
			responseBody: fmt.Sprintf(`{"name":"alice@foo.bar","folders":{"Inbox":{"mails":[{"from":[{"address":"bob@foo.bar"}],"to":[{"address":"alice@foo.bar"}],"messageId":"foo-1@mokapi.io","time":"%v","subject":"Hello Alice","contentType":"text/plain","body":"foobar"}]}}}`, now.Format(time.RFC3339Nano)),
		},
		{
			name: "get smtp mail",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				app.Mail.Set("foo", &runtime.MailInfo{
					Config: &mail.Config{
						Info: mail.Info{Name: "foo"},
					},
					Store: &mail.Store{
						Mailboxes: map[string]*mail.Mailbox{
							"alice@foo.bar": {
								Name: "alice@foo.bar",
								Folders: map[string]*mail.Folder{
									"Inbox": {
										Messages: []*mail.Mail{
											{
												Message: &smtp.Message{
													Sender:      nil,
													From:        []smtp.Address{{Address: "bob@foo.bar"}},
													To:          []smtp.Address{{Address: "alice@foo.bar"}},
													MessageId:   "foo-1@mokapi.io",
													Time:        now,
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
				})
				return app
			},
			requestUrl:   "http://foo.api/api/services/smtp/mails/foo-1@mokapi.io",
			contentType:  "application/json",
			responseBody: fmt.Sprintf(`{"from":[{"address":"bob@foo.bar"}],"to":[{"address":"alice@foo.bar"}],"messageId":"foo-1@mokapi.io","time":"%v","subject":"Hello Alice","contentType":"text/plain","body":"foobar"}`, now.Format(time.RFC3339Nano)),
		},
		{
			name: "get smtp mail attachment content",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{})
				app.Mail.Set("foo", &runtime.MailInfo{
					Config: &mail.Config{
						Info: mail.Info{Name: "foo"},
					},
					Store: &mail.Store{
						Mailboxes: map[string]*mail.Mailbox{
							"alice@foo.bar": {
								Name: "alice@foo.bar",
								Folders: map[string]*mail.Folder{
									"Inbox": {
										Messages: []*mail.Mail{
											{
												Message: &smtp.Message{
													Sender:    nil,
													From:      []smtp.Address{{Address: "bob@foo.bar"}},
													To:        []smtp.Address{{Address: "alice@foo.bar"}},
													MessageId: "foo-1@mokapi.io",
													Time:      now,
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
				})
				return app
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

			h := New(tc.app(), static.Api{})

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
