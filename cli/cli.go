package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/viper"
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
		cmdParts := []string{cmdName}
		cmdParts = append(cmdParts, cmdArgs...)
		return fmt.Errorf("unable to complete command [%s] %w", strings.Join(cmdParts, " "), errOpen)
	}
	return nil
}

func ConfigureViper(configFilePath string, applicationName string, verbose bool, environmentVariablePrefix string) {
	if environmentVariablePrefix == "" {
		environmentVariablePrefix = strings.ReplaceAll(applicationName, "-", "_")
	}

	if configFilePath != "" {
		viper.SetConfigFile(configFilePath)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(fmt.Sprintf(".%s", applicationName))
	}

	viper.SetEnvPrefix(environmentVariablePrefix)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}
