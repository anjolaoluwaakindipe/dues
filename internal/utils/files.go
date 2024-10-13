package utils

import (
	"io/fs"
	"os"
	"path/filepath"
)

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)

	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func WalkSubdirectories(path string, callback func(string)) {
	filepath.WalkDir(path, func(sub_path string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			callback(sub_path)
		}
		return nil
	})
}
