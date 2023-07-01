package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
	"time"
)

func TestServer_Fetch(t *testing.T) {
	testcases := []struct {
		name    string
		handler imap.Handler
		test    func(t *testing.T, c *imaptest.Client)
	}{
		{
			name: "fetch one fast",
			handler: &imaptest.Handler{
				FetchFunc: func(request *imap.FetchRequest, session map[string]interface{}) ([]imap.FetchResult, error) {
					date, err := time.Parse(time.RFC3339, "2023-03-16T13:07:04+01:00")
					require.NoError(t, err)
					return []imap.FetchResult{
						{
							SequenceNumber: 1,
							Flags:          nil,
							InternalDate:   date,
							Size:           43078,
						},
					}, nil
				},
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				lines, err := c.Send("FETCH 1 FAST")
				require.NoError(t, err)
				require.Equal(t, "* 1 FETCH (FLAGS () INTERNALDATE \"16-Mar-2023 13:07:04 +0100\" RFC822.SIZE 43078)", lines[0])
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p, err := try.GetFreePort()
			require.NoError(t, err)
			s := &imap.Server{
				Addr:    fmt.Sprintf(":%v", p),
				Handler: tc.handler,
			}
			defer s.Close()
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imaptest.NewClient(fmt.Sprintf("localhost:%v", p))

			tc.test(t, c)
		})
	}
}
