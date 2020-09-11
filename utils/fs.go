package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func ReadFloat64(file string) (float64, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	return val, err
}

func FindFile(paths, names []string) string {
	for _, path := range paths {
		for _, name := range names {
			p := filepath.Join(path, name)
			if PathExists(p) {
				return p
			}
		}
	}
	return ""
}
