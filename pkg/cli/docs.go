package cli

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const markdownExtension = ".md"

// DefaultGroup is the group of flags that were registered without one.
const DefaultGroup = "General"

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// slug turns a flag name into an anchor that is valid in a URL fragment and in a CSS selector,
// e.g. providers-git-repositories[<index>]-url becomes providers-git-repositories-index-url.
func slug(name string) string {
	return strings.Trim(nonAlphanumeric.ReplaceAllString(strings.ToLower(name), "-"), "-")
}

func (c *Command) GenMarkdown(dir string) error {
	basename := strings.ReplaceAll(c.Name, " ", "_") + markdownExtension
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return c.WriteMarkdown(f)
}

func (c *Command) WriteMarkdown(out io.StringWriter) error {
	if c.Name == "" {
		return fmt.Errorf("command must specify a name")
	}

	page := &strings.Builder{}
	fmt.Fprintf(page, `---
title: %s CLI Flags
description: A complete list of all %s flags, with descriptions, defaults, and examples of how to set the option in config file, environment variables, or CLI.
---`+"\n\n", c.Name, c.Name)
	writeCommand(page, &docSection{cmd: c, envPrefix: c.EnvPrefix})

	_, err := out.WriteString(page.String())
	return err
}

// docSection is a command in the generated page: the root command or, one level deeper, a child command.
type docSection struct {
	cmd *Command
	// level is the number of heading levels the section is nested deeper than the root command.
	level        int
	envPrefix    string
	anchorPrefix string
}

func (s *docSection) heading(level int) string {
	return strings.Repeat("#", level+s.level) + " "
}

func (s *docSection) anchor(name string) string {
	return slug(s.anchorPrefix + name)
}

func (s *docSection) child(c *Command) *docSection {
	return &docSection{
		cmd:          c,
		level:        s.level + 1,
		envPrefix:    cmp.Or(c.EnvPrefix, s.envPrefix),
		anchorPrefix: slug(c.Name) + "-",
	}
}

// writeAnchoredHeading writes a heading whose anchor has an id for links and CSS selectors and a name for old browsers.
func writeAnchoredHeading(w *strings.Builder, level string, anchor, text string) {
	fmt.Fprintf(w, "%s<a id=\"%s\" name=\"%s\"></a>%s\n\n", level, anchor, anchor, text)
}

func writeCommand(w *strings.Builder, s *docSection) {
	c := s.cmd
	if s.level == 0 {
		fmt.Fprintf(w, "# %s\n\n", c.Name)
	} else {
		writeAnchoredHeading(w, s.heading(1), slug(c.Name), c.Name)
	}

	if c.Short != "" {
		fmt.Fprintf(w, "%s\n\n", c.Short)
	}
	if c.Long != "" {
		fmt.Fprintf(w, "%sDescription\n\n%s\n\n", s.heading(2), c.Long)
	}
	if c.Use != "" {
		fmt.Fprintf(w, "%sUsage\n\n```bash tab=Bash\n%s\n```\n\n", s.heading(2), c.Use)
	}
	if c.Example != "" {
		fmt.Fprintf(w, "%sExample\n\n```bash\n%s\n```\n\n", s.heading(2), dedent(c.Example))
	}

	writeCommandsTable(w, c.Commands)
	writeFlags(w, s)
	for _, child := range c.Commands {
		writeCommand(w, s.child(child))
	}
}

func writeCommandsTable(w *strings.Builder, commands []*Command) {
	if len(commands) == 0 {
		return
	}
	w.WriteString("## Commands\n\n| Command | Description |\n|---------|-------------|\n")
	for _, child := range commands {
		fmt.Fprintf(w, "| [%s](#%s) | %s |\n", child.Name, slug(child.Name), escapeCell(child.Short))
	}
	w.WriteString("\n")
}

