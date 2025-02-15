package ldap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCompare(t *testing.T) {
	testcases := []struct {
		name    string
		handler func(t *testing.T, rw ResponseWriter, req *Request)
		test    func(t *testing.T, c Client)
	}{
		{
			name: "success",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				m := req.Message.(*CompareRequest)
				require.Equal(t, "cn=foo", m.Dn)
				require.Equal(t, "cn", m.Attribute)
				require.Equal(t, "foo", m.Value)

				err := rw.Write(&CompareResponse{
					ResultCode: CompareTrue,
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.Compare(&CompareRequest{
					Dn:        "cn=foo",
					Attribute: "cn",
					Value:     "foo",
				})
				require.NoError(t, err)
				require.Equal(t, CompareTrue, r.ResultCode)
			},
		},
		{
			name: "false",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				err := rw.Write(&CompareResponse{
					ResultCode: CompareFalse,
					Message:    "foo",
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.Compare(&CompareRequest{
					Dn: "cn=foo",
				})
				require.NoError(t, err)
				require.Equal(t, CompareFalse, r.ResultCode)
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
