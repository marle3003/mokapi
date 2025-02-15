package ldap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAdd(t *testing.T) {
	testcases := []struct {
		name    string
		handler func(t *testing.T, rw ResponseWriter, req *Request)
		test    func(t *testing.T, c Client)
	}{
		{
			name: "success",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				m := req.Message.(*AddRequest)
				require.Equal(t, "cn=foo", m.Dn)
				require.Equal(t, "cn", m.Attributes[0].Type)
				require.Equal(t, []string{"foo"}, m.Attributes[0].Values)

				err := rw.Write(&AddResponse{
					ResultCode: Success,
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.Add(&AddRequest{
					Dn: "cn=foo",
					Attributes: []Attribute{
						{
							Type:   "cn",
							Values: []string{"foo"},
						},
					},
				})
				require.NoError(t, err)
				require.Equal(t, Success, r.ResultCode)
			},
		},
		{
			name: "error",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				err := rw.Write(&AddResponse{
					ResultCode: EntryAlreadyExists,
					Message:    "foo",
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.Add(&AddRequest{
					Dn: "cn=foo",
				})
				require.NoError(t, err)
				require.Equal(t, EntryAlreadyExists, r.ResultCode)
				require.Equal(t, "foo", r.Message)
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
