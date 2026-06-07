package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/generator"
	"mokapi/sortedmap"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_Run_Ldap(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "get LDAP API",
			app: runtimetest.NewApp(
				runtimetest.WithLdap(&directory.Config{
					Info: directory.Info{Name: "foo"},
				}),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApis()`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, []mcp.ApiSummary{
					{Name: "foo", Type: "ldap"},
				}, r.Result)
			},
		},
		{
			name: "get specific LDAP API",
			app: runtimetest.NewApp(
				runtimetest.WithLdap(&directory.Config{
					Info:    directory.Info{Name: "foo", Description: "bar"},
					Address: "localhost:389",
				}),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo')`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, &mcp.LdapAPI{}, r.Result)
				api := r.Result.(*mcp.LdapAPI)
				require.Equal(t, "foo", api.Name)
				require.Equal(t, "bar", api.Description)
				require.Equal(t, "ldap", api.Type)
				require.Equal(t, "localhost:389", api.Address)
			},
		},
		{
			name: "get entries",
			app: func() *runtime.App {
				entries := &sortedmap.LinkedHashMap[string, directory.Entry]{}
				entries.Set("foo", directory.Entry{
					Dn: "dn",
					Attributes: map[string][]string{
						"foo": {"bar"},
					},
				})

				app := runtimetest.NewApp(
					runtimetest.WithLdap(&directory.Config{
						Info:    directory.Info{Name: "foo", Description: "bar"},
						Address: "localhost:389",
						Entries: entries,
					}),
				)
				return app
			}(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo').getEntries()`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, []mcp.LdapEntry{}, r.Result)
				result := r.Result.([]mcp.LdapEntry)
				require.Len(t, result, 1)
				require.Contains(t, result, mcp.LdapEntry{
					Dn:         "dn",
					Attributes: map[string][]string{"foo": {"bar"}},
				})
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(123456)

			s := mcp.NewService(tc.app)
			tc.test(t, s)
		})
	}
}
