package source

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"kcl-lang.io/kcl-go/pkg/service"
	"kcl-lang.io/kcl-go/pkg/spec/gpyrpc"
)

const (
	// DefaultEntryFile is the default entry file for the package folder.
	DefaultEntryFile = "main.k"
	// DefaultEntryFile is the default entry config for the package folder.
	DefaultEntryConfig = "kcl.yaml"
	// EntryFilePattern is the wildcard pattern for kcl files.
	EntryFilePattern = "*.k"
)

var (
	// ErrNoEntry denotes that no entry found for the given package.
	ErrNoEntry = errors.New("no kcl entry found, please check the code package")
)

// GetSourceFromDir returns the kcl source code located at a filepath,
func GetSourceFromDir(dir string) (string, error) {
	// 0. TODO: kcl.mod entries
	// 1. kcl.yaml
	path := filepath.Join(dir, DefaultEntryConfig)
	if FileExists(path) {
		client := service.NewKclvmServiceClient()
		resp, err := client.LoadSettingsFiles(&gpyrpc.LoadSettingsFiles_Args{
			WorkDir: dir,
			Files:   []string{path},
		})
		if err != nil {
			return "", err
		}
		return GetSourceFromEntryFiles(resp.GetKclCliConfigs().Files)
	}
	// 2. main.k
	path = filepath.Join(dir, DefaultEntryFile)
	if FileExists(path) {
		bytes, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
	// 3. All k files in the folder
	matches, err := filepath.Glob(filepath.Join(dir, EntryFilePattern))
	if err != nil || len(matches) < 1 {
		return "", ErrNoEntry
	}
	var entries []string
	for _, match := range matches {
		if !strings.HasPrefix(match, "_") && !strings.HasPrefix(match, "test_") {
			entries = append(entries, match)
		}
	}

	return GetSourceFromEntryFiles(entries)
}

// GetSourceFromEntryFiles returns the kcl entry located at a directory
func GetSourceFromEntryFiles(entries []string) (string, error) {
	if len(entries) == 0 {
		return "", ErrNoEntry
	}

	var bt bytes.Buffer

	for _, entry := range entries {
		bytes, err := os.ReadFile(entry)
		if err != nil {
			return "", err
		}
		_, err = bt.Write(bytes)
		if err != nil {
			return "", err
		}
	}

	return bt.String(), nil
}

// FileExists mark whether the path exists.
func FileExists(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil || fi.IsDir() {
		return false
	}
	return true
}
