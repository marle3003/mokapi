package ldap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSearch(t *testing.T) {
	testcases := []struct {
		name    string
		handler func(t *testing.T, rw ResponseWriter, req *Request)
		test    func(t *testing.T, c Client)
	}{
		{
			name: "search no results",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				err := rw.Write(&SearchResponse{
					Status: Success,
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.Search(&SearchRequest{Filter: "(objectClass=foo)"})
				require.NoError(t, err)
				require.Equal(t, Success, r.Status)
				require.Len(t, r.Results, 0)
				require.Equal(t, "", r.Message)
			},
		},
		{
			name: "search with one results",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				err := rw.Write(&SearchResponse{
					Status: Success,
					Results: []SearchResult{
						{
							Dn:         "foo",
							Attributes: map[string][]string{"foo": {"bar"}},
						},
					},
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.Search(&SearchRequest{Filter: "(objectClass=foo)"})
				require.NoError(t, err)
				require.Equal(t, Success, r.Status)
				require.Len(t, r.Results, 1)
				require.Equal(t, "foo", r.Results[0].Dn)
				require.Len(t, r.Results[0].Attributes, 1)
				require.Equal(t, []string{"bar"}, r.Results[0].Attributes["foo"])
				require.Equal(t, "", r.Message)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s, c := newTestServer(HandlerFunc(func(rw ResponseWriter, req *Request) {
				tc.handler(t, rw, req)
			}))
			defer s.Close()
			require.NoError(t, c.Dial())
			tc.test(t, c)
		})
	}
}

func TestSearch_Filter(t *testing.T) {
	testcases := []struct {
		name    string
		handler func(t *testing.T, rw ResponseWriter, req *Request)
		test    func(t *testing.T, c Client)
	}{
		{
			name: "equals",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(objectClass=foo)", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(objectClass=foo)"})
				require.NoError(t, err)
			},
		},
		{
			name: "present",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(objectClass=*)", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(objectClass=*)"})
				require.NoError(t, err)
			},
		},
		{
			name: "Absence (Negation)",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(!(attribute=*))", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(!(attribute=*))"})
				require.NoError(t, err)
			},
		},
		{
			name: "less or equal",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(attribute<=abc)", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(attribute<=abc)"})
				require.NoError(t, err)
			},
		},
		{
			name: "greater or equal",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(attribute>=abc)", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(attribute>=abc)"})
				require.NoError(t, err)
			},
		},
		{
			name: "and",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(&(attribute=abc)(foo=bar))", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(&(attribute=abc)(foo=bar))"})
				require.NoError(t, err)
			},
		},
		{
			name: "or",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(|(attribute=abc)(foo=bar))", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(|(attribute=abc)(foo=bar))"})
				require.NoError(t, err)
			},
		},
		{
			name: "and + or",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(|(&(attribute=abc)(foo=bar))(&(attribute=abc)(bar=foo)))", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(|(&(attribute=abc)(foo=bar))(&(attribute=abc)(bar=foo)))"})
				require.NoError(t, err)
			},
		},
		{
			name: "substring startWith",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(sn=F*)", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(sn=F*)"})
				require.NoError(t, err)
			},
		},
		{
			name: "substring endWith",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(mail=*@foo.bar)", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(mail=*@foo.bar)"})
				require.NoError(t, err)
			},
		},
		{
			name: "substring any",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(givenName=*Bob*)", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(givenName=*Bob*)"})
				require.NoError(t, err)
			},
		},
		{
			name: "substring complex",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				search := req.Message.(*SearchRequest)
				require.Equal(t, "(attribute=*foo*bar*com)", search.Filter)
				rw.Write(&SearchResponse{Status: Success})
			},
			test: func(t *testing.T, c Client) {
				_, err := c.Search(&SearchRequest{Filter: "(attribute=*foo*bar*com)"})
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s, c := newTestServer(HandlerFunc(func(rw ResponseWriter, req *Request) {
				tc.handler(t, rw, req)
			}))
			defer s.Close()
			require.NoError(t, c.Dial())
			tc.test(t, c)
		})
	}
}
