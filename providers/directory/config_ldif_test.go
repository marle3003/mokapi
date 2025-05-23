package directory

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/try"
	"testing"
)

func TestLdif_Parse(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		reader dynamic.Reader
		test   func(t *testing.T, ld *Ldif, err error)
	}{
		{
			name:  "one record with dn",
			input: "dn: dc=mokapi, dc=io",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &AddRecord{
					Dn: "dc=mokapi,dc=io",
				}, ld.Records[0])
			},
		},
		{
			name:  "two records with dn",
			input: "dn: dc=mokapi, dc=io\n\ndn: ou=Sales, dc=mokapi, dc=io",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 2)
				require.Equal(t, &AddRecord{
					Dn: "dc=mokapi,dc=io",
				}, ld.Records[0])
				require.Equal(t, &AddRecord{
					Dn: "ou=Sales,dc=mokapi,dc=io",
				}, ld.Records[1])
			},
		},
		{
			name:  "multiple line breaks between records",
			input: "dn: dc=mokapi, dc=io\n\n\n\n\ndn: ou=Sales, dc=mokapi, dc=io",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 2)
				require.Equal(t, &AddRecord{
					Dn: "dc=mokapi,dc=io",
				}, ld.Records[0])
				require.Equal(t, &AddRecord{
					Dn: "ou=Sales,dc=mokapi,dc=io",
				}, ld.Records[1])
			},
		},
		{
			name:  "multiple line-breaks inside record",
			input: "dn: dc=mokapi, dc=io\n\n\n\n\ncn: foo",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.EqualError(t, err, "no DN set at line 5: cn: foo")
			},
		},
		{
			name:  "multi line value",
			input: "dn: dc=mokapi, dc=io\ndescription: foo \n bar",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &AddRecord{
					Dn:         "dc=mokapi,dc=io",
					Attributes: map[string][]string{"description": {"foo bar"}},
				}, ld.Records[0])
			},
		},
		{
			name:  "multi line key",
			input: "dn: dc=mokapi, dc=io\nattr\n ibute: foo",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &AddRecord{
					Dn:         "dc=mokapi,dc=io",
					Attributes: map[string][]string{"attribute": {"foo"}},
				}, ld.Records[0])
			},
		},
		{
			name:  "empty record",
			input: "dn: dc=mokapi, dc=io\n# comment",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &AddRecord{
					Dn: "dc=mokapi,dc=io",
				}, ld.Records[0])
			},
		},
		{
			name:  "comment",
			input: "dn: dc=mokapi, dc=io\n# comment\nfoo: bar",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &AddRecord{
					Dn:         "dc=mokapi,dc=io",
					Attributes: map[string][]string{"foo": {"bar"}},
				}, ld.Records[0])
			},
		},
		{
			name:  "version set",
			input: "version: 1\ndn: dc=mokapi, dc=io",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &AddRecord{
					Dn: "dc=mokapi,dc=io",
				}, ld.Records[0])
			},
		},
		{
			name:  "empty value",
			input: "dn: dc=mokapi, dc=io\nfoo:",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &AddRecord{
					Dn:         "dc=mokapi,dc=io",
					Attributes: map[string][]string{"foo": {""}},
				}, ld.Records[0])
			},
		},
		{
			name:  "change type: add",
			input: "dn: dc=mokapi, dc=io\nchangetype: add",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &AddRecord{
					Dn: "dc=mokapi,dc=io",
				}, ld.Records[0])
			},
		},
		{
			name:  "change type: delete",
			input: "dn: dc=mokapi, dc=io\nchangetype: delete",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &DeleteRecord{
					Dn: "dc=mokapi,dc=io",
				}, ld.Records[0])
			},
		},
		{
			name:  "change type: modify",
			input: "dn: dc=mokapi, dc=io\nchangetype: modify",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &ModifyRecord{
					Dn: "dc=mokapi,dc=io",
				}, ld.Records[0])
			},
		},
		{
			name:  "modify: add",
			input: "dn: dc=mokapi, dc=io\nchangetype: modify\nadd: description\ndescription: first\ndescription: second\n-",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &ModifyRecord{
					Dn: "dc=mokapi,dc=io",
					Actions: []*ModifyAction{
						{
							Type:       "add",
							Name:       "description",
							Attributes: map[string][]string{"description": {"first", "second"}},
						},
					},
				}, ld.Records[0])
			},
		},
		{
			name:  "modify: replace",
			input: "dn: dc=mokapi, dc=io\nchangetype: modify\nreplace: postalCode\npostalCode: 12345\n-",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &ModifyRecord{
					Dn: "dc=mokapi,dc=io",
					Actions: []*ModifyAction{
						{
							Type:       "replace",
							Name:       "postalCode",
							Attributes: map[string][]string{"postalCode": {"12345"}},
						},
					},
				}, ld.Records[0])
			},
		},
		{
			name:  "modify: delete",
			input: "dn: dc=mokapi, dc=io\nchangetype: modify\ndelete: seeAlso\n-",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &ModifyRecord{
					Dn: "dc=mokapi,dc=io",
					Actions: []*ModifyAction{
						{
							Type: "delete",
							Name: "seeAlso",
						},
					},
				}, ld.Records[0])
			},
		},
		{
			name:  "file",
			input: "dn: dc=mokapi, dc=io\nphoto:< file:///path/to/photo.jpg",
			reader: &dynamictest.Reader{
				Data: map[string]*dynamic.Config{
					"file:///path/to/photo.jpg": {
						Raw: []byte("1234"),
					},
				},
			},
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &AddRecord{
					Dn: "dc=mokapi,dc=io",
					Attributes: map[string][]string{
						"photo": {"1234"},
					},
				}, ld.Records[0])
			},
		},
		{
			name:  "base64",
			input: "dn: dc=mokapi, dc=io\ndescription:: Zm9vYmFy",
			test: func(t *testing.T, ld *Ldif, err error) {
				require.NoError(t, err)
				require.Len(t, ld.Records, 1)
				require.Equal(t, &AddRecord{
					Dn: "dc=mokapi,dc=io",
					Attributes: map[string][]string{
						"description": {"foobar"},
					},
				}, ld.Records[0])
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reader := tc.reader
			if reader == nil {
				reader = &dynamictest.Reader{}
			}

			ld := &Ldif{}
			cfg := &dynamic.Config{Raw: []byte(tc.input)}
			err := ld.Parse(cfg, reader)
			tc.test(t, ld, err)
		})
	}
}

