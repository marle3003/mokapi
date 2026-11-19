package mokapi_test

import (
	"mokapi/pkg/cli"
	"mokapi/pkg/cmd/mokapi"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var (
	anchorIds   = regexp.MustCompile(`<a id="([^"]*)"`)
	groupTitles = regexp.MustCompile(`(?m)^### (.+)$`)
	// | --name | [-x |] ENV | key | ...
	detailKeys = regexp.MustCompile(`(?m)^\| --([^\[\s]\S*) \| (?:-\w \| )?\S+ \| (\S+) \| `)
	// | <a id="..." name="..."></a>`--name` | key | ...
	variantKeys = regexp.MustCompile("(?m)^\\| <a id=\"[^\"]+\" name=\"[^\"]+\"></a>`--(\\S+)` \\| (\\S+) \\| ")
)

// fileOnlySections have an anchor but no flag.
var fileOnlySections = []string{"javascript-modules", "certificates"}

// flagsWithoutConfigKey exist only on the command line or are not documented for the config file.
var flagsWithoutConfigKey = []string{"help", "version", "config-file", "generate-cli-skeleton"}

func configKeys(docs string) map[string]string {
	keys := map[string]string{}
	for _, re := range []*regexp.Regexp{detailKeys, variantKeys} {
		for _, m := range re.FindAllStringSubmatch(docs, -1) {
			keys[m[1]] = m[2]
		}
	}
	return keys
}

// flagsWithoutGroup are created by the cli package itself.
var flagsWithoutGroup = []string{"help", "version"}

func generateDocs(t *testing.T) string {
	t.Helper()
	sb := &strings.Builder{}
	require.NoError(t, mokapi.NewCmdMokapi().WriteMarkdown(sb))
	return sb.String()
}

// overviewOf returns the part of the page between the flags heading and the end of the flags div.
// rootPage returns the page without the sections of the child commands.
func rootPage(docs string) string {
	root, _, _ := strings.Cut(docs, `<a id="sample-data" name="sample-data"></a>`)
	return root
}

func countFlags(t *testing.T, cmd *cli.Command) int {
	t.Helper()
	count := 0
	done := map[string]bool{}
	require.NoError(t, cmd.Flags().VisitAll(func(f *cli.Flag) error {
		if !done[f.Name] {
			done[f.Name] = true
			count++
		}
		return nil
	}))
	return count
}

// hasYamlKey reports whether the YAML sets the config file key. [] stands for a list item and <name> for any map key.
func hasYamlKey(source string, key string) bool {
	var current any
	if err := yaml.Unmarshal([]byte(source), &current); err != nil {
		return false
	}
	for _, segment := range strings.Split(strings.ReplaceAll(key, "[]", ".[]"), ".") {
		switch node := current.(type) {
		case []any:
			if segment != "[]" || len(node) == 0 {
				return false
			}
			current = node[0]
		case map[string]any:
			if segment == "<name>" && len(node) > 0 {
				for _, v := range node {
					current = v
				}
			} else if v, ok := node[segment]; ok {
				current = v
			} else {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func overviewOf(t *testing.T, docs string) string {
	t.Helper()
	_, rest, ok := strings.Cut(docs, "## Flags")
	require.True(t, ok, "page has no flags overview")
	overview, _, ok := strings.Cut(rest, "</div>")
	require.True(t, ok, "flags overview is not closed")
	return overview
}

func TestDocs_Consistency(t *testing.T) {
	t.Run("every flag belongs to a group", func(t *testing.T) {
		var missing []string
		err := mokapi.NewCmdMokapi().Flags().VisitAll(func(f *cli.Flag) error {
			if f.Group == "" && !slices.Contains(flagsWithoutGroup, f.Name) {
				missing = append(missing, f.Name)
			}
			return nil
		})
		require.NoError(t, err)
		require.Empty(t, missing)
	})

	t.Run("every flag has a config file key", func(t *testing.T) {
		docs := rootPage(generateDocs(t))
		keys := configKeys(docs)
		require.Len(t, keys, len(anchorIds.FindAllString(docs, -1))-len(fileOnlySections), "every flag needs a key column")

		var missing []string
		for name, key := range keys {
			if key == "-" && !slices.Contains(flagsWithoutConfigKey, name) {
				missing = append(missing, name)
			}
		}
		require.Empty(t, missing)
		for _, name := range flagsWithoutConfigKey {
			require.Equal(t, "-", keys[name], name)
		}
	})

	t.Run("every flag has a long description", func(t *testing.T) {
		var missing []string
		cmd := mokapi.NewCmdMokapi()
		for _, c := range append([]*cli.Command{cmd}, cmd.Commands...) {
			require.NoError(t, c.Flags().VisitAll(func(f *cli.Flag) error {
				// VisitAll also lists a flag under its shorthand and aliases
				if f.Long == "" && !slices.Contains(flagsWithoutGroup, f.Name) && !slices.Contains(missing, c.Name+" "+f.Name) {
					missing = append(missing, c.Name+" "+f.Name)
				}
				return nil
			}))
		}
		require.Empty(t, missing)
	})

	t.Run("singular and plural flags explain when to use which", func(t *testing.T) {
		pairs := [][2]string{
			{"config", "configs"},
			{"providers-file-filename", "providers-file-filenames"},
			{"providers-file-directory", "providers-file-directories"},
			{"providers-git-url", "providers-git-urls"},
			{"providers-git-repository", "providers-git-repositories"},
			{"providers-http-url", "providers-http-urls"},
			{"providers-npm-global-folder", "providers-npm-global-folders"},
			{"providers-npm-package", "providers-npm-packages"},
		}
		flags := mokapi.NewCmdMokapi().Flags()
		for _, pair := range pairs {
			require.Contains(t, flags.Lookup(pair[0]).Long, "--"+pair[1], pair[0])
			require.Contains(t, flags.Lookup(pair[1]).Long, "--"+pair[0], pair[1])
		}
	})

	t.Run("file examples use the config file key of the flag", func(t *testing.T) {
		keys := configKeys(rootPage(generateDocs(t)))
		var wrong []string
		require.NoError(t, mokapi.NewCmdMokapi().Flags().VisitAll(func(f *cli.Flag) error {
			key := strings.ReplaceAll(strings.ReplaceAll(keys[f.Name], "&lt;", "<"), "&gt;", ">")
			if key == "-" {
				return nil
			}
			for _, example := range f.Examples {
				for _, code := range example.Codes {
					if code.Title == "File" && !hasYamlKey(code.Source, key) && !slices.Contains(wrong, f.Name) {
						wrong = append(wrong, f.Name+" should set "+key)
					}
				}
			}
			return nil
		}))
		require.Empty(t, wrong)
	})

	t.Run("env examples name the env variable of the flag", func(t *testing.T) {
		var wrong []string
		require.NoError(t, mokapi.NewCmdMokapi().Flags().VisitAll(func(f *cli.Flag) error {
			variable := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			// an index or a name in a flag is only an example
			pattern := regexp.MustCompile("^MOKAPI_" + strings.NewReplacer(
				regexp.QuoteMeta("[<INDEX>]"), `\[\d+\]`,
				regexp.QuoteMeta("<NAME>"), `[A-Za-z]+`,
			).Replace(regexp.QuoteMeta(variable)) + "$")
			for _, example := range f.Examples {
				for _, code := range example.Codes {
					name, _, _ := strings.Cut(code.Source, "=")
					if code.Title == "Env" && !pattern.MatchString(name) && !slices.Contains(wrong, f.Name) {
						wrong = append(wrong, f.Name+" should use "+pattern.String()+" but uses "+name)
					}
				}
			}
			return nil
		}))
		require.Empty(t, wrong)
	})

	t.Run("file examples are yaml or json", func(t *testing.T) {
		var wrong []string
		require.NoError(t, mokapi.NewCmdMokapi().Flags().VisitAll(func(f *cli.Flag) error {
			for _, example := range f.Examples {
				for _, code := range example.Codes {
					if code.Title == "File" && code.Language != "yaml" && code.Language != "json" && !slices.Contains(wrong, f.Name) {
						wrong = append(wrong, f.Name)
					}
				}
			}
			return nil
		}))
		require.Empty(t, wrong)
	})

	t.Run("config file keys", func(t *testing.T) {
		keys := configKeys(rootPage(generateDocs(t)))
		require.Equal(t, "providers.http.tlsSkipVerify", keys["providers-http-tls-skip-verify"])
		require.Equal(t, "providers.git.repositories[].url", keys["providers-git-repositories[<index>]-url"])
		require.Equal(t, "providers.git.repositories[].auth.gitHub.appId", keys["providers-git-repositories[<index>]-auth-github-appid"])
		require.Equal(t, "event.store.&lt;name&gt;.size", keys["event-store-<name>-size"])
		require.Equal(t, "event.store.default.size", keys["event-store-default-size"])
		require.Equal(t, "rootCaCert", keys["root-ca-cert"])
		require.Equal(t, "data-gen.optionalProperties", keys["data-gen-optional-properties"])
		require.Equal(t, "configs", keys["config"])
	})

	t.Run("anchors are unique and free of <>[]", func(t *testing.T) {
		var ids []string
		for _, m := range anchorIds.FindAllStringSubmatch(generateDocs(t), -1) {
			ids = append(ids, m[1])
		}
		require.NotEmpty(t, ids)

		seen := map[string]bool{}
		for _, id := range ids {
			require.NotContains(t, id, "<", id)
			require.NotContains(t, id, ">", id)
			require.NotContains(t, id, "[", id)
			require.NotContains(t, id, "]", id)
			require.False(t, seen[id], "duplicate anchor %s", id)
			seen[id] = true
		}

		cmd := mokapi.NewCmdMokapi()
		expected := countFlags(t, cmd) + len(fileOnlySections) + len(cmd.Commands)
		for _, child := range cmd.Commands {
			expected += countFlags(t, child)
		}
		require.Equal(t, expected, len(ids), "every flag, command and file-only option needs exactly one anchor")
	})

	t.Run("groups are in registration order", func(t *testing.T) {
		overview := overviewOf(t, generateDocs(t))
		var groups []string
		for _, m := range groupTitles.FindAllStringSubmatch(overview, -1) {
			groups = append(groups, m[1])
		}
		require.Equal(t, []string{
			"Log",
			"API & Dashboard",
			"MCP Server",
			"Health",
			"File Provider",
			"HTTP Provider",
			"Git Provider",
			"NPM Provider",
			"Certificates",
			"Events Store",
			"Data Generator",
			"General",
			"Config file only",
		}, groups)
	})

	t.Run("indexed variants are not part of the overview", func(t *testing.T) {
		overview := overviewOf(t, generateDocs(t))
		require.NotContains(t, overview, "<index>")
		require.NotContains(t, overview, "&lt;index&gt;")
		require.NotContains(t, overview, "<name>")
		require.NotContains(t, overview, "&lt;name&gt;")
		require.Contains(t, overview, "--[providers-git-repositories](#providers-git-repositories)")
	})

	t.Run("sample-data is documented as command", func(t *testing.T) {
		docs := generateDocs(t)
		require.Contains(t, docs, "## Commands\n\n| Command | Description |\n|---------|-------------|\n| [sample-data](#sample-data) | Generate sample data from a schema |\n")
		require.Contains(t, docs, `<a id="sample-data" name="sample-data"></a>sample-data`)
		require.Contains(t, docs, "| --count | -n | MOKAPI_COUNT | - | int | 1 |\n")
		require.Contains(t, docs, `<a id="sample-data-count" name="sample-data-count"></a>count`)
	})

	t.Run("file-only options have a section", func(t *testing.T) {
		docs := generateDocs(t)
		require.Contains(t, docs, `<a id="javascript-modules" name="javascript-modules"></a>javascript.modules`)
		require.Contains(t, docs, `<a id="certificates" name="certificates"></a>certificates`)
		require.Contains(t, docs, "```yaml tab=File (YAML)\njavascript:")
		require.Contains(t, docs, "```json tab=File (JSON)\n{\n  \"certificates\"")
	})
}
