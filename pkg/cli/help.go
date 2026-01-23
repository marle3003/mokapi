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

	flags := c.Flags()
	if flags.Len() > 0 {
		_, _ = fmt.Fprintf(w, "\nFlags:")

		maxNameLen, hasShort := flagsInfo(flags)
		_ = flags.Visit(func(flag *Flag) error {

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
			if flag.Short != "" {
				space := strings.Repeat(" ", maxNameLen-len(flag.Name))
				_, _ = fmt.Fprintf(w, " %s  %s", space, flag.Short)
			}
			_, _ = fmt.Fprintln(w)
			return nil
		})
	}
}

func flagsInfo(flags *FlagSet) (int, bool) {
	maxNameLen := 0
	hasShort := false
	_ = flags.Visit(func(f *Flag) error {
		if len(f.Name) > maxNameLen {
			maxNameLen = len(f.Name)
		}
		if f.Shorthand != "" {
			hasShort = true
		}
		return nil
	})
	return maxNameLen, hasShort
}
