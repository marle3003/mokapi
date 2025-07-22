package mail

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStore_Update(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*Config
		test    func(s *Store, t *testing.T)
	}{
		{
			name: "add mailbox",
			configs: []*Config{
				{
					Mailboxes: []MailboxConfig{
						{
							Name: "foo@example.com",
						},
					},
				},
				{
					Mailboxes: []MailboxConfig{
						{
							Name: "bar@example.com",
						},
					},
				},
			},
			test: func(s *Store, t *testing.T) {
				require.Len(t, s.Mailboxes, 2)
			},
		},
		{
			name: "update mailbox",
			configs: []*Config{
				{
					Mailboxes: []MailboxConfig{
						{
							Name: "foo@example.com",
						},
					},
				},
				{
					Mailboxes: []MailboxConfig{
						{
							Name:     "foo@example.com",
							Password: "secret",
							Folders: []FolderConfig{
								{Name: "Trash"},
							},
						},
					},
				},
			},
			test: func(s *Store, t *testing.T) {
				require.Len(t, s.Mailboxes, 1)
				require.Equal(t, "secret", s.Mailboxes["foo@example.com"].Password)
				require.Contains(t, s.Mailboxes["foo@example.com"].Folders, "Trash")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := NewStore(tc.configs[0])
			for _, cfg := range tc.configs[1:] {
				s.Update(cfg)
			}
			tc.test(s, t)
		})
	}
}
