package cli

import "runtime"

// GetOpenCommand returns the command and its required arguments according to
// the current operation system
func GetOpenCommand(args ...string) (string, []string) {
	switch runtime.GOOS {
	case "windows":
		cmdArgs := []string{"/C", "start"}
		for _, a := range args {
			cmdArgs = append(cmdArgs, a)
		}
		return "cmd", cmdArgs
	case "darwin":
		return "open", args
	default:
		return "xdg-open", args
	}
}
