package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestExpunge_Response(t *testing.T) {
	testcases := []struct {
		request  string
		response []string
		handler  newHandler
	}{
		{
			request: "EXPUNGE",
			response: []string{
				"* 1 EXPUNGE",
				"A0002 OK EXPUNGE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					ExpungeFunc: func(set *imap.IdSet, w imap.ExpungeWriter, session map[string]interface{}) error {
						require.Nil(t, set, "no IdSet value is set")

						return w.Write(uint32(1))
					},
				}
			},
		},
		{
			request: "UID EXPUNGE 145",
			response: []string{
				"* 145 EXPUNGE",
				"A0002 OK EXPUNGE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				count := 0
				return &imaptest.Handler{
					ExpungeFunc: func(set *imap.IdSet, w imap.ExpungeWriter, session map[string]interface{}) error {
						count++
						// close does also call expunge
						if count == 1 {
							require.NotNil(t, set, "IdSet value is set")
							require.Equal(t, imap.IdNum(145), set.Ids[0].(imap.IdNum))
						}

						return w.Write(uint32(145))
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
