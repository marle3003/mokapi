package ldap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient(t *testing.T) {
	c := NewClient("ldap.forumsys.com:389")
	err := c.Dial()
	require.NoError(t, err)
	_, err = c.Bind("cn=read-only-admin,dc=example,dc=com", "password")
	require.NoError(t, err)
	r, err := c.Search(&SearchRequest{
		BaseDN:            "dc=example,dc=com",
		Scope:             ScopeSingleLevel,
		DereferencePolicy: 0,
		SizeLimit:         0,
		TimeLimit:         10,
		TypesOnly:         false,
		Filter:            "(objectClass=*)",
	})
	require.NoError(t, err)
	_ = r
}
