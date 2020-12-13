package json

import (
	"bytes"
	j "encoding/json"
	"fmt"
	"io"
	"os"
)

// WriteToJSONFile writes the specified object as a JSON file to the specified path
func WriteToJSONFile(path string, object interface{}, isOverwrite bool) error {
	if path == "" {
		return fmt.Errorf("path is not specified")
	}
	if object == nil {
		return fmt.Errorf("object cannot be empty")
	}

	if _, err := os.Stat(path); os.IsExist(err) {
		if !isOverwrite {
			return fmt.Errorf("File [%s] is already exist", path)
		}
		errRemove := os.Remove(path)
		if errRemove != nil {
			return errRemove
		}
	}

	file, errOpen := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if errOpen != nil {
		return errOpen
	}
	defer file.Close()

	encoder := j.NewEncoder(file)
	err := encoder.Encode(object)
	return err
}

// GetJSONString returns a JSON string of the specified object
func GetJSONString(object interface{}) (string, error) {
	buf := bytes.NewBufferString("")
	encoder := j.NewEncoder(buf)
	err := encoder.Encode(object)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ParseJSON parses JSON string from the specified reader and fill the content
// into the output object
func ParseJSON(json io.Reader, output interface{}) error {
	err := j.NewDecoder(json).Decode(&output)
	if err != nil {
		return err
	}
	return nil
}
