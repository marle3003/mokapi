package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestStore_Response(t *testing.T) {
	testcases := []struct {
		request  string
		response []string
		handler  imap.Handler
	}{
		{
			request: "STORE 1 +FLAGS (\\Seen)",
			response: []string{
				"* 1 FETCH (FLAGS (\\Seen))",
				"A0002 OK STORE completed",
			},
			handler: &imaptest.Handler{
				StoreFunc: func(request *imap.StoreRequest, response imap.FetchResponse, session map[string]interface{}) error {
					require.Equal(t, "add", request.Action)
					require.Equal(t, []imap.Flag{"\\Seen"}, request.Flags)

					msg := response.NewMessage(1)
					msg.WriteFlags(imap.FlagSeen)
					return nil
				},
			},
		},
		{
			request: "STORE 1 -FLAGS (\\Seen)",
			response: []string{
				"* 1 FETCH (FLAGS ())",
				"A0002 OK STORE completed",
			},
			handler: &imaptest.Handler{
				StoreFunc: func(request *imap.StoreRequest, response imap.FetchResponse, session map[string]interface{}) error {
					require.Equal(t, "remove", request.Action)
					require.Equal(t, []imap.Flag{"\\Seen"}, request.Flags)

					msg := response.NewMessage(1)
					msg.WriteFlags()
					return nil
				},
			},
		},
		{
			request: "STORE 1 FLAGS (\\Seen)",
			response: []string{
				"* 1 FETCH (FLAGS (\\Seen))",
				"A0002 OK STORE completed",
			},
			handler: &imaptest.Handler{
				StoreFunc: func(request *imap.StoreRequest, response imap.FetchResponse, session map[string]interface{}) error {
					require.Equal(t, "replace", request.Action)
					require.Equal(t, []imap.Flag{"\\Seen"}, request.Flags)

					msg := response.NewMessage(1)
					msg.WriteFlags(imap.FlagSeen)
					return nil
				},
			},
		},
		{
			request: "STORE 1 +FLAGS.SILENT (\\Seen)",
			response: []string{
				"A0002 OK STORE completed",
			},
			handler: &imaptest.Handler{
				StoreFunc: func(request *imap.StoreRequest, response imap.FetchResponse, session map[string]interface{}) error {
					require.Equal(t, "add", request.Action)
					require.Equal(t, []imap.Flag{"\\Seen"}, request.Flags)
					require.True(t, request.Silent, "silent should be true")

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
