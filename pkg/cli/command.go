package cli

import (
	"fmt"
	"os"
	"slices"
)

type Command struct {
	Name      string
	Use       string
	Short     string
	Long      string
	Example   string
	Config    any
	Commands  []*Command
	Run       func(cmd *Command, args []string) error
	EnvPrefix string

	configFileName string
	configPaths    []string
	configFile     string
	args           []string
	flags          *FlagSet
}

func (c *Command) Execute() error {
	args := c.args
	if args == nil {
		args = os.Args[1:]
	}

	cmd := c
	envPrefix := c.EnvPrefix

	if len(args) > 0 {
		for _, child := range c.Commands {
			if child.Name == args[0] {
				cmd = child
				args = args[1:]
				if cmd.EnvPrefix != "" {
					envPrefix = cmd.EnvPrefix
				}
			}
		}
	}

	positional, err := parseFlags(args, envPrefix, cmd.Flags())
	if err != nil {
		return err
	}

	if cmd.Config != nil {
		err = c.readConfigFile()
		if err != nil {
			return err
		}
	}

	if cmd.Config != nil {
		b := flagConfigBinder{}
		err = b.Decode(cmd.Flags(), cmd.Config)
		if err != nil {
			return fmt.Errorf("failed to bind flags to config: %w", err)
		}
	}

	if cmd.Run != nil {
		return cmd.Run(cmd, positional)
	} else {
		return fmt.Errorf("no command run specified")
	}
}

func (c *Command) SetArgs(args []string) {
	c.args = args
}

func (c *Command) Flags() *FlagSet {
	if c.flags == nil {
		c.flags = &FlagSet{
			orderedFlags: make(map[string]int),
			setConfigFile: func(s string) {
				c.configFile = s
			},
		}
	}
	return c.flags
}

func (c *Command) SetConfigName(name string) {
	c.configFileName = name
}

func (c *Command) SetConfigFile(file string) {
	c.configFile = file
}

func (c *Command) SetConfigPath(path ...string) {
	for _, p := range path {
		if !slices.Contains(c.configPaths, p) {
			c.configPaths = append(c.configPaths, p)
		}
	}
}
