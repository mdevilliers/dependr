package dependabot

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type (
	node struct {
		// N holds a reference to the parsed dependabot configuration
		N *yaml.Node
		// repo holds the sanitised file information for the dependabot config file
		repo repo
	}

	Update struct {
		PackageEcoSystem string
		Directory        string
		Schedule         struct {
			Interval string
		}
	}

	repo struct {
		// root is the absolute path to the folder containing filePath
		root string
		// dependabotFilePath is the local path (from the root) to the dependabot configuration
		dependabotFilePath string
	}
)

var (
	ErrMissingConfigFile = errors.New("error finding dependabot config")
)

// Load will search the path for a dependabot file, returning an
// initialsed parsed config or an error
func Load(path string) (*node, error) {

	repo, err := NewRepo(path)
	if err != nil {
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

func NewRepo(path string) (*repo, error) {

	fullpath, err := ensureAbsolutePath(path)
	if err != nil {
		return nil, err
	}

	fi, err := os.Lstat(fullpath)
	if err != nil {
		return nil, err
	}

	isFile := false
	isDirectory := false
	switch mode := fi.Mode(); {
	case mode.IsRegular():
		isFile = true
	case mode.IsDir():
		isDirectory = true
	default:
		return nil, errors.New("unsupported file path")
	}

	parent := fullpath
	if isFile {
		parent = filepath.Dir(fullpath)
	}

	root, err := getCommandOutput(parent, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}

	files := []string{
		"dependabot.yml", "dependabot.yaml", ".github/dependabot.yml", ".github/dependabot.yaml",
	}

	fileName := "unknown"

	if isDirectory {
		// look for dependabot files from root
		found := false
		for _, f := range files {
			if fileExists(filepath.Join(root, f)) {
				fileName = f
				found = true
				break
			}
		}
		if !found {
			return &repo{root: root}, ErrMissingConfigFile
		}
	}
	if isFile {
		// is the file a whitelisted dependabot file
		found := false
		for _, f := range files {
			if f == fi.Name() {
				fileName = f
				found = true
				break
			}
		}
		if !found {
			return &repo{root: root}, ErrMissingConfigFile
		}
	}

	return &repo{
		root:               root,
		dependabotFilePath: fileName,
	}, nil
}

// fileExists return true if exists
func fileExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// ensureAbsolutePath makes the path absolute or returns an error
func ensureAbsolutePath(path string) (string, error) {
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(wd, path), nil
	}
	return path, nil
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
