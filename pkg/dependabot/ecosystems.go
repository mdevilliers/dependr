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
		files map[string]string
	}
)

var (
	//nolint:lll
	// https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file#package-ecosystem
	// https://docs.github.com/en/code-security/supply-chain-security/understanding-your-software-supply-chain/about-the-dependency-graph#supported-package-ecosystems
	wellKnown = ecosystems{
		files: map[string]string{
			"Gemfile.lock":        "bundler",
			"Gemfile":             "bundler", // TODO: support *.gemspec
			"Cargo.toml":          "cargo",
			"Cargo.lock":          "cargo",
			"composer.json":       "composer",
			"composer.lock":       "composer",
			"Dockerfile":          "docker", // TODO: support 'artisinal' dockerfile names
			"mix.exs":             "hex",
			"elm-package.json":    "elm",
			".gitmodules":         "gitsubmodule",
			"go.mod":              "gomod",
			"go.sum":              "gomod",
			"build.gradle":        "gradel",
			"pom.xml":             "maven",
			"package-lock.json":   "npm",
			"package.json":        "npm",
			"yarn.lock":           "npm",
			".csproj":             "nuget",
			".vbproj":             "nuget",
			".nuspec":             "nuget",
			".vcxproj":            "nuget",
			".fsproj":             "nuget",
			"packages.config":     "nuget",
			"requirements.txt":    "pip",
			"pipfile":             "pip",
			"pipfile.lock":        "pip",
			"setup.py":            "pip",
			".terraform.lock.hcl": "terraform",
		},
	}
)

func (n *node) Scan() error {

	updates := map[string]Update{}

	// walk the file system looking for well known files
	// append updates as required
	err := filepath.Walk(n.repo.root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.Contains(path, "node_modules") {
				return nil
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

			v, found := wellKnown.files[info.Name()]
			if found {
				rel, err := filepath.Rel(n.repo.root, path)
				if err != nil {
					return err
				}
				rel = filepath.Dir(rel)
				update := newUpdate(v, dot(rel))
				updates[key(update)] = update
			}

			return nil
		})

	if err != nil {
		return errors.Wrap(err, "error iterating root folder")
	}

	// check for github actions
	githubActionsPath := filepath.Join(n.repo.root, ".github/workflows/")
	if pathExists(githubActionsPath) {

		update := newUpdate("github-actions", ".")
		updates[key(update)] = update

	}

	fmt.Println(n, updates)
	return nil
}

func newUpdate(ecosystem, directory string) Update {
	return Update{
		PackageEcoSystem: ecosystem,
		Directory:        directory,
		Schedule: Schedule{
			Interval: "weekly",
		},
	}
}

func key(u Update) string {
	return fmt.Sprintf("%s%s", u.PackageEcoSystem, u.Directory)
}
