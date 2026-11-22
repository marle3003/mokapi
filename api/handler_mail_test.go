package api

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/smtp"
	"mokapi/try"
	"net/http"
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
	// the file of newMailApp: the service lists it in its configs
	const file = `"configs":[{"id":"64613435-3062-6462-3033-316532633233","url":"file://foo.yml","provider":"test","time":"2023-12-27T13:01:30Z"}]`
	newMessage := func(id string, date time.Time, subject string) *mail.Mail {
		return &mail.Mail{Message: &smtp.Message{
			From:        []smtp.Address{{Address: "bob@foo.bar"}},
			To:          []smtp.Address{{Address: "alice@foo.bar"}},
			MessageId:   id,
			Date:        date,
			Subject:     subject,
			ContentType: "text/plain",
			Body:        "foobar",
			Size:        10,
		}}
	}
	alice := func(folders ...string) *mail.Config {
		cfg := &mail.Config{Info: mail.Info{Name: "foo"}, Mailboxes: map[string]*mail.MailboxConfig{"alice@foo.bar": {Folders: map[string]*mail.FolderConfig{}}}}
		for _, f := range folders {
			cfg.Mailboxes["alice@foo.bar"].Folders[f] = &mail.FolderConfig{}
		}
		return cfg
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
				return runtimetest.NewApp(runtimetest.WithMail(&mail.Config{Info: mail.Info{Name: "foo", Description: "bar", Version: "2.1"}}))
			},
			requestUrl:   "http://foo.api/api/services",
			contentType:  "application/json",
			responseBody: `[{"name":"foo","description":"bar","version":"2.1","type":"mail","metrics":{"mails_total":0,"mail_timestamp":0}}]`,
		},
		{
			name: "get /api/services/mail",
			app: func() *runtime.App {
				return newMailApp(&mail.Config{Info: mail.Info{Name: "foo"}})
			},
			requestUrl:   "http://foo.api/api/services/mail/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo",` + file + `,"settings":{"maxRecipients":0,"autoCreateMailbox":true}}`,
		},
		{
			name: "get smtp service",
			app: func() *runtime.App {
				return newMailApp(&mail.Config{
					Info:      mail.Info{Name: "foo"},
					Mailboxes: map[string]*mail.MailboxConfig{"alice@foo.bar": {Username: "alice", Password: "foo"}},
				})
			},
			requestUrl:   "http://foo.api/api/services/mail/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo","mailboxes":[{"name":"alice@foo.bar","username":"alice","password":"foo","numMessages":0}],` + file + `,"settings":{"maxRecipients":0,"autoCreateMailbox":true}}`,
		},
		{
			name: "get smtp service with rules",
			app: func() *runtime.App {
				return newMailApp(&mail.Config{
					Info: mail.Info{Name: "foo"},
					Rules: map[string]*mail.Rule{"foo": {
						Name:      "foo",
						Sender:    mail.NewRuleExpr("alice@foo.bar"),
						Recipient: mail.NewRuleExpr("alice@foo.bar"),
						Subject:   mail.NewRuleExpr("foo"),
						Body:      mail.NewRuleExpr("bar"),
						Action:    "deny",
					}},
				})
			},
			requestUrl:   "http://foo.api/api/services/mail/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo","rules":[{"name":"foo","sender":"alice@foo.bar","recipient":"alice@foo.bar","subject":"foo","body":"bar","action":"deny"}],` + file + `,"settings":{"maxRecipients":0,"autoCreateMailbox":true}}`,
		},
		{
			name: "get smtp mailboxes",
			app: func() *runtime.App {
				app := runtimetest.NewApp(runtimetest.WithMail(&mail.Config{
					Info: mail.Info{Name: "foo"},
					Mailboxes: map[string]*mail.MailboxConfig{
						"alice@foo.bar": {Username: "alice", Password: "foo"},
						"bob@foo.bar":   {Folders: map[string]*mail.FolderConfig{"INBOX": {}}},
					},
				}))
				putMessages(app, "bob@foo.bar", "INBOX", &mail.Mail{})
				return app
			},
			requestUrl:   "http://foo.api/api/services/mail/foo/mailboxes",
			contentType:  "application/json",
			responseBody: `[{"name":"alice@foo.bar","username":"alice","password":"foo","numMessages":0},{"name":"bob@foo.bar","numMessages":1}]`,
		},
		{
			name: "get smtp mailbox",
			app: func() *runtime.App {
				return runtimetest.NewApp(runtimetest.WithMail(alice("INBOX")))
			},
			requestUrl:   "http://foo.api/api/services/mail/foo/mailboxes/alice@foo.bar",
			contentType:  "application/json",
			responseBody: `{"name":"alice@foo.bar","numMessages":0,"folders":["INBOX"]}`,
		},
		{
			name: "get smtp mailbox folder",
			app: func() *runtime.App {
				app := runtimetest.NewApp(runtimetest.WithMail(alice("INBOX")))
				putMessages(app, "alice@foo.bar", "INBOX", newMessage("foo-1@mokapi.io", now, "Hello Alice"))
				return app
			},
			requestUrl:   "http://foo.api/api/services/mail/foo/mailboxes/alice@foo.bar",
			contentType:  "application/json",
			responseBody: `{"name":"alice@foo.bar","numMessages":1,"folders":["INBOX"]}`,
		},
		{
			name: "get messages in mailbox",
			app: func() *runtime.App {
				app := runtimetest.NewApp(runtimetest.WithMail(alice("INBOX")))
				putMessages(app, "alice@foo.bar", "INBOX",
					newMessage("foo-1@mokapi.io", mustTime("2023-12-27T13:01:30+00:00"), "Hello Alice 1"),
					newMessage("foo-2@mokapi.io", mustTime("2023-12-27T13:05:30+00:00"), "Hello Alice 2"),
				)
				return app
			},
			requestUrl:   "http://foo.api/api/services/mail/foo/mailboxes/alice@foo.bar/messages",
			contentType:  "application/json",
			responseBody: `[{"messageId":"foo-2@mokapi.io","from":[{"address":"bob@foo.bar"}],"to":[{"address":"alice@foo.bar"}],"subject":"Hello Alice 2","date":"2023-12-27T13:05:30Z"},{"messageId":"foo-1@mokapi.io","from":[{"address":"bob@foo.bar"}],"to":[{"address":"alice@foo.bar"}],"subject":"Hello Alice 1","date":"2023-12-27T13:01:30Z"}]`,
		},
		{
			name: "get smtp mail",
			app: func() *runtime.App {
				app := runtimetest.NewApp(runtimetest.WithMail(alice("INBOX")))
				putMessages(app, "alice@foo.bar", "INBOX", newMessage("foo-1@mokapi.io", now, "Hello Alice"))
				return app
			},
			requestUrl:   "http://foo.api/api/services/mail/messages/foo-1@mokapi.io",
			contentType:  "application/json",
			responseBody: fmt.Sprintf(`{"service":"foo","data":{"from":[{"address":"bob@foo.bar"}],"to":[{"address":"alice@foo.bar"}],"messageId":"foo-1@mokapi.io","date":"%v","subject":"Hello Alice","contentType":"text/plain","body":"foobar","size":10}}`, now.Format(time.RFC3339Nano)),
		},
		{
			name: "get smtp mail attachment content",
			app: func() *runtime.App {
				app := runtimetest.NewApp(runtimetest.WithMail(alice("INBOX")))
				putMessages(app, "alice@foo.bar", "INBOX", &mail.Mail{Message: &smtp.Message{
					From:      []smtp.Address{{Address: "bob@foo.bar"}},
					To:        []smtp.Address{{Address: "alice@foo.bar"}},
					MessageId: "foo-1@mokapi.io",
					Date:      now,
					Attachments: []smtp.Attachment{
						{Name: "foo", ContentType: "text/plain", Data: []byte("foobar")},
					},
				}})
				return app
			},
			requestUrl:   "http://foo.api/api/services/mail/mails/foo-1@mokapi.io/attachments/foo",
			contentType:  "text/plain",
			responseBody: "foobar",
		},
		{
			name: "get SMTP mail with encodings",
			app: func() *runtime.App {
				app := runtimetest.NewApp(runtimetest.WithMail(alice("INBOX")))
				putMessages(app, "alice@foo.bar", "INBOX", &mail.Mail{Message: &smtp.Message{
					From: []smtp.Address{{Address: "bob@foo.bar"}},
					To: []smtp.Address{
						{Address: "juergen@example.com", Name: "=?UTF-8?Q?J=C3=BCrgen?="},
					},
					MessageId:   "foo-1@mokapi.io",
					Date:        now,
					Subject:     "=?UTF-8?Q?=C2=A1Buenos_d=C3=ADas!?= and =?UTF-8?B?bW9rYXBp?=",
					ContentType: "text/plain",
					Body:        "foobar",
					Size:        10,
				}})
				return app
			},
			requestUrl:   "http://foo.api/api/services/mail/messages/foo-1@mokapi.io",
			contentType:  "application/json",
			responseBody: fmt.Sprintf(`{"service":"foo","data":{"from":[{"address":"bob@foo.bar"}],"to":[{"name":"Jürgen","address":"juergen@example.com"}],"messageId":"foo-1@mokapi.io","date":"%v","subject":"¡Buenos días! and mokapi","contentType":"text/plain","body":"foobar","size":10}}`, now.Format(time.RFC3339Nano)),
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

// newMailApp adds the config as a file with a fixed time, so that the configs of the service are
// the same in every run.
func newMailApp(cfg *mail.Config) *runtime.App {
	file := dynamictest.NewConfig(cfg, dynamictest.WithInfo(dynamictest.NewConfigInfo()))
	file.Info.Time = time.Date(2023, 12, 27, 13, 1, 30, 0, time.UTC)
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	app.Mail.Add(file)
	return app
}

// putMessages stores the messages in a folder that exists in the config of the service "foo".
func putMessages(app *runtime.App, mailbox, folder string, messages ...*mail.Mail) {
	mb, _ := app.Mail.Get("foo").Mailbox(mailbox)
	mb.Folders[folder].Messages = messages
}

func TestHandler_Mail_ServiceItems(t *testing.T) {
	app := runtimetest.NewApp(runtimetest.WithMail(&mail.Config{
		Info: mail.Info{Name: "foo", Version: "1.0"},
		Mailboxes: map[string]*mail.MailboxConfig{
			"carol@foo.bar": {},
			"alice@foo.bar": {},
			"bob@foo.bar":   {},
		},
	}))

	try.Handler(t,
		http.MethodGet,
		"http://foo.api/api/services",
		nil,
		"",
		New(app, static.Api{}),
		try.HasStatusCode(200),
		try.HasBody(`[{"name":"foo","version":"1.0","type":"mail","metrics":{"mails_total":0,"mail_timestamp":0},"items":[{"name":"alice@foo.bar"},{"name":"bob@foo.bar"},{"name":"carol@foo.bar"}]}]`))
}

func TestHandler_Mail_Settings(t *testing.T) {
	app := runtimetest.NewApp()
	app.Mail.Add(newSmtpFile("a.yml", &mail.Config{Info: mail.Info{Name: "foo"}, Settings: &mail.Settings{MaxRecipients: 1, AutoCreateMailbox: true}}))
	app.Mail.Add(newSmtpFile("b.yml", &mail.Config{Info: mail.Info{Name: "foo"}, Settings: &mail.Settings{MaxRecipients: 5, AutoCreateMailbox: false}}))

	try.Handler(t,
		http.MethodGet,
		"http://foo.api/api/services/mail/foo",
		nil,
		"",
		New(app, static.Api{}),
		try.HasStatusCode(200),
		try.BodyContains(`"settings":{"maxRecipients":5,"autoCreateMailbox":false}`))
}

func newSmtpFile(url string, cfg *mail.Config) *dynamic.Config {
	return dynamictest.NewConfig(cfg, dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl(url))))
}

// An info without a config snapshot is what a request sees when the last file of a mail service
// is removed while the request is running.
func TestHandler_Mail_ConfigRemoved(t *testing.T) {
	newApp := func() *runtime.App {
		return runtimetest.NewApp(
			runtimetest.WithMail(&mail.Config{Info: mail.Info{Name: "bar", Version: "1.0"}}),
			runtimetest.WithMailInfo("foo", &runtime.MailInfo{}),
		)
	}

	t.Run("service list skips it", func(t *testing.T) {
		try.Handler(t,
			http.MethodGet,
			"http://foo.api/api/services",
			nil,
			"",
			New(newApp(), static.Api{}),
			try.HasStatusCode(200),
			try.HasBody(`[{"name":"bar","version":"1.0","type":"mail","metrics":{"mails_total":0,"mail_timestamp":0}}]`))
	})

	for _, path := range []string{"", "/mailboxes", "/mailboxes/alice@foo.bar", "/mailboxes/alice@foo.bar/messages"} {
		t.Run("returns 404 for "+path, func(t *testing.T) {
			try.Handler(t,
				http.MethodGet,
				"http://foo.api/api/services/mail/foo"+path,
				nil,
				"",
				New(newApp(), static.Api{}),
				try.HasStatusCode(404))
		})
	}

	t.Run("returns 404 after the file was removed from the store", func(t *testing.T) {
		app := runtimetest.NewApp()
		file := newSmtpFile("a.yml", &mail.Config{Info: mail.Info{Name: "foo"}})
		app.Mail.Add(file)
		app.Mail.Remove(file)

		for _, path := range []string{"", "/mailboxes"} {
			try.Handler(t,
				http.MethodGet,
				"http://foo.api/api/services/mail/foo"+path,
				nil,
				"",
				New(app, static.Api{}),
				try.HasStatusCode(404))
		}
	})
}
