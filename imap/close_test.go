package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
)

func TestClose_Response(t *testing.T) {
	calledUnselect := false
	calledExpunge := false
	handler := &imaptest.Handler{
		UnselectFunc: func(session map[string]interface{}) error {
			calledUnselect = true
			return nil
		},
		ExpungeFunc: func(set *imap.IdSet, w imap.ExpungeWriter, session map[string]interface{}) error {
			require.Nil(t, set, "no IdSet value is set")
			calledExpunge = true
			return nil
		},
	}

	p := try.GetFreePort()
	s := &imap.Server{
		Addr:    fmt.Sprintf(":%v", p),
		Handler: handler,
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

	res, err := c.Send("CLOSE")
	require.NoError(t, err)
	require.Equal(t, []string{
		"A0002 OK CLOSE completed",
	}, res)

	require.True(t, calledUnselect)
	require.True(t, calledExpunge)
}
