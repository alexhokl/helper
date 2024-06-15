package iohelper

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"os"
)

// ReadStringFromFile returns content of the file in the specified path as string
func ReadStringFromFile(path string) (string, error) {
	bytes, err := ReadBytesFromFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ReadBytesFromFile returns content of the file in the specified path as bytes
func ReadBytesFromFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if !IsFileExist(path) {
		return nil, fmt.Errorf("file [%s] does not exist", path)
	}
	file, errFile := os.ReadFile(path)
	if errFile != nil {
		return nil, errFile
	}
	return file, nil
}

func ReadFirstLineFromFile(path string) (string, error) {
	bytes, err := ReadFirstLineBytesFromFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ReadFirstLineBytesFromFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if !IsFileExist(path) {
		return nil, fmt.Errorf("file [%s] does not exist", path)
	}
	file, errFile := os.Open(path)
	if errFile != nil {
		return nil, errFile
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Bytes(), nil
	}
	return nil, nil
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

// CreateDirectory creates and ensure a directory in the specified path
func CreateDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if IsDirectoryExist(path) {
		return nil
	}
	return os.MkdirAll(path, 0755)
}

// GenerateCRC32Checksum generates CRC32 checksum of the file in the specified path
func GenerateCRC32Checksum(path string) (uint32, error) {
	fileBytes, err := ReadBytesFromFile(path)
	if err != nil {
		return 0, err
	}
	checksum := crc32.Checksum(fileBytes, crc32.MakeTable(crc32.Castagnoli))
	return checksum, nil
}
