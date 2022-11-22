package cmds

import (
	"fmt"
	"os"
	"path/filepath"

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
		},
		Action: func(c *cli.Context) error {
			path := c.Value("path").(string)

			// if path isn't absolute ensure that it is
			if !filepath.IsAbs(path) {
				wd, err := os.Getwd()
				if err != nil {
					return errors.Wrap(err, "error getting working directory")
				}
				path = filepath.Join(wd, path)
			}

			fmt.Println(path)
			// load file (or return base file in none exists) , preserving comments
			// look for evidence of eco-systems
			// add a default configuration
			// write file

			d, err := dependabot.Load(path)
			if err != nil {
				return err
			}
			err = d.Scan()

			return err
		},
	}
}
