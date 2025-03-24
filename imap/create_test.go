package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestCreate_Response(t *testing.T) {
	testcases := []struct {
		request  string
		response []string
		handler  newHandler
	}{
		{
			request: "CREATE foo",
			response: []string{
				"A0002 OK CREATE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					CreateFunc: func(name string, opt *imap.CreateOptions, session map[string]interface{}) error {
						require.Equal(t, "foo", name)

						return nil
					},
				}
			},
		},
		{
			request: "CREATE \"foo\"",
			response: []string{
				"A0002 OK CREATE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					CreateFunc: func(name string, opt *imap.CreateOptions, session map[string]interface{}) error {
						require.Equal(t, "foo", name)

						return nil
					},
				}
			},
		},
		{
			request: "CREATE \"My Folder\"",
			response: []string{
				"A0002 OK CREATE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					CreateFunc: func(name string, opt *imap.CreateOptions, session map[string]interface{}) error {
						require.Equal(t, "My Folder", name)

						return nil
					},
				}
			},
		},
		{
			request: "CREATE Inbox/Subfolder",
			response: []string{
				"A0002 OK CREATE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					CreateFunc: func(name string, opt *imap.CreateOptions, session map[string]interface{}) error {
						require.Equal(t, "Inbox/Subfolder", name)

						return nil
					},
				}
			},
		},
		{
			request: "CREATE \"📧 Unicode 🚀\"",
			response: []string{
				"A0002 OK CREATE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					CreateFunc: func(name string, opt *imap.CreateOptions, session map[string]interface{}) error {
						require.Equal(t, "📧 Unicode 🚀", name)

						return nil
					},
				}
			},
		},
		{
			request: "CREATE Trash (USE (\\Trash))",
			response: []string{
				"A0002 OK CREATE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					CreateFunc: func(name string, opt *imap.CreateOptions, session map[string]interface{}) error {
						require.Equal(t, "Trash", name)
						require.Equal(t, []imap.MailboxFlags{imap.Trash}, opt.Flags)

						return nil
					},
				}
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.request, func(t *testing.T) {
			p := try.GetFreePort()
			s := &imap.Server{
				Addr:    fmt.Sprintf(":%v", p),
				Handler: tc.handler(t),
			}
			defer s.Close()
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imap.NewClient(fmt.Sprintf("localhost:%v", p))
			defer c.Close()

			_, err := c.Dial()
			require.NoError(t, err)

			err = c.PlainAuth("", "", "")
			require.NoError(t, err)

			res, err := c.Send(tc.request)
			require.NoError(t, err)
			require.Equal(t, tc.response, res)
		})
	}
}