func writeFlag(w *strings.Builder, f *Flag, variants []*Flag, s *docSection) {
	writeAnchoredHeading(w, s.heading(2), s.anchor(f.Name), markdownKey(f.Name))

	if description := cmp.Or(f.Long, f.Short); description != "" {
		fmt.Fprintf(w, "%s\n\n", description)
	}

	env := s.envPrefix + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
	key := markdownKey(configKey(s.cmd.Config, f.Name))
	if f.Shorthand != "" {
		w.WriteString("| Flag | Shorthand | Env | Config file | Type | Default |\n|------|:---------:|------|-------------|:----:|:-------:|\n")
		fmt.Fprintf(w, "| --%s | -%s | %s | %s | %s | %s |\n", f.Name, f.Shorthand, env, key, f.Type, defaultValueOf(f))
	} else {
		w.WriteString("| Flag | Env | Config file | Type | Default |\n|------|------|-------------|:----:|:-------:|\n")
		fmt.Fprintf(w, "| --%s | %s | %s | %s | %s |\n", f.Name, env, key, f.Type, defaultValueOf(f))
	}
	w.WriteString("\n")

	writeExamples(w, f.Examples, s)

	if len(f.Aliases) > 0 {
		fmt.Fprintf(w, "%sAliases\n\n", s.heading(3))
		for _, alias := range f.Aliases {
			fmt.Fprintf(w, "- %s\n", alias)
		}
		w.WriteString("\n")
	}

	writeVariants(w, variants, s)
}

// dedent removes the blank lines around text and the indentation that all lines have in common.
func dedent(text string) string {
	lines := strings.Split(strings.Trim(text, "\n"), "\n")
	indent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if n := len(line) - len(strings.TrimLeft(line, " \t")); indent < 0 || n < indent {
			indent = n
		}
	}
	for i, line := range lines {
		if len(line) >= indent && indent > 0 {
			lines[i] = line[indent:]
		}
	}
	return strings.Join(lines, "\n")
}

func defaultValueOf(f *Flag) string {
	switch v := f.DefaultValue.(type) {
	case nil:
		return "-"
	case []string:
		if len(v) == 0 {
			return "-"
		}
		return strings.Join(v, ", ")
	case string:
		return cmp.Or(v, "-")
	}
	return fmt.Sprintf("%v", f.DefaultValue)
}

func writeExamples(w *strings.Builder, examples []Example, s *docSection) {
	for _, example := range examples {
		if example.Title != "" {
			fmt.Fprintf(w, "%s%s\n\n", s.heading(3), example.Title)
		}
		if example.Description != "" {
			fmt.Fprintf(w, "%s\n\n", example.Description)
		}
		for _, code := range example.Codes {
			language := cmp.Or(code.Language, "bash")
			if code.Title != "" {
				fmt.Fprintf(w, "```%s tab=%s\n%s\n```\n", language, code.Title, code.Source)
			} else {
				fmt.Fprintf(w, "```%s\n%s\n```\n", language, code.Source)
			}
		}
		w.WriteString("\n")
	}
}

func writeConfigFileOption(w *strings.Builder, o configFileOption, s *docSection) {
	writeAnchoredHeading(w, s.heading(2), slug(o.key), o.key)
	if description := cmp.Or(o.doc.Long, o.doc.Short); description != "" {
		fmt.Fprintf(w, "%s\n\n", description)
	}
	writeExamples(w, o.doc.Examples, s)
}

func writeVariants(w *strings.Builder, variants []*Flag, s *docSection) {
	if len(variants) == 0 {
		return
	}

	w.WriteString(s.heading(3) + "Sub-options\n\n| Flag | Config file | Type | Description |\n|------|-------------|:----:|-------------|\n")
	for _, v := range variants {
		anchor := s.anchor(v.Name)
		fmt.Fprintf(w, "| <a id=\"%s\" name=\"%s\"></a>`--%s` | %s | %s | %s |\n",
			anchor, anchor, v.Name, markdownKey(configKey(s.cmd.Config, v.Name)), v.Type, escapeCell(v.Short))
	}
	w.WriteString("\n")

	// the table shows only the short description, so details of a variant must not get lost
	for _, v := range variants {
		if v.Long == "" && len(v.Examples) == 0 {
			continue
		}
		fmt.Fprintf(w, "%s`--%s`\n\n", s.heading(4), v.Name)
		if v.Long != "" {
			fmt.Fprintf(w, "%s\n\n", v.Long)
		}
		writeExamples(w, v.Examples, s)
	}
}

