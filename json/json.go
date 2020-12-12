package json

import (
	"encoding/json"
	"fmt"
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

	encoder := json.NewEncoder(file)
	err := encoder.Encode(object)
	return err
}
