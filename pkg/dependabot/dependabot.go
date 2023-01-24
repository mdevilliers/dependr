package dependabot

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type (
	node struct {
		// repo holds the sanitised file information for the dependabot config file
		repo repo
	}

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
	osMkdirAll  = os.MkdirAll

	// allow redirecting the git command for testing
	osGetRootFolder = getRootFolder
)

// Load will search the path for a dependabot file, returning an
// initialsed parsed config or an error
func Load(path string) (*node, error) {

	repo, err := newRepo(path, false)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing supplied path %s", path)
	}
	return &node{
		repo: *repo,
	}, nil
}

// LoadOrCreate will find and load an existing dependabotconfig
// or promise to create one if missing and a config is required.
// The path must path to a folder that exists.
func LoadOrCreate(path string) (*node, error) {

	repo, err := newRepo(path, true)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing supplied path %s", path)
	}
	return &node{
		repo: *repo,
	}, nil
}

// A repo encapsulates all of the file path information for
// a github repository.
func newRepo(path string, createIfMissing bool) (*repo, error) {

	fullpath, err := ensureAbsolutePath(path)
	if err != nil {
		return nil, err
	}
	ret := &repo{}

	fi, err := osStat(fullpath)
	if err != nil {
		return nil, err
	}

	var fileName string

	files := []string{
		"dependabot.yml", "dependabot.yaml", ".github/dependabot.yml", ".github/dependabot.yaml",
	}

	switch mode := fi.Mode(); {
	case mode.IsRegular():
		ret.root, err = osGetRootFolder(filepath.Dir(fullpath))
		if err != nil {
			return nil, errors.Wrapf(err, "error resolving to a github repo - %s", fullpath)
		}
		// is the file a whitelisted dependabot file
		fileName = isADependabotFile(fi.Name(), files)
		if fileName == "" {
			return ret, ErrMissingConfigFile
		}

	case mode.IsDir():
		ret.root, err = osGetRootFolder(fullpath)
		if err != nil {
			return nil, errors.Wrapf(err, "error resolving to a github repo - %s", fullpath)
		}
		// look for dependabot files from root
		fileName = findDependabotFile(ret.root, files)
		if fileName == "" {
			if createIfMissing {
				ret.dependabotFileExists = false
				ret.dependabotFilePath = ".github/dependabot.yml"
				return ret, nil
			}
			return ret, ErrMissingConfigFile
		}

	default:
		return nil, errors.New("unsupported file path")
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