func TestConfig_LDIF(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		reader dynamic.Reader
		test   func(t *testing.T, c *Config, err error)
	}{
		{
			name:  "one ldif file",
			input: `{ "files": [ "./config.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/config.ldif": {Raw: []byte("dn: dc=mokapi,dc=io\nfoo: bar")},
			}},
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 3, c.Entries.Len())
				require.Equal(t, "dc=mokapi,dc=io", c.Entries.Lookup("dc=mokapi,dc=io").Dn)
				require.Equal(t, []string{"bar"}, c.Entries.Lookup("dc=mokapi,dc=io").Attributes["foo"])
			},
		},
		{
			name:  "two files modify add",
			input: `{ "files": [ "./config1.ldif", "./config2.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/config1.ldif": {Raw: []byte("dn: dc=mokapi,dc=io\nfoo: bar")},
				"file:/config2.ldif": {Raw: []byte("dn: dc=mokapi,dc=io\nchangetype: modify\nadd: foo\nfoo: yuh\n-")},
			}},
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 3, c.Entries.Len())
				require.Equal(t, []string{"bar", "yuh"}, c.Entries.Lookup("dc=mokapi,dc=io").Attributes["foo"])
			},
		},
		{
			name:  "modify delete specific value",
			input: `{ "files": [ "./config1.ldif", "./config2.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/config1.ldif": {Raw: []byte("dn: dc=mokapi,dc=io\nfoo: bar1\nfoo: bar2")},
				"file:/config2.ldif": {Raw: []byte("dn: dc=mokapi,dc=io\nchangetype: modify\ndelete: foo\nfoo: bar2")},
			}},
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 3, c.Entries.Len())
				e, ok := c.Entries.Get("dc=mokapi,dc=io")
				require.True(t, ok)
				require.Contains(t, e.Attributes, "foo")
				require.Equal(t, "bar1", e.Attributes["foo"][0])
			},
		},
		{
			name:  "modify delete",
			input: `{ "files": [ "./config1.ldif", "./config2.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/config1.ldif": {Raw: []byte("dn: dc=mokapi,dc=io\nfoo: bar")},
				"file:/config2.ldif": {Raw: []byte("dn: dc=mokapi,dc=io\nchangetype: modify\ndelete: foo\n")},
			}},
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 3, c.Entries.Len())
				require.Len(t, c.Entries.Lookup("dc=mokapi,dc=io").Attributes, 0)
			},
		},
		{
			name:  "two files modify replace",
			input: `{ "files": [ "./config1.ldif", "./config2.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/config1.ldif": {Raw: []byte("dn: dc=mokapi,dc=io\nfoo: bar")},
				"file:/config2.ldif": {Raw: []byte("dn: dc=mokapi,dc=io\nchangetype: modify\nreplace: foo\nfoo: yuh")},
			}},
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 3, c.Entries.Len())
				require.Equal(t, []string{"yuh"}, c.Entries.Lookup("dc=mokapi,dc=io").Attributes["foo"])
			},
		},
		{
			name:  "set Root DSE",
			input: `{ "files": [ "./config1.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/config1.ldif": {Raw: []byte("dn:\nvendorName: foo")},
			}},
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 2, c.Entries.Len())
				require.Equal(t, []string{"foo"}, c.Entries.Lookup("").Attributes["vendorName"])
			},
		},
		{
			name:  "delete entry",
			input: `{ "files": [ "./config1.ldif", "./config2.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/config1.ldif": {Raw: []byte("dn: cn=foo")},
				"file:/config2.ldif": {Raw: []byte("dn: cn=foo\nchangetype: delete")},
			}},
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 2, c.Entries.Len())
				_, ok := c.Entries.Get("cn=foo")
				require.False(t, ok)
			},
		},
		{
			name:  "copy entry",
			input: `{ "files": [ "./config1.ldif", "./config2.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/config1.ldif": {Raw: []byte("dn: cn=foo")},
				"file:/config2.ldif": {Raw: []byte("dn: cn=foo\nchangetype: modrdn\nnewrdn: cn=bar\ndeleteoldrdn: 0")},
			}},
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 4, c.Entries.Len())
				_, ok := c.Entries.Get("cn=foo")
				require.True(t, ok, "cn=foo")
				_, ok = c.Entries.Get("cn=bar")
				require.True(t, ok, "cn=bar")
			},
		},
		{
			name:  "rename entry",
			input: `{ "files": [ "./config1.ldif", "./config2.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/config1.ldif": {Raw: []byte("dn: cn=foo")},
				"file:/config2.ldif": {Raw: []byte("dn: cn=foo\nchangetype: modrdn\nnewrdn: cn=bar\ndeleteoldrdn: 1")},
			}},
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 3, c.Entries.Len())
				_, ok := c.Entries.Get("cn=foo")
				require.False(t, ok, "cn=foo")
				_, ok = c.Entries.Get("cn=bar")
				require.True(t, ok, "cn=bar")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var c *Config
			err := json.Unmarshal([]byte(tc.input), &c)
			if err != nil {
				tc.test(t, c, err)
			} else {
				err = c.Parse(&dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("file:/foo.yml")}}, tc.reader)
				tc.test(t, c, err)
			}
		})
	}
}
