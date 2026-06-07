package mcp_test

import (
	"context"
	"mokapi/imap"
	"mokapi/mcp"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/generator"
	"mokapi/smtp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_Run_Mail(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "get Mail API",
			app: runtimetest.NewApp(
				runtimetest.WithMail(&mail.Config{
					Version: "1.0",
					Info: mail.Info{
						Name:        "mail",
						Description: "description",
					},
				}),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApis()`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, []mcp.ApiSummary{
					{Name: "mail", Type: "mail"},
				}, r.Result)
			},
		},
		{
			name: "get specific Mail API",
			app: runtimetest.NewApp(
				runtimetest.WithMail(&mail.Config{
					Version: "1.0",
					Info: mail.Info{
						Name:        "foo",
						Description: "description",
					},
					Servers: map[string]*mail.Server{
						"smtp": {
							Host:        "smtp.example.com",
							Description: "A smtp server",
							Protocol:    "smtp",
						},
					},
				}),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo')`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, &mcp.MailAPI{}, r.Result)
				api := r.Result.(*mcp.MailAPI)
				require.Equal(t, "foo", api.Name)
				require.Equal(t, "description", api.Description)
				require.Equal(t, "mail", api.Type)
				require.Equal(t, []mcp.MailServer{
					{
						Protocol:    "smtp",
						Host:        "smtp.example.com",
						Description: "A smtp server",
					},
				}, api.Servers)
			},
		},
		{
			name: "get mailboxes",
			app: runtimetest.NewApp(
				runtimetest.WithMail(&mail.Config{
					Version: "1.0",
					Info: mail.Info{
						Name:        "foo",
						Description: "description",
					},
					Mailboxes: map[string]*mail.MailboxConfig{
						"foo@mokapi.io": {Description: "foo description"},
						"bar@mokapi.io": {Description: "bar description"},
					},
				}),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo').getMailboxes()`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, []mcp.MailboxSummary{}, r.Result)
				result := r.Result.([]mcp.MailboxSummary)
				require.Len(t, result, 2)
				require.Contains(t, result, mcp.MailboxSummary{
					Name:        "foo@mokapi.io",
					Description: "foo description",
				})
				require.Contains(t, result, mcp.MailboxSummary{
					Name:        "bar@mokapi.io",
					Description: "bar description",
				})
			},
		},
		{
			name: "get mailbox",
			app: func() *runtime.App {
				app := runtimetest.NewApp(
					runtimetest.WithMail(&mail.Config{
						Version: "1.0",
						Info: mail.Info{
							Name:        "foo",
							Description: "description",
						},
						Mailboxes: map[string]*mail.MailboxConfig{
							"foo@mokapi.io": {
								Username:    "foo",
								Password:    "secret",
								Description: "foo description",
								Folders: map[string]*mail.FolderConfig{
									"temp": {Flags: []string{"\\HasNoChildren"}},
								},
							},
						},
					}),
				)
				m := app.Mail.Get("foo")
				f := m.Store.Mailboxes["foo@mokapi.io"].Folders["temp"]
				f.Append(&smtp.Message{
					Subject: "Hello world!",
				})
				return app
			}(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo').getMailbox('foo@mokapi.io')`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, &mcp.Mailbox{}, r.Result)
				result := r.Result.(*mcp.Mailbox)
				require.Equal(t, "foo@mokapi.io", result.Name)
				require.Equal(t, "foo", result.Username)
				require.Equal(t, "secret", result.Password)
				require.Equal(t, "foo description", result.Description)
				require.Len(t, result.Folders, 1)
				require.Equal(t, "temp", result.Folders["temp"].Name)
				require.Equal(t, []imap.MailboxFlags{imap.HasNoChildren}, result.Folders["temp"].Flags)
				require.Len(t, result.Folders["temp"].Mails, 1)
				require.Equal(t, "Hello world!", result.Folders["temp"].Mails[0].Subject)
			},
		},
		{
			name: "send mail",
			app: runtimetest.NewApp(
				runtimetest.WithMail(&mail.Config{
					Version: "1.0",
					Info: mail.Info{
						Name:        "foo",
						Description: "description",
					},
					Mailboxes: map[string]*mail.MailboxConfig{
						"foo@mokapi.io": {
							Username:    "foo",
							Password:    "secret",
							Description: "foo description",
							Folders: map[string]*mail.FolderConfig{
								"temp": {Flags: []string{"\\HasNoChildren"}},
							},
						},
					},
				}),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const api = mokapi.getApi('foo')
api.sendMail('foo@mokapi.io', { subject: 'Hello world!'})
api.getMailbox('foo@mokapi.io')`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, &mcp.Mailbox{}, r.Result)
				result := r.Result.(*mcp.Mailbox)
				require.Equal(t, "foo@mokapi.io", result.Name)
				require.Equal(t, "foo", result.Username)
				require.Equal(t, "secret", result.Password)
				require.Equal(t, "foo description", result.Description)
				require.Len(t, result.Folders, 2)
				require.Contains(t, result.Folders, "temp")
				require.Contains(t, result.Folders, "INBOX")
				require.Len(t, result.Folders["INBOX"].Mails, 1)
				require.Equal(t, "Hello world!", result.Folders["INBOX"].Mails[0].Subject)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(123456)

			s := mcp.NewService(tc.app)
			tc.test(t, s)
		})
	}
}
