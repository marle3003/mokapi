package mcp_test

import (
	"context"
	"io"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/mcp"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/generator"
	"testing"

	"github.com/sirupsen/logrus"
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
				entries := directory.NewEntryStore()
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

func TestService_Run_Ldap_ConfigRemoved(t *testing.T) {
	logrus.SetOutput(io.Discard)

	// an info whose last file was removed has no config snapshot (nil)
	t.Run("info without config is not found by getApi", func(t *testing.T) {
		app := runtimetest.NewApp(runtimetest.WithLdapInfo("foo", &runtime.LdapInfo{}))
		s := mcp.NewService(app)

		r, err := s.GetRunResponse(context.Background(), mcp.RunInput{Code: `mokapi.getApi('foo')`})

		require.NoError(t, err)
		require.Nil(t, r.Result)
	})

	t.Run("info without config is skipped by getApis", func(t *testing.T) {
		app := runtimetest.NewApp(
			runtimetest.WithLdapInfo("foo", &runtime.LdapInfo{}),
			runtimetest.WithLdap(&directory.Config{Info: directory.Info{Name: "bar"}}),
		)
		s := mcp.NewService(app)

		r, err := s.GetRunResponse(context.Background(), mcp.RunInput{Code: `mokapi.getApis()`})

		require.NoError(t, err)
		require.Equal(t, []mcp.ApiSummary{{Name: "bar", Type: "ldap"}}, r.Result)
	})

	t.Run("API removed while the object is in use returns not found", func(t *testing.T) {
		app := runtimetest.NewApp()
		file := dynamictest.NewConfig(&directory.Config{Info: directory.Info{Name: "foo"}}, dynamictest.WithInfo(dynamictest.NewConfigInfo()))
		app.Ldap.Add(file, enginetest.NewEngine())
		r, err := mcp.NewService(app).GetRunResponse(context.Background(), mcp.RunInput{Code: `mokapi.getApi('foo')`})
		require.NoError(t, err)
		l := r.Result.(*mcp.LdapAPI)

		app.Ldap.Remove(file)

		require.NotPanics(t, func() {
			_, err = l.GetEntries()
			require.EqualError(t, err, "ldap API 'foo' not found")
		})
	})

	t.Run("entries are copies", func(t *testing.T) {
		entries := directory.NewEntryStore()
		entries.Set("cn=foo", directory.Entry{Dn: "cn=foo", Attributes: map[string][]string{"mail": {"a@foo.bar"}}})
		app := runtimetest.NewApp(runtimetest.WithLdap(&directory.Config{Info: directory.Info{Name: "foo"}, Entries: entries}))
		r, err := mcp.NewService(app).GetRunResponse(context.Background(), mcp.RunInput{Code: `mokapi.getApi('foo')`})
		require.NoError(t, err)
		l := r.Result.(*mcp.LdapAPI)

		result, err := l.GetEntries()
		require.NoError(t, err)
		result[0].Attributes["mail"][0] = "changed"
		result[0].Attributes["other"] = []string{"x"}

		again, err := l.GetEntries()
		require.NoError(t, err)
		require.Equal(t, map[string][]string{"mail": {"a@foo.bar"}}, again[0].Attributes)
	})
}
