package environment

import (
	"os"
	"path/filepath"
)

func detectPaths() (map[string]string, error) {
	paths := make(map[string]string, 2)
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	paths["home"] = filepath.Clean(home)

	wd, err := os.Getwd()
	if err != nil {
		paths["workspace"] = ""
	} else {
		paths["workspace"] = filepath.Clean(wd)
	}
	return paths, nil
}
