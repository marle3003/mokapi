package ldap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestModify(t *testing.T) {
	testcases := []struct {
		name    string
		handler func(t *testing.T, rw ResponseWriter, req *Request)
		test    func(t *testing.T, c Client)
	}{
		{
			name: "success",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				m := req.Message.(*ModifyRequest)
				require.Equal(t, "cn=foo", m.Dn)
				require.Equal(t, DeleteOperation, m.Items[0].Operation)
				require.Equal(t, "cn", m.Items[0].Modification.Type)
				require.Equal(t, []string{"Alice"}, m.Items[0].Modification.Values)

				err := rw.Write(&ModifyResponse{
					ResultCode: Success,
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.Modify(&ModifyRequest{
					Dn: "cn=foo",
					Items: []ModificationItem{
						{
							Operation: DeleteOperation,
							Modification: Modification{
								Type:   "cn",
								Values: []string{"Alice"},
							},
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
				err := rw.Write(&ModifyResponse{
					ResultCode: NoSuchObject,
					Message:    "foo",
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.Modify(&ModifyRequest{
					Dn: "cn=foo",
					Items: []ModificationItem{
						{
							Operation: DeleteOperation,
							Modification: Modification{
								Type:   "cn",
								Values: []string{"Alice"},
							},
						},
					},
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
