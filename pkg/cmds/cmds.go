package cmds

import (
	"github.com/urfave/cli/v2"
)

// Commands returns all registered commands
func Commands() []*cli.Command {
	return []*cli.Command{
		sniffCmd(),
	}
}
