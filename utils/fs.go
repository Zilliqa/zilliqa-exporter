package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
)

func PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func PathIsFile(path string) bool {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !stat.IsDir()
}

func PathIsDir(path string) bool {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return stat.IsDir()
}

func EnsureDir(dir string) error {
	if PathExists(dir) && !PathIsDir(dir) {
		return errors.New(fmt.Sprintf("path %s exists but is not a directory", dir))
	}
	return os.MkdirAll(dir, 0755)
}
