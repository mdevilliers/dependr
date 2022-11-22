package dependabot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type (
	ecosystems struct {
		files map[string][]string
	}
)

var (
	//nolint:lll
	// https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file#package-ecosystem
	// https://docs.github.com/en/code-security/supply-chain-security/understanding-your-software-supply-chain/about-the-dependency-graph#supported-package-ecosystems
	wellKnown = ecosystems{
		files: map[string][]string{
			"bundler":  {"Gemfile.lock", "Gemfile"}, // TODO: support *.gemspec
			"cargo":    {"Cargo.toml", "Cargo.lock"},
			"composer": {"composer.json", "composer.lock"},
			"docker":   {"Dockerfile"}, // TODO: support other weird dockerfile names
			// "hex" : []string{},
			// "elm" : []string{},
			// * "gitsubmodule" : []string{},
			// * "github-actions" : []string{},
			"gomod": {"go.mod", "go.sum"},
			// * "gradle" : []string{} ,
			// * "maven" : []string{},
			// * "npm" : []string{},
			// "nuget" : []string{},
			// * "pip" : []string{},
			// * "terraform" : []string{},
		},
	}
)

func (n *node) Scan() error {

	updates := map[string]Update{}

	// walk the file system looking for weel known files
	// append updates as required
	err := filepath.Walk(n.root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.Contains(path, ".git/") || strings.HasSuffix(path, ".git") {
				return nil
			}
			if info.IsDir() {
				return nil
			}

			dot := func(i string) string {
				if i == "" {
					return "."
				}
				return fmt.Sprintf("./%s", i)
			}

			key := func(u Update) string {
				return fmt.Sprintf("%s%s", u.PackageEcoSystem, u.Directory)
			}

			for k, v := range wellKnown.files {
				for _, wk := range v {
					if wk == info.Name() {
						rel, err := filepath.Rel(n.root, path)
						if err != nil {
							return err
						}
						rel = filepath.Dir(rel)

						update := Update{
							PackageEcoSystem: k,
							Directory:        dot(rel),
						}
						update.Schedule.Interval = "weekly"
						updates[key(update)] = update

					}
				}
			}

			return nil
		})

	if err != nil {
		return errors.Wrap(err, "error iterating root folder")
	}

	fmt.Println(updates)
	return nil
}
