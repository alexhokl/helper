package cli

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// GetOpenCommand returns the command and its required arguments according to
// the current operation system
func GetOpenCommand(args ...string) (string, []string) {
	switch runtime.GOOS {
	case "windows":
		cmdArgs := []string{"/C", "start"}
		cmdArgs = append(cmdArgs, args...)
		return "cmd", cmdArgs
	case "darwin":
		return "open", args
	default:
		return "xdg-open", args
	}
}

func OpenInBrowser(url string) error {
	cmdName, cmdArgs := GetOpenCommand(url)
	_, errOpen := exec.Command(cmdName, cmdArgs...).Output()
	if errOpen != nil {
		cmdParts := []string { cmdName }
		cmdParts = append(cmdParts, cmdArgs...)
		return fmt.Errorf("unable to complete command [%s] %w", strings.Join(cmdParts, " "), errOpen)
	}
	return nil
}
