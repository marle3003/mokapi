package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestUid_Response(t *testing.T) {
	testcases := []struct {
		request  string
		response []string
		handler  imap.Handler
	}{
		{
			request: "UID FETCH 1403 (FLAGS)",
			response: []string{
				"* 1403 FETCH (FLAGS (\\Seen))",
				"A0002 OK FETCH completed",
			},
			handler: &imaptest.Handler{
				UidFetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
					msg := response.NewMessage(1403)
					msg.WriteFlags(imap.FlagSeen)
					return nil
				},
			},
		},
		{
			request: "UID FETCH 1103 (BODY.PEEK[])",
			response: []string{
				"* 1103 FETCH (BODY[] {36}",
				"From: bob@foo.bar",
				"",
				"Hello World",
				"",
				")",
				"A0002 OK FETCH completed",
			},
			handler: &imaptest.Handler{
				UidFetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
					msg := response.NewMessage(1103)
					w := msg.WriteBody2(request.Options.Body[0])
					w.WriteHeader("From", "bob@foo.bar")
					w.WriteBody("Hello World")
					w.Close()
					return nil
				},
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
				Handler: tc.handler,
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
