package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestSubscribe_Response(t *testing.T) {
	testcases := []struct {
		request  string
		response []string
		handler  newHandler
	}{
		{
			request: "SUBSCRIBE TRASH)",
			response: []string{
				"A0002 OK SUBSCRIBE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SubscribeFunc: func(mailbox string, session map[string]interface{}) error {
						require.Equal(t, "TRASH", mailbox)
						return nil
					},
				}
			},
		},
		{
			request: "UNSUBSCRIBE TRASH)",
			response: []string{
				"A0002 OK UNSUBSCRIBE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					UnsubscribeFunc: func(mailbox string, session map[string]interface{}) error {
						require.Equal(t, "TRASH", mailbox)
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
