package dependabot

import (
	"io/fs"
	"os"
	"testing"

	"github.com/mdevilliers/depender/pkg/dependabot/dependabotfakes"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate os.FileInfo

func Test_Load_Dependabot_File_Exists(t *testing.T) {

	osGetRootFolder = func(string) (string, error) {
		return "/user/foo", nil
	}
	osGetwd = func() (string, error) {
		return "/user/", nil
	}
	osStat = func(name string) (os.FileInfo, error) {
		f := &dependabotfakes.FakeFileInfo{}
		f.ModeReturns(fs.ModeAppend) // file
		f.NameReturns("dependabot.yml")

		return f, nil
	}

	n, err := Load("./foo/dependabot.yml")

	require.Nil(t, err)
	require.True(t, n.repo.dependabotFileExists)
	require.Equal(t, "dependabot.yml", n.repo.dependabotFilePath)
	require.Equal(t, "/user/foo", n.repo.root)
}

func Test_Load_Dependabot_File_Not_Exists(t *testing.T) {

	osGetRootFolder = func(string) (string, error) {
		return "/user/foo", nil
	}
	osGetwd = func() (string, error) {
		return "/user/", nil
	}
	osStat = func(name string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	_, err := Load("./foo/dependabot.yml")

	require.NotNil(t, err)
}

func Test_LoadOrCreate_Dependabot_File_Exists(t *testing.T) {

	osGetRootFolder = func(string) (string, error) {
		return "/user/foo", nil
	}
	osGetwd = func() (string, error) {
		return "/user/", nil
	}
	osStat = func(name string) (os.FileInfo, error) {
		f := &dependabotfakes.FakeFileInfo{}
		f.ModeReturns(fs.ModeAppend) // file
		f.NameReturns("dependabot.yml")

		return f, nil
	}

	n, err := LoadOrCreate("./foo/dependabot.yml")

	require.Nil(t, err)
	require.True(t, n.repo.dependabotFileExists)
	require.Equal(t, "dependabot.yml", n.repo.dependabotFilePath)
	require.Equal(t, "/user/foo", n.repo.root)
}

func Test_LoadOrCreate_Dependabot_Folder_Not_Exists(t *testing.T) {

	osGetRootFolder = func(string) (string, error) {
		return "/user/foo", nil
	}
	osGetwd = func() (string, error) {
		return "/user/", nil
	}
	osStat = func(name string) (os.FileInfo, error) {
		if name == "/user/foo" {
			f := &dependabotfakes.FakeFileInfo{}
			f.ModeReturns(fs.ModeDir)
			f.NameReturns("foo")
			return f, nil
		}
		return nil, os.ErrNotExist
	}

	n, err := LoadOrCreate("./foo/")

	require.Nil(t, err)
	require.False(t, n.repo.dependabotFileExists)
	require.Equal(t, ".github/dependabot.yml", n.repo.dependabotFilePath)
	require.Equal(t, "/user/foo", n.repo.root)
}

func Test_Load_Not_A_GitRepo(t *testing.T) {
	osGetRootFolder = func(string) (string, error) {
		return "", errors.New("booyah!")
	}
	_, err := Load("./foo/dependabot.yml")

	require.NotNil(t, err)
}

func Test_Load_Dependabot_Folder_Exists(t *testing.T) {

	osGetRootFolder = func(string) (string, error) {
		return "/user/foo", nil
	}
	osGetwd = func() (string, error) {
		return "/user/", nil
	}

	osStat = func(name string) (os.FileInfo, error) {
		if name == "/user/foo" {
			f := &dependabotfakes.FakeFileInfo{}
			f.ModeReturns(fs.ModeDir)
			f.NameReturns("foo")

			return f, nil
		}
		if name == "/user/foo/.github/dependabot.yml" {
			f := &dependabotfakes.FakeFileInfo{}
			f.ModeReturns(fs.ModeAppend)
			f.NameReturns(".github/dependabot.yml")
			return f, nil
		}
		return nil, os.ErrNotExist
	}

	n, err := Load("./foo/")

	require.Nil(t, err)
	require.True(t, n.repo.dependabotFileExists)
	require.Equal(t, ".github/dependabot.yml", n.repo.dependabotFilePath)
	require.Equal(t, "/user/foo", n.repo.root)
}

func Test_Load_Dependabot_Folder_Not_Exists(t *testing.T) {

	osGetRootFolder = func(string) (string, error) {
		return "/user/foo", nil
	}
	osGetwd = func() (string, error) {
		return "/user/", nil
	}
	osStat = func(name string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	_, err := Load("./foo/")

	require.NotNil(t, err)
}
