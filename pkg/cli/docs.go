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
		use := fmt.Sprintf("```bash\n%s\n```\n\n", c.Use)
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

type flagGroup struct {
	Group    string
	Subgroup string
	Flags    []*Flag
}

func groupFlags(flags *FlagSet) ([]flagGroup, error) {
	type key struct {
		group    string
		subgroup string
	}

	mFlags := map[*Flag]bool{}
	flagList := []*Flag{}
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
	var groups []flagGroup

	for _, f := range flagList {
		g, sg := splitFlagName(f.Name)
		k := key{g, sg}
		m[k] = append(m[k], f)
		order = append(order, k)
	}

	for _, k := range order {
		groups = append(groups, flagGroup{
			Group:    k.group,
			Subgroup: k.subgroup,
			Flags:    m[k],
		})
	}

	return groups, nil
}

func renderFlagGroup(groups []flagGroup) (string, error) {
	var sb strings.Builder
	currentGroup := ""
	for _, g := range groups {

		if g.Group != currentGroup {
			if _, err := sb.WriteString(fmt.Sprintf("### %s\n\n", g.Group)); err != nil {
				return "", err
			}
			currentGroup = g.Group
		}

		if g.Subgroup != "" {
			if _, err := sb.WriteString(fmt.Sprintf("#### %s\n\n", g.Subgroup)); err != nil {
				return "", err
			}
		}

		if _, err := sb.WriteString(`| Flag | Shorthand | Type | Default | Usage |
|------|-----------|------|---------|-------|
`,
		); err != nil {
			return "", err
		}
		for _, flag := range g.Flags {
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
			if _, err := sb.WriteString(fmt.Sprintf("| --%s | %s | %s | %s | %s |\n", flag.Name, short, flag.Value.Type(), defaultValue, flag.Usage)); err != nil {
				return "", err
			}
		}
		if _, err := sb.WriteString("\n"); err != nil {
			return "", err
		}
	}

	return sb.String(), nil
}

func splitFlagName(name string) (group, subgroup string) {
	parts := strings.Split(name, "-")

	switch len(parts) {
	case 1:
		return "General", ""
	default:
		return caser.String(parts[0]), caser.String(parts[1])
	}
}
