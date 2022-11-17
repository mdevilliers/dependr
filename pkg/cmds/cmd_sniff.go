package cmds

import "github.com/urfave/cli/v2"

func sniffCmd() *cli.Command {
	return &cli.Command{
		Name:  "sniff",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {

			// load file (or return base file in none exists) , preserving comments
			// look for evidence of eco-systems
			// add a default configuration
			// write file

			return nil
		},
	}
}
