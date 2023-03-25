package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/structs"
	"github.com/spf13/cobra"
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

/// BindFlagsAndEnvToViper binds the flags and environment variables to viper
///
/// Example:
/// type config struct {
///   Param1 string `mapstructure:"param1" structs:"param1" env:"PARAM1"`
/// }
///
/// func init() {
///    rootCmd.AddCommand(listCmd)
///    flags := listCmd.Flags()
///    flags.StringVar(&listOpts.format, "format", "Format")
///    cli.BindFlagsAndEnvToViper(listCmd, listOpts)
/// }
func BindFlagsAndEnvToViper(cmd *cobra.Command, params interface{}) (err error) {
	for _, field := range structs.Fields(params) {
        key := field.Tag("structs")
        env := field.Tag("env")
        err = viper.BindPFlag(key, cmd.Flags().Lookup(key))
        if err != nil {
            return err
        }
        err = viper.BindEnv(key, env)
        if err != nil {
            return err
        }
    }
    return nil
}
