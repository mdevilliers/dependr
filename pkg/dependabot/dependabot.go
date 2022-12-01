package dependabot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type (
	node struct {
		// repo holds the sanitised file information for the dependabot config file
		repo repo
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

	repo struct {
		// root is the absolute path to the folder containing filePath
		root string
		// dependabotFilePath is the local path (from the root) to the dependabot configuration
		dependabotFilePath   string
		dependabotFileExists bool
	}
)

var (
	ErrMissingConfigFile = errors.New("error finding dependabot config")

	// allow redirecting os functions for testing
	osReadFile  = os.ReadFile
	osStat      = os.Stat
	osGetwd     = os.Getwd
	osWriteFile = os.WriteFile

	// allow redirecting the git command for testing
	osGetRootFolder = getRootFolder
)

// Load will search the path for a dependabot file, returning an
// initialsed parsed config or an error
func Load(path string) (*node, error) {

	repo, err := newRepo(path)
	if err != nil {
		if repo == nil {
			return nil, errors.Wrapf(err, "error parsing supplied path %s", path)
		}
		return &node{repo: *repo}, errors.Wrapf(err, "error parsing supplied path %s", path)
	}
	return &node{
		repo: *repo,
	}, nil
}

// LoadOrCreate will find and load an existing dependabotconfig
// or promise to create one if missing and a config is required
func LoadOrCreate(path string) (*node, error) {

	n, err := Load(path)
	if err != nil && errors.Is(err, ErrMissingConfigFile) {
		n.repo.dependabotFilePath = ".github/dependabot.yml"
		return n, nil
	}
	return n, err
}

// A repo encapsulates all of the file path information for
// a github repository.
func newRepo(path string) (*repo, error) {

	fullpath, err := ensureAbsolutePath(path)
	if err != nil {
		return nil, err
	}

	root, err := osGetRootFolder(fullpath)
	if err != nil {
		return nil, errors.Wrapf(err, "error resolving to a github repo - %s", fullpath)
	}
	ret := &repo{root: root}

	fi, err := osStat(fullpath)
	if err != nil {
		return ret, ErrMissingConfigFile
	}

	isFile := false
	isDirectory := false
	switch mode := fi.Mode(); {
	case mode.IsRegular():
		isFile = true
	case mode.IsDir():
		isDirectory = true
	default:
		return ret, errors.New("unsupported file path")
	}

	files := []string{
		"dependabot.yml", "dependabot.yaml", ".github/dependabot.yml", ".github/dependabot.yaml",
	}

	fileName := "unknown"

	if isDirectory {
		// look for dependabot files from root
		fileName = findDependabotFile(root, files)
		if fileName == "" {
			return ret, ErrMissingConfigFile
		}
	}
	if isFile {
		// is the file a whitelisted dependabot file
		fileName = isADependabotFile(fi.Name(), files)
		if fileName == "" {
			return ret, ErrMissingConfigFile
		}
	}

	ret.dependabotFileExists = true
	ret.dependabotFilePath = fileName

	return ret, nil
}

// returns a matched file if it exists or an empty string
func findDependabotFile(root string, files []string) string {
	for _, f := range files {
		if pathExists(filepath.Join(root, f)) {
			return f
		}
	}
	return ""
}

// checks if a file exists in the list of files and returns either
// the name of the match or an empty string.
func isADependabotFile(fileName string, files []string) string {
	for _, f := range files {
		if f == fileName {
			return f
		}
	}
	return ""
}

// pathExists return true if exists
func pathExists(path string) bool {
	if _, err := osStat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// ensureAbsolutePath makes the path absolute or returns an error
func ensureAbsolutePath(path string) (string, error) {
	if !filepath.IsAbs(path) {
		wd, err := osGetwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(wd, path), nil
	}
	return path, nil
}

// getRootFolder returns the path to the 'root' folder or an error
func getRootFolder(dir string) (string, error) {
	path, err := getCommandOutput(dir, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	bits := strings.Split(path, "\n")
	return bits[0], nil
}

// getCommandOutput evaluates the given command and returns the trimmed output
func getCommandOutput(dir string, name string, args ...string) (string, error) {
	e := exec.Command(name, args...)
	if dir != "" {
		e.Dir = dir
	}
	data, err := e.CombinedOutput()
	text := string(data)
	text = strings.TrimSpace(text)
	return text, err
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
