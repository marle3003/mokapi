package api

import (
	"mokapi/config/dynamic/mail"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"regexp"
	"testing"
)

func TestHandler_Smtp(t *testing.T) {
	mustCompile := func(s string) *mail.RuleExpr {
		r, _ := regexp.Compile(s)
		return mail.NewRuleExpr(r)
	}

	testcases := []struct {
		name         string
		app          *runtime.App
		requestUrl   string
		responseBody string
	}{
		{
			name: "get smtp services",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{Info: mail.Info{Name: "foo", Description: "bar", Version: "1.0"}},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"smtp"}]`,
		},
		{
			name: "/api/services/smtp",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{Info: mail.Info{Name: "foo"}},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo",
			responseBody: `{"name":"foo","server":""}`,
		},
		{
			name: "get smtp service with mailbox",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{
							Info:      mail.Info{Name: "foo"},
							Mailboxes: []mail.Mailbox{{Name: "alice@foo.bar", Username: "alice", Password: "foo"}},
						},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo",
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
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo",
			responseBody: `{"name":"foo","server":"","rules":[{"sender":"alice@foo.bar","recipient":"alice@foo.bar","subject":"foo","body":"bar","action":"deny"}]}`,
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
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(tc.responseBody))
		})
	}
}
