package pungi

import "github.com/spf13/cobra"

// Command defines one command
type Command struct {
	cmdName, usageText, desc string
	runnable                 Runnable
	keys                     map[string]*key
	args                     cobra.PositionalArgs
}

func (c *Command) Key(name string, value interface{}, desc string) *Command {
	c.keys[name] = &key{
		name:  name,
		value: value,
		desc:  desc,
	}
	return c
}

func (c *Command) Args(args cobra.PositionalArgs) *Command {
	c.args = args
	return c
}

func Cmd(usageText, desc string, runnable func(conf *Conf, args []string) error) *Command {
	return &Command{
		cmdName:   firstWord(usageText),
		usageText: usageText,
		desc:      desc,
		runnable:  runnable,
		keys:      make(map[string]*key),
	}
}