func writeFlags(w *strings.Builder, s *docSection) {
	c := s.cmd
	flags := c.Flags()

	w.WriteString("<div class=\"flags\">\n\n")
	w.WriteString(s.heading(2) + "Flags\n\n")

	details := &strings.Builder{}
	for _, g := range groupFlags(flags) {
		fmt.Fprintf(w, "%s%s\n\n| Flag | Description |\n|------|-------------|\n", s.heading(3), g.name)
		for _, f := range g.flags {
			fmt.Fprintf(w, "| --[%s](#%s) | %s |\n", f.Name, s.anchor(f.Name), escapeCell(f.Short))
			writeFlag(details, f.Flag, f.variants, s)
		}
		w.WriteString("\n")
	}

	writeConfigFileOptionRows(w, c.configFileOptions, s)
	for _, o := range c.configFileOptions {
		writeConfigFileOption(details, o, s)
	}

	w.WriteString("</div>\n\n")
	w.WriteString(details.String())
}

func writeConfigFileOptionRows(w *strings.Builder, options []configFileOption, s *docSection) {
	if len(options) == 0 {
		return
	}
	w.WriteString(s.heading(3) + "Config file only\n\n| Option | Description |\n|--------|-------------|\n")
	for _, o := range options {
		fmt.Fprintf(w, "| [%s](#%s) | %s |\n", o.key, slug(o.key), escapeCell(o.doc.Short))
	}
	w.WriteString("\n")
}

type flagGroup struct {
	name  string
	flags []*documentedFlag
}

// documentedFlag is a flag with its indexed variants, e.g. providers-git-repositories with
// providers-git-repositories[<index>]-url.
type documentedFlag struct {
	*Flag
	variants []*Flag
}

// groupFlags lists the flags per group. Groups are in order of their first flag, flags in
// registration order. Flags without a group, e.g. help, are listed last in DefaultGroup.
func groupFlags(flags *FlagSet) []*flagGroup {
	var all []*Flag
	done := map[string]bool{}
	_ = flags.VisitAll(func(f *Flag) error {
		// skip aliases
		if !done[f.Name] {
			done[f.Name] = true
			all = append(all, f)
		}
		return nil
	})
	slices.SortStableFunc(all, func(f1, f2 *Flag) int { return f1.Order - f2.Order })

	byName := map[string]*documentedFlag{}
	var documented []*documentedFlag
	var variants []*Flag
	for _, f := range all {
		if isVariant(f) {
			variants = append(variants, f)
			continue
		}
		d := &documentedFlag{Flag: f}
		byName[f.Name] = d
		documented = append(documented, d)
	}
	for _, v := range variants {
		if parent, ok := byName[parentName(v.Name)]; ok {
			parent.variants = append(parent.variants, v)
		} else {
			documented = append(documented, &documentedFlag{Flag: v})
		}
	}

	var groups []*flagGroup
	add := func(name string, f *documentedFlag) {
		i := slices.IndexFunc(groups, func(g *flagGroup) bool { return g.name == name })
		if i < 0 {
			groups = append(groups, &flagGroup{name: name})
			i = len(groups) - 1
		}
		groups[i].flags = append(groups[i].flags, f)
	}
	for _, f := range documented {
		if f.Group != "" {
			add(f.Group, f)
		}
	}
	for _, f := range documented {
		if f.Group == "" {
			add(DefaultGroup, f)
		}
	}
	return groups
}

func isVariant(f *Flag) bool {
	return strings.ContainsAny(f.Name, "[<")
}

// parentName returns the flag a variant belongs to:
// providers-git-repositories[<index>]-url -> providers-git-repositories,
// event-store-<name>-size -> event-store.
func parentName(variant string) string {
	name, _, _ := strings.Cut(variant, "[")
	name, _, _ = strings.Cut(name, "<")
	return strings.TrimSuffix(name, "-")
}

// escapeCell keeps a pipe in a description, e.g. (debug|info), from ending the table cell.
func escapeCell(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

// markdownKey escapes the placeholders of a name or config file key, e.g. event.store.<name>.size.
func markdownKey(key string) string {
	return strings.NewReplacer("<", "&lt;", ">", "&gt;").Replace(key)
}
