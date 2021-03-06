package iohelper

import (
	"fmt"
	"io/ioutil"
	"os"
)

// ReadStringFromFile returns content of the file in the specified path as string
func ReadStringFromFile(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	if !IsFileExist(path) {
		return "", fmt.Errorf("File [%s] does not exist", path)
	}
	file, errFile := ioutil.ReadFile(path)
	if errFile != nil {
		return "", errFile
	}
	return string(file), nil
}

// IsFileExist return true if a file exist in the specified path
func IsFileExist(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.Mode().IsRegular()
}

// IsDirectoryExist return true if a file exist in the specified path
func IsDirectoryExist(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.Mode().IsDir()
}
