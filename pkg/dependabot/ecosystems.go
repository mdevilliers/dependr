package dependabot

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type (
	ecosystems struct {
		files map[string]string
	}

	Doc struct {
		Updates []Update
	}
	Update struct {
		PackageEcoSystem string   `yaml:"package-ecosystem"`
		Directory        string   `yaml:"directory"`
		Schedule         Schedule `yaml:"schedule"`
	}
	Schedule struct {
		Interval string
	}

	Updates map[string]Update
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

func (n *node) Scan() error { //nolint:funlen

	updates := Updates{}

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
				if i == "." {
					return "/"
				}
				return fmt.Sprintf("/%s", i)
			}

			v, found := wellKnown.files[info.Name()]
			if found {
				rel, err := filepath.Rel(n.repo.root, path)
				if err != nil {
					return err
				}
				rel = filepath.Dir(rel)
				update := newDefaultUpdate(v, dot(rel))
				updates.Add(update)
			}

			return nil
		})

	if err != nil {
		return errors.Wrap(err, "error iterating root folder")
	}

	// check for github actions
	githubActionsPath := filepath.Join(n.repo.root, ".github/workflows/")
	if pathExists(githubActionsPath) {
		update := newDefaultUpdate("github-actions", "/")
		updates.Add(update)
	}

	var p yaml.Node

	if n.repo.dependabotFileExists {
		data, err := osReadFile(path.Join(n.repo.root, n.repo.dependabotFilePath))
		if err != nil {
			return errors.Wrapf(err, "error loading file: %s", n.repo.dependabotFilePath)
		}
		if err := yaml.Unmarshal(data, &p); err != nil {
			return errors.Wrapf(err, "error loading: %s", n.repo.dependabotFilePath)
		}

		// load existing file in to a Doc instance (loosing comments)
		var doc Doc
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return errors.Wrapf(err, "error loading: %s", n.repo.dependabotFilePath)
		}

		// iterate through Doc.Updates removing duplicates
		for _, u := range doc.Updates {
			updates.RemoveIfExists(u)
		}

		// append what is left...
		if err = updates.ApplyAllTo(&p); err != nil {
			return errors.Wrap(err, "error applying updates to existing files")
		}
	} else {

		// check if we need to do anything, return if nothing to do.
		if updates.Empty() {
			return nil
		}

		//nolint:lll
		data := `# To get started with Dependabot version updates, you'll need to specify which
# package ecosystems to update and where the package manifests are located.
# Please see the documentation for all configuration options:
# https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file

version: 2
updates:
`
		if err := yaml.Unmarshal([]byte(data), &p); err != nil {
			return errors.Wrapf(err, "error loading: %s", n.repo.dependabotFilePath)
		}

		if err = updates.ApplyAllTo(&p); err != nil {
			return errors.Wrap(err, "error applying updates to a new file")
		}
	}

	bytes, err := yaml.Marshal(&p)
	if err != nil {
		return errors.Wrap(err, "error marshalling yaml")
	}

	fullPath := filepath.Join(n.repo.root, n.repo.dependabotFilePath)

	// ensure directory exists
	dir := path.Dir(fullPath)
	if !pathExists(dir) {
		if err := osMkdirAll(dir, os.ModePerm); err != nil {
			return errors.Wrapf(err, "error creating folder : %s", dir)
		}
	}

	if err := osWriteFile(fullPath, bytes, 0600); err != nil { //nolint:gomnd
		return errors.Wrapf(err, "error writing dependabot file: %s", fullPath)
	}

	return nil
}

func newDefaultUpdate(ecosystem, directory string) Update {
	return Update{
		PackageEcoSystem: ecosystem,
		Directory:        directory,
		Schedule: Schedule{
			Interval: "weekly",
		},
	}
}

func (u Updates) Add(update Update) {
	key := fmt.Sprintf("%s%s", update.PackageEcoSystem, update.Directory)
	u[key] = update
}

func (u Updates) RemoveIfExists(update Update) {
	key := fmt.Sprintf("%s%s", update.PackageEcoSystem, update.Directory)
	_, found := u[key]
	if found {
		delete(u, key)
	}
}

func (u Updates) ToArray() []Update {
	all := []Update{}
	for _, v := range u {
		all = append(all, v)
	}
	return all
}

func (u Updates) Empty() bool {
	return len(u) == 0
}

func (u Updates) ApplyAllTo(n *yaml.Node) error {

	all := u.ToArray()

	// we need to convert the existing updates to
	// yaml and then parse again into Node(s)
	d, err := yaml.Marshal(all)
	if err != nil {
		return err
	}
	var yy yaml.Node
	if err = yaml.Unmarshal(d, &yy); err != nil {
		return err
	}

	root := n.Content[0]
	var previous *yaml.Node
	for i, child := range root.Content {
		if previous != nil && previous.Value == "updates" {
			// should have found the sequence (array) of existing updates
			if root.Content[i].Tag == "!!null" {
				root.Content[i] = yy.Content[0]
			} else {
				root.Content[i].Content = append(root.Content[i].Content, yy.Content[0].Content...)
			}
			return nil
		}
		previous = child
	}

	return nil
}
