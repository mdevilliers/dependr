package cmds

import (
	"github.com/mdevilliers/depender/pkg/dependabot"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func scanCmd() *cli.Command {
	return &cli.Command{
		Name: "scan",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "create-if-missing",
				Aliases: []string{"c"},
				Usage:   "create dependabot.yml file if missing. Defaults to .github/dependabot.yml path",
			},
		},
		Action: func(c *cli.Context) error {
			path := c.Args().First()
			create := c.Value("create-if-missing").(bool)

			type scanner interface {
				Scan() error
			}
			var s scanner
			var err error

			if create {
				s, err = dependabot.LoadOrCreate(path)
			} else {
				s, err = dependabot.Load(path)
			}

			if err != nil {
				return errors.Wrap(err, "error loading configuration")
			}
			return s.Scan()
		},
	}
}
