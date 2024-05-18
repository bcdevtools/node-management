package utils

import (
	"io"
	"os"
)

func FileInfo(path string) (fileMode os.FileMode, exists, isDir bool, error error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		error = err
		return
	}

	fileMode = fi.Mode().Perm()
	exists = true
	isDir = fi.IsDir()
	return
}

func IsEmptyDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}

	return false, err // Either not empty or error, suits both cases
}
