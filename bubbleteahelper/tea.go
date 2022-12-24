package bubbleteahelper

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func SetupLogFile(logFilePath string, prefix string) error {
	if f, err := tea.LogToFile(logFilePath, prefix); err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	} else {
		defer func() {
			err = f.Close()
			if err != nil {
				log.Fatal("failed to close log file: %w", err)
			}
		}()
	}
	return nil
}
