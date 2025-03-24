package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestMove_Response(t *testing.T) {
	testcases := []struct {
		request  string
		response []string
		handler  newHandler
	}{
		{
			request: "MOVE 1 foo",
			response: []string{
				"* [COPYUID 1435 1 23] COPY",
				"* 1 EXPUNGE",
				"A0002 OK MOVE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					MoveFunc: func(set *imap.IdSet, dest string, w imap.MoveWriter, session map[string]interface{}) error {
						require.Equal(t, imap.IdNum(1), set.Ids[0].(imap.IdNum))
						require.Equal(t, "foo", dest)

						err := w.WriteCopy(&imap.Copy{
							UIDValidity: 1435,
							SourceUIDs:  imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
							DestUIDs:    imap.IdSet{Ids: []imap.Set{imap.IdNum(23)}},
						})
						require.NoError(t, err)
						return w.WriteExpunge(1)
					},
				}
			},
		},
		{
			request: "MOVE 1:2 foo",
			response: []string{
				"* [COPYUID 1435 1:2 23:24] COPY",
				"* 1 EXPUNGE",
				"A0002 OK MOVE completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					MoveFunc: func(set *imap.IdSet, dest string, w imap.MoveWriter, session map[string]interface{}) error {
						require.Equal(t, uint32(1), set.Ids[0].(*imap.Range).Start.Value)
						require.Equal(t, "foo", dest)

						err := w.WriteCopy(&imap.Copy{
							UIDValidity: 1435,
							SourceUIDs:  imap.IdSet{Ids: []imap.Set{&imap.Range{Start: imap.SeqNum{Value: 1}, End: imap.SeqNum{Value: 2}}}},
							DestUIDs:    imap.IdSet{Ids: []imap.Set{&imap.Range{Start: imap.SeqNum{Value: 23}, End: imap.SeqNum{Value: 24}}}},
						})
						require.NoError(t, err)
						return w.WriteExpunge(1)
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
