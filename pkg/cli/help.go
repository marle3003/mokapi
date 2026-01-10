package cli

import (
	"fmt"
	"os"
	"strings"
)

func (c *Command) printHelp() {
	w := c.output
	if c.output == nil {
		w = os.Stdout
	}

	if c.Long != "" {
		_, _ = fmt.Fprintf(w, "\n\n%s\n", c.Long)
	} else if c.Short != "" {
		_, _ = fmt.Fprintf(w, "\n\n%s\n", c.Short)
	}
	if c.Use != "" {
		_, _ = fmt.Fprintf(w, "\nUsage:\n  %s\n", c.Use)
	}

	groups, _ := groupFlags(c.Flags())
	if len(groups) > 0 {
		_, _ = fmt.Fprintf(w, "\nFlags:")

		for _, g := range groups {
			maxNameLen, hasShort := flagsInfo(g.Flags)
			for _, flag := range g.Flags {

				_, _ = fmt.Fprintln(w)
				if hasShort {
					short := flag.Shorthand
					if short != "" {
						short = fmt.Sprintf("-%s,", short)
					} else {
						short = strings.Repeat(" ", 3)
					}
					_, _ = fmt.Fprintf(w, "  %s", short)
				}
				_, _ = fmt.Fprintf(w, " --%s", flag.Name)
				if flag.Usage != "" {
					space := strings.Repeat(" ", maxNameLen-len(flag.Name))
					_, _ = fmt.Fprintf(w, " %s  %s", space, flag.Usage)
				}
			}
			_, _ = fmt.Fprintln(w)
		}
	}
}

func flagsInfo(flag []*Flag) (int, bool) {
	maxNameLen := 0
	hasShort := false
	for _, f := range flag {
		if len(f.Name) > maxNameLen {
			maxNameLen = len(f.Name)
		}
		if f.Shorthand != "" {
			hasShort = true
		}
	}
	return maxNameLen, hasShort
}
