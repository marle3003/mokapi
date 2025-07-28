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

func TestSearch_Response(t *testing.T) {
	testcases := []struct {
		name     string
		request  string
		response []string
		handler  newHandler
	}{
		{
			name:    "search all but nothing found",
			request: "SEARCH",
			response: []string{
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						return &imap.SearchResponse{All: &imap.IdSet{}}, nil
					},
				}
			},
		},
		{
			name:    "search all with result",
			request: "SEARCH",
			response: []string{
				"* SEARCH 1",
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						set := &imap.IdSet{}
						set.AddId(1)

						return &imap.SearchResponse{All: set}, nil
					},
				}
			},
		},
		{
			name:    "search specific all with result",
			request: "SEARCH ALL",
			response: []string{
				"* SEARCH 1",
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						set := &imap.IdSet{}
						set.AddId(1)

						return &imap.SearchResponse{All: set}, nil
					},
				}
			},
		},
		{
			name:    "search with sequence number",
			request: "SEARCH 1:*",
			response: []string{
				"* SEARCH 1 2",
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, "1:*", request.Criteria.Seq.String())

						set := &imap.IdSet{}
						set.AddId(1)
						set.AddId(2)

						return &imap.SearchResponse{All: set}, nil
					},
				}
			},
		},
		{
			name:    "search flags",
			request: "SEARCH ANSWERED DELETED DRAFT FLAGGED RECENT SEEN",
			response: []string{
				"* SEARCH 1",
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, []imap.Flag{"\\Answered", "\\Deleted", "\\Draft", "\\Flagged", "\\Recent", "\\Seen"}, request.Criteria.Flag)

						set := &imap.IdSet{}
						set.AddId(1)

						return &imap.SearchResponse{All: set}, nil
					},
				}
			},
		},
		{
			name:    "search not flags",
			request: "SEARCH UNANSWERED UNDELETED UNDRAFT UNFLAGGED UNSEEN",
			response: []string{
				"* SEARCH 1",
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Empty(t, request.Criteria.Flag)
						require.Equal(t, []imap.Flag{"\\Answered", "\\Deleted", "\\Draft", "\\Flagged", "\\Seen"}, request.Criteria.NotFlag)

						set := &imap.IdSet{}
						set.AddId(1)

						return &imap.SearchResponse{All: set}, nil
					},
				}
			},
		},
		{
			name:    "search NEW",
			request: "SEARCH NEW",
			response: []string{
				"* SEARCH 1",
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, []imap.Flag{"\\Recent"}, request.Criteria.Flag)
						require.Equal(t, []imap.Flag{"\\Seen"}, request.Criteria.NotFlag)

						set := &imap.IdSet{}
						set.AddId(1)

						return &imap.SearchResponse{All: set}, nil
					},
				}
			},
		},
		{
			name:    "search OLD",
			request: "SEARCH OLD",
			response: []string{
				"* SEARCH 1",
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Empty(t, request.Criteria.Flag)
						require.Equal(t, []imap.Flag{"\\Recent"}, request.Criteria.NotFlag)

						set := &imap.IdSet{}
						set.AddId(1)

						return &imap.SearchResponse{All: set}, nil
					},
				}
			},
		},
		{
			name:    "search addresses",
			request: "SEARCH BCC alice CC bob FROM carol TO john@example.com",
			response: []string{
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, []imap.HeaderCriteria{
							{Name: "Bcc", Value: "alice"},
							{Name: "Cc", Value: "bob"},
							{Name: "From", Value: "carol"},
							{Name: "To", Value: "john@example.com"},
						}, request.Criteria.Headers)

						return &imap.SearchResponse{}, nil
					},
				}
			},
		},
		{
			name:    "search before date",
			request: "SEARCH BEFORE 26-Jul-2025",
			response: []string{
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, time.Date(2025, time.July, 26, 0, 0, 0, 0, time.UTC), request.Criteria.Before)

						return &imap.SearchResponse{}, nil
					},
				}
			},
		},
		{
			name:    "search body",
			request: "SEARCH BODY foo BODY bar",
			response: []string{
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, []string{"foo", "bar"}, request.Criteria.Body)

						return &imap.SearchResponse{}, nil
					},
				}
			},
		},
		{
			name:    "search header",
			request: "SEARCH HEADER From alice",
			response: []string{
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, []imap.HeaderCriteria{{Name: "From", Value: "alice"}}, request.Criteria.Headers)

						return &imap.SearchResponse{}, nil
					},
				}
			},
		},
		{
			name:    "search larger and smaller",
			request: "SEARCH LARGER 1234 SMALLER 5678",
			response: []string{
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, int64(1234), request.Criteria.Larger)
						require.Equal(t, int64(5678), request.Criteria.Smaller)

						return &imap.SearchResponse{}, nil
					},
				}
			},
		},
		{
			name:    "search not",
			request: "SEARCH NOT FROM alice",
			response: []string{
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, []imap.SearchCriteria{
							{Headers: []imap.HeaderCriteria{{Name: "From", Value: "alice"}}},
						}, request.Criteria.Not)

						return &imap.SearchResponse{}, nil
					},
				}
			},
		},
		{
			name:    "search or",
			request: "SEARCH OR FROM alice FROM carol",
			response: []string{
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, [][2]imap.SearchCriteria{
							{
								{Headers: []imap.HeaderCriteria{{Name: "From", Value: "alice"}}},
								{Headers: []imap.HeaderCriteria{{Name: "From", Value: "carol"}}},
							},
						}, request.Criteria.Or)

						return &imap.SearchResponse{}, nil
					},
				}
			},
		},
		{
			name:    "search or with parentheses",
			request: "SEARCH OR (FROM alice SINCE 26-Jul-2025) FROM carol",
			response: []string{
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, [][2]imap.SearchCriteria{
							{
								{
									Headers: []imap.HeaderCriteria{{Name: "From", Value: "alice"}},
									Since:   time.Date(2025, time.July, 26, 0, 0, 0, 0, time.UTC),
								},
								{Headers: []imap.HeaderCriteria{{Name: "From", Value: "carol"}}},
							},
						}, request.Criteria.Or)

						return &imap.SearchResponse{}, nil
					},
				}
			},
		},
		{
			name:    "search subject",
			request: "SEARCH SUBJECT foo",
			response: []string{
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, []imap.HeaderCriteria{{Name: "Subject", Value: "foo"}}, request.Criteria.Headers)

						return &imap.SearchResponse{}, nil
					},
				}
			},
		},
		{
			name:    "search text",
			request: "SEARCH TEXT foo",
			response: []string{
				"A0002 OK SEARCH completed",
			},
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SearchFunc: func(request *imap.SearchRequest) (*imap.SearchResponse, error) {
						require.Equal(t, []string{"foo"}, request.Criteria.Text)

						return &imap.SearchResponse{}, nil
					},
				}
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
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
