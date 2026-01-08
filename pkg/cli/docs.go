package cli

import (
	"fmt"
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

	err = c.Flags().Visit(func(flag *Flag) error {
		if len(flag.Examples) == 0 {
			return nil
		}
		return writeFlag(flag, dir)
	})
	if err != nil {
		return err
	}

	return nil
}

func writeCommand(c *Command, f *os.File) error {
	if _, err := f.WriteString(fmt.Sprintf("# %s\n\n", c.Name)); err != nil {
		return err
	}

	if _, err := f.WriteString(fmt.Sprintf("> %s\n\n", c.Short)); err != nil {
		return err
	}

	if c.Long != "" {
		if _, err := f.WriteString(fmt.Sprintf("## Description\n\n %s\n\n", c.Long)); err != nil {
			return err
		}
	}

	if c.Use != "" {
		use := fmt.Sprintf("```bash\n%s\n```", c.Use)
		if _, err := f.WriteString(fmt.Sprintf("## Usage\n\n%s\n\n", use)); err != nil {
			return err
		}
	}

	groups, err := groupFlags(c.Flags())
	if err != nil {
		return err
	}
	if len(groups) > 0 {
		s, err := renderFlagGroup(groups)
		if err != nil {
			return err
		}
		if _, err = f.WriteString(s); err != nil {
			return err
		}
	}

	return nil
}

func writeFlag(f *Flag, dir string) error {
	basename := f.Name + markdownExtension
	filename := filepath.Join(dir, basename)
	w, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer w.Close()

	if _, err = w.WriteString(fmt.Sprintf("# %s\n\n", f.Name)); err != nil {
		return err
	}

	if f.Description != "" {
		if _, err = w.WriteString(fmt.Sprintf("%s\n\n", f.Description)); err != nil {
			return err
		}
	} else {
		if _, err = w.WriteString(fmt.Sprintf("%s\n\n", f.Usage)); err != nil {
			return err
		}
	}

	if _, err = w.WriteString("## Examples\n\n"); err != nil {
		return err
	}

	for _, example := range f.Examples {
		if example.Title != "" {
			if _, err = w.WriteString(fmt.Sprintf("### %s\n\n", example.Title)); err != nil {
				return err
			}
		}

		if example.Description != "" {
			if _, err = w.WriteString(fmt.Sprintf("%s\n\n", example.Description)); err != nil {
				return err
			}
		}

		for _, code := range example.Codes {
			var s string
			if code.Title != "" {
				s = fmt.Sprintf("```bash tab=%s\n%s\n```\n", code.Title, code.Source)
			} else {
				s = fmt.Sprintf("```bash\n%s\n```\n", code)
			}
			if _, err = w.WriteString(s); err != nil {
				return err
			}
		}
		if _, err = w.WriteString("\n"); err != nil {
			return err
		}
	}

	if len(f.Aliases) > 0 {
		if _, err = w.WriteString(fmt.Sprintf("## Aliases\n\n")); err != nil {
			return err
		}

		for _, alias := range f.Aliases {
			if _, err = w.WriteString(fmt.Sprintf("- %s", alias)); err != nil {
				return err
			}
		}

	}

	return nil
}

type flagGroup struct {
	Group     string
	Subgroups []flagGroup
	Flags     []*Flag
}

func groupFlags(flags *FlagSet) ([]*flagGroup, error) {
	type key struct {
		group    string
		subgroup string
	}

	mFlags := map[*Flag]bool{}
	var flagList []*Flag
	_ = flags.Visit(func(flag *Flag) error {
		_, ok := mFlags[flag]
		if !ok {
			mFlags[flag] = true
			flagList = append(flagList, flag)
		}
		return nil
	})

	m := map[key][]*Flag{}
	var order []key

	for _, f := range flagList {
		g, sg := splitFlagName(f.Name)
		k := key{g, sg}
		if _, ok := m[k]; !ok {
			order = append(order, k)
		}
		m[k] = append(m[k], f)
	}

	var groups []*flagGroup
	var currentGroup *flagGroup
	for _, k := range order {
		if currentGroup == nil || currentGroup.Group != k.group {
			currentGroup = &flagGroup{Group: k.group}
			if k.subgroup == "" {
				currentGroup.Flags = m[k]
			}
			groups = append(groups, currentGroup)
		}
		if k.subgroup != "" {
			currentGroup.Subgroups = append(currentGroup.Subgroups, flagGroup{
				Group: k.subgroup,
				Flags: m[k],
			})
		}
	}

	return groups, nil
}

func renderFlagGroup(groups []*flagGroup) (string, error) {
	var sb strings.Builder
	for _, g := range groups {

		if len(g.Flags) == 0 && len(g.Subgroups) == 1 {
			sub := g.Subgroups[0]
			sb.WriteString(fmt.Sprintf("### %s-%s\n\n", g.Group, sub.Group))
			s, err := renderFlags(sub.Flags)
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
			continue
		}

		if len(g.Flags) > 0 {
			sb.WriteString(fmt.Sprintf("### %s\n\n", g.Group))
			s, err := renderFlags(g.Flags)
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
		}

		for _, sub := range g.Subgroups {
			if len(sub.Flags) == 0 {
				continue
			}

			if sub.Group != "" {
				sb.WriteString(fmt.Sprintf("#### %s\n\n", sub.Group))
			}

			s, err := renderFlags(sub.Flags)
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
		}
	}

	return sb.String(), nil
}

func renderFlags(flags []*Flag) (string, error) {
	var sb strings.Builder
	if _, err := sb.WriteString(`| Flag | Shorthand | Type | Default | Usage |
|------|-----------|------|---------|-------|
`,
	); err != nil {
		return "", err
	}
	for _, flag := range flags {
		short := flag.Shorthand
		if short == "" {
			short = "-"
		} else {
			short = "-" + short
		}
		defaultValue := "-"
		if flag.DefaultValue != nil && flag.DefaultValue != "" {
			defaultValue = fmt.Sprintf("%v", flag.DefaultValue)
		}
		name := flag.Name
		if flag.Examples != nil {
			name = fmt.Sprintf("[%s](./%s.md)", name, name)
		}
		if _, err := sb.WriteString(fmt.Sprintf("| --%s | %s | %s | %s | %s |\n", name, short, flag.Value.Type(), defaultValue, flag.Usage)); err != nil {
			return "", err
		}
	}
	if _, err := sb.WriteString("\n"); err != nil {
		return "", err
	}

	return sb.String(), nil
}

func splitFlagName(name string) (group, subgroup string) {
	parts := strings.Split(name, "-")

	switch len(parts) {
	case 1:
		return "General", ""
	case 2:
		return fmt.Sprintf("%s", caser.String(parts[0])), ""
	default:
		return caser.String(parts[0]), caser.String(parts[1])
	}
}
