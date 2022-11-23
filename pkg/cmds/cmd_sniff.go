package cmds

import (
	"github.com/mdevilliers/depender/pkg/dependabot"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func sniffCmd() *cli.Command {
	return &cli.Command{
		Name: "sniff",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "path",
				Aliases:  []string{"p"},
				Usage:    "path to dependabot file or root of project",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "create-if-missing",
				Aliases: []string{"c"},
				Usage:   "create dependabot.yml file if missing. Defaults to .github/dependabot.yml path",
			},
		},
		Action: func(c *cli.Context) error {
			path := c.Value("path").(string)
			create := c.Value("create-if-missing").(bool)

			// load file (or return base file in none exists) , preserving comments
			// TODO : look for evidence of eco-systems
			// TODO : add a default configuration
			// TODO : write file

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
