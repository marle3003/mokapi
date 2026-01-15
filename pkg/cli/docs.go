package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const markdownExtension = ".md"

func (c *Command) GenMarkdown(dir string) error {
	basename := strings.ReplaceAll(c.Name, " ", "_") + markdownExtension
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = writeCommand(c, f); err != nil {
		return err
	}

	return nil
}

func writeCommand(c *Command, f *os.File) error {
	if _, err := f.WriteString(fmt.Sprintf("# %s\n\n", c.Name)); err != nil {
		return err
	}

	if _, err := f.WriteString(fmt.Sprintf("%s\n\n", c.Short)); err != nil {
		return err
	}

	if c.Long != "" {
		if _, err := f.WriteString(fmt.Sprintf("## Description\n\n %s\n\n", c.Long)); err != nil {
			return err
		}
	}

	if c.Use != "" {
		use := fmt.Sprintf("```bash tab=Bash\n%s\n```\n\n", c.Use)
		if _, err := f.WriteString(fmt.Sprintf("## Usage\n\n%s\n\n", use)); err != nil {
			return err
		}
	}

	return writeFlags(f, c.Flags(), c.EnvPrefix)
}

func writeFlag(w io.StringWriter, f *Flag, envPrefix string) error {
	escapedName := strings.ReplaceAll(f.Name, "<", "&lt;")
	escapedName = strings.ReplaceAll(escapedName, ">", "&gt;")
	if _, err := w.WriteString(fmt.Sprintf("## <a name=%s></a>%s\n\n", escapedName, escapedName)); err != nil {
		return err
	}

	if f.Long != "" {
		if _, err := w.WriteString(fmt.Sprintf("%s\n\n", f.Long)); err != nil {
			return err
		}
	} else {
		if _, err := w.WriteString(fmt.Sprintf("%s\n\n", f.Short)); err != nil {
			return err
		}
	}

	if f.Shorthand != "" {
		if _, err := w.WriteString(`| Flag | Shorthand | Env  | Type | Default |
|------|:---------:|------|:----:|:-------:|
`); err != nil {
			return err
		}
	} else {
		if _, err := w.WriteString(`| Flag | Env  | Type | Default |
|------|------|:----:|:-------:|
`); err != nil {
			return err
		}
	}

	var defaultValue string
	switch v := f.DefaultValue.(type) {
	case []string:
		switch len(v) {
		case 0:
			defaultValue = "-"
		case 1:
			defaultValue = v[0]
		default:
			defaultValue = strings.Join(v, ", ")
		}
	case string:
		if v == "" {
			defaultValue = "-"
		} else {
			defaultValue = v
		}
	default:
		if v == nil {
			defaultValue = "-"
		} else {
			defaultValue = fmt.Sprintf("%v", v)
		}
	}

	env := fmt.Sprintf("%s%s", envPrefix, strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_")))

	if f.Shorthand != "" {
		if _, err := w.WriteString(fmt.Sprintf("| --%s | -%s | %s | %s | %s |\n", f.Name, f.Shorthand, env, f.Type, defaultValue)); err != nil {
			return err
		}
	} else {
		if _, err := w.WriteString(fmt.Sprintf("| --%s | %s | %s | %s |\n", f.Name, env, f.Type, defaultValue)); err != nil {
			return err
		}
	}

	if len(f.Examples) > 0 {
		for _, example := range f.Examples {
			if example.Title != "" {
				if _, err := w.WriteString(fmt.Sprintf("### %s\n\n", example.Title)); err != nil {
					return err
				}
			}

			if example.Description != "" {
				if _, err := w.WriteString(fmt.Sprintf("%s\n\n", example.Description)); err != nil {
					return err
				}
			}

			for _, code := range example.Codes {
				codeType := "bash"
				if code.Language != "" {
					codeType = code.Language
				}
				var s string
				if code.Title != "" {
					s = fmt.Sprintf("```%s tab=%s\n%s\n```\n", codeType, code.Title, code.Source)
				} else {
					s = fmt.Sprintf("```%s\n%s\n```\n", codeType, code)
				}
				if _, err := w.WriteString(s); err != nil {
					return err
				}
			}
			if _, err := w.WriteString("\n\n"); err != nil {
				return err
			}
		}
	}

	if len(f.Aliases) > 0 {
		if _, err := w.WriteString(fmt.Sprintf("### Aliases\n\n")); err != nil {
			return err
		}

		for _, alias := range f.Aliases {
			if _, err := w.WriteString(fmt.Sprintf("- %s\n", alias)); err != nil {
				return err
			}
		}

	}

	_, err := w.WriteString("\n")
	return err
}

func writeFlags(w io.StringWriter, flags *FlagSet, envPrefix string) error {
	if flags.Len() == 0 {
		return nil
	}

	if _, err := w.WriteString("<div class=\"flags\">\n\n"); err != nil {
		return err
	}

	if _, err := w.WriteString("## Flags \n\n"); err != nil {
		return err
	}

	if _, err := w.WriteString(`| Name | Usage |
|------|-------|
`,
	); err != nil {
		return err
	}

	details := &strings.Builder{}
	done := map[string]bool{}
	err := flags.VisitAll(func(f *Flag) error {
		// skip aliases
		if done[f.Name] {
			return nil
		}
		done[f.Name] = true

		_, err := w.WriteString(getFlagRow(f))
		if err != nil {
			return err
		}
		return writeFlag(details, f, envPrefix)
	})
	if err != nil {
		return err
	}

	if _, err = w.WriteString("\n\n</div>\n\n"); err != nil {
		return err
	}

	_, err = w.WriteString(details.String())
	return err
}

func getFlagRow(f *Flag) string {
	short := f.Shorthand
	if short == "" {
		short = "-"
	} else {
		short = "-" + short
	}
	name := fmt.Sprintf("[%s](#%s)", f.Name, f.Name)
	return fmt.Sprintf("| --%s | %s |\n", name, f.Short)
}
