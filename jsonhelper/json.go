package jsonhelper

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

	return j.NewEncoder(file).Encode(object)
}

// GetJSONString returns a JSON string of the specified object
func GetJSONString(object interface{}) (string, error) {
	buf, err := GetJSONStringBuffer(object)
	return buf.String(), err
}

// GetJSONStringBuffer returns a buffer of JSON string of the specified object
func GetJSONStringBuffer(object interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBufferString("")
	encoder := j.NewEncoder(buf)
	err := encoder.Encode(object)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// ParseJSONReader parses JSON string from the specified reader and fill the content
// into the output object
func ParseJSONReader(json io.Reader, output interface{}) error {
	err := j.NewDecoder(json).Decode(&output)
	if err != nil {
		return err
	}
	return nil
}

// ParseJSONString parses JSON string from the specified reader and fill the content
// into the output object
func ParseJSONString(jsonString string, output interface{}) error {
	buf := bytes.NewBufferString(jsonString)
	return ParseJSONReader(buf, output)
}
