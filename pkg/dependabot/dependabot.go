package dependabot

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type (
	node struct {
		N        yaml.Node
		root     string
		filePath string
	}

	Update struct {
		PackageEcoSystem string
		Directory        string
		Schedule         struct {
			Interval string
		}
	}
)

var (
	ErrMissingConfigFile = errors.New("error finding dependabot config")
)

// Load will search the path for a dependabot file, returning an
// initialsed parsed config or an error
func Load(path string) (*node, error) {

	fi, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	switch mode := fi.Mode(); {
	case mode.IsRegular():
		return loadFile(path)
	case mode.IsDir():

		files := []string{
			"dependabot.yml",
			"dependabot.yaml",
			".github/dependabot.yml",
			".github/dependabot.yaml",
		}

		for _, file := range files {
			p := filepath.Join(path, file)
			n, err := loadFile(p)
			if err == nil { // check for no error
				return n, nil
			}
		}
		return nil, ErrMissingConfigFile
	default:
		return nil, ErrMissingConfigFile
	}
}

func loadFile(path string) (*node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrMissingConfigFile
		}
		return nil, err
	}

	var n yaml.Node
	if err := yaml.Unmarshal(data, &n); err != nil {
		return nil, err
	}

	// Assume we are in a git repository
	// Running :
	// git rev-parse --show-toplevel
	// Will return the root folder
	parent := filepath.Dir(path)
	response, err := getCommandOutput(parent, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	node := &node{
		N:        n,
		root:     response,
		filePath: strings.ReplaceAll(path, response+"/", ""),
	}

	return node, nil
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

func (n *node) EnsureUpdateExists(update Update) {
}
