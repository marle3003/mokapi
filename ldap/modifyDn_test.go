package ldap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestModifyDn(t *testing.T) {
	testcases := []struct {
		name    string
		handler func(t *testing.T, rw ResponseWriter, req *Request)
		test    func(t *testing.T, c Client)
	}{
		{
			name: "success",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				m := req.Message.(*ModifyDNRequest)
				require.Equal(t, "cn=foo", m.Dn)
				require.Equal(t, "cn=bar", m.NewRdn)
				require.Equal(t, true, m.DeleteOldDn)

				err := rw.Write(&ModifyDNResponse{
					ResultCode: Success,
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.ModifyDn(&ModifyDNRequest{
					Dn:          "cn=foo",
					NewRdn:      "cn=bar",
					DeleteOldDn: true,
				})
				require.NoError(t, err)
				require.Equal(t, Success, r.ResultCode)
			},
		},
		{
			name: "error",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				err := rw.Write(&ModifyResponse{
					ResultCode: NoSuchObject,
					Message:    "foo",
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.ModifyDn(&ModifyDNRequest{
					Dn: "cn=foo",
				})
				require.NoError(t, err)
				require.Equal(t, NoSuchObject, r.ResultCode)
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
