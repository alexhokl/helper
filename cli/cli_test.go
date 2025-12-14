package cli

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestGetOpenCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCmd     string
		wantArgsLen int
		wantArgsHas []string
	}{
		{
			name:        "no args",
			args:        []string{},
			wantArgsLen: 0,
		},
		{
			name:        "single url",
			args:        []string{"https://example.com"},
			wantArgsHas: []string{"https://example.com"},
		},
		{
			name:        "multiple args",
			args:        []string{"https://example.com", "--new-window"},
			wantArgsHas: []string{"https://example.com", "--new-window"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := GetOpenCommand(tt.args...)

			// Verify command based on OS
			switch runtime.GOOS {
			case "windows":
				if cmd != "cmd" {
					t.Errorf("GetOpenCommand() on windows: cmd = %q, want %q", cmd, "cmd")
				}
				// Windows should have /C and start prepended
				if len(args) < 2 {
					t.Errorf("GetOpenCommand() on windows: args too short, got %v", args)
				} else {
					if args[0] != "/C" {
						t.Errorf("GetOpenCommand() on windows: args[0] = %q, want %q", args[0], "/C")
					}
					if args[1] != "start" {
						t.Errorf("GetOpenCommand() on windows: args[1] = %q, want %q", args[1], "start")
					}
				}
			case "darwin":
				if cmd != "open" {
					t.Errorf("GetOpenCommand() on darwin: cmd = %q, want %q", cmd, "open")
				}
				// Args should be passed through unchanged
				if len(args) != len(tt.args) {
					t.Errorf("GetOpenCommand() on darwin: len(args) = %d, want %d", len(args), len(tt.args))
				}
			default:
				if cmd != "xdg-open" {
					t.Errorf("GetOpenCommand() on linux: cmd = %q, want %q", cmd, "xdg-open")
				}
				// Args should be passed through unchanged
				if len(args) != len(tt.args) {
					t.Errorf("GetOpenCommand() on linux: len(args) = %d, want %d", len(args), len(tt.args))
				}
			}

			// Verify that expected args are present
			for _, wantArg := range tt.wantArgsHas {
				found := false
				for _, arg := range args {
					if arg == wantArg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetOpenCommand() args = %v, should contain %q", args, wantArg)
				}
			}
		})
	}
}

func TestGetOpenCommandReturnsCorrectCommandForCurrentOS(t *testing.T) {
	cmd, _ := GetOpenCommand("test")

	switch runtime.GOOS {
	case "windows":
		if cmd != "cmd" {
			t.Errorf("Expected 'cmd' on Windows, got %q", cmd)
		}
	case "darwin":
		if cmd != "open" {
			t.Errorf("Expected 'open' on Darwin, got %q", cmd)
		}
	default:
		if cmd != "xdg-open" {
			t.Errorf("Expected 'xdg-open' on Linux/other, got %q", cmd)
		}
	}
}

func TestConfigureViper(t *testing.T) {
	// Save original viper state and restore after test
	originalViper := viper.GetViper()
	defer func() {
		viper.Reset()
		*viper.GetViper() = *originalViper
	}()

	tests := []struct {
		name                      string
		configFilePath            string
		applicationName           string
		verbose                   bool
		environmentVariablePrefix string
		setupFunc                 func() string // returns temp dir path
		cleanupFunc               func(string)
	}{
		{
			name:                      "with explicit config file path",
			configFilePath:            "/tmp/test-config.yaml",
			applicationName:           "test-app",
			verbose:                   false,
			environmentVariablePrefix: "TEST_APP",
		},
		{
			name:                      "without config file path uses home dir",
			configFilePath:            "",
			applicationName:           "test-app",
			verbose:                   false,
			environmentVariablePrefix: "TEST_APP",
		},
		{
			name:                      "empty env prefix uses app name",
			configFilePath:            "",
			applicationName:           "my-app",
			verbose:                   false,
			environmentVariablePrefix: "",
		},
		{
			name:                      "app name with hyphens converts to underscores for env prefix",
			configFilePath:            "",
			applicationName:           "my-cool-app",
			verbose:                   true,
			environmentVariablePrefix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			// ConfigureViper should not panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("ConfigureViper() panicked: %v", r)
					}
				}()
				ConfigureViper(tt.configFilePath, tt.applicationName, tt.verbose, tt.environmentVariablePrefix)
			}()
		})
	}
}

func TestConfigureViperWithExistingConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")
	configContent := []byte("test_key: test_value\n")
	if err := os.WriteFile(configFile, configContent, 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	viper.Reset()
	defer viper.Reset()

	ConfigureViper(configFile, "test-app", false, "TEST")

	// Verify the config file was read
	if viper.ConfigFileUsed() != configFile {
		t.Errorf("ConfigureViper() config file = %q, want %q", viper.ConfigFileUsed(), configFile)
	}

	// Verify the value was read
	if viper.GetString("test_key") != "test_value" {
		t.Errorf("ConfigureViper() test_key = %q, want %q", viper.GetString("test_key"), "test_value")
	}
}

// testConfig is a sample config struct for testing BindFlagsAndEnvToViper
type testConfig struct {
	Param1 string `mapstructure:"param1" structs:"param1" env:"PARAM1"`
	Param2 int    `mapstructure:"param2" structs:"param2" env:"PARAM2"`
}

func TestBindFlagsAndEnvToViper(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	// Create a test command with flags
	cmd := &cobra.Command{
		Use: "test",
	}

	opts := testConfig{}
	cmd.Flags().StringVar(&opts.Param1, "param1", "", "Parameter 1")
	cmd.Flags().IntVar(&opts.Param2, "param2", 0, "Parameter 2")

	// Bind flags and env to viper
	err := BindFlagsAndEnvToViper(cmd, opts)
	if err != nil {
		t.Fatalf("BindFlagsAndEnvToViper() returned error: %v", err)
	}

	// Set environment variable and verify it's bound
	t.Setenv("PARAM1", "env_value")

	// Viper should pick up the environment variable
	if viper.GetString("param1") != "env_value" {
		t.Errorf("BindFlagsAndEnvToViper() param1 from env = %q, want %q", viper.GetString("param1"), "env_value")
	}
}

func TestBindFlagsAndEnvToViperWithEmptyStruct(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	cmd := &cobra.Command{
		Use: "test",
	}

	type emptyConfig struct{}
	opts := emptyConfig{}

	err := BindFlagsAndEnvToViper(cmd, opts)
	if err != nil {
		t.Errorf("BindFlagsAndEnvToViper() with empty struct returned error: %v", err)
	}
}

func TestBindFlagsAndEnvToViperWithMissingFlag(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	cmd := &cobra.Command{
		Use: "test",
	}

	// Config has param1, but we don't add the flag to the command
	type configWithMissingFlag struct {
		Param1 string `mapstructure:"param1" structs:"param1" env:"PARAM1"`
	}
	opts := configWithMissingFlag{}

	// This should return an error when flag doesn't exist
	// (viper.BindPFlag returns error for nil flags)
	err := BindFlagsAndEnvToViper(cmd, opts)
	if err == nil {
		t.Error("BindFlagsAndEnvToViper() with missing flag should return error")
	}
}

func TestLogUnableToMarkFlagAsRequired(t *testing.T) {
	// This function just logs, so we verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogUnableToMarkFlagAsRequired() panicked: %v", r)
		}
	}()

	testErr := errors.New("test error")
	LogUnableToMarkFlagAsRequired("test-flag", testErr)
}

// Benchmark tests
func BenchmarkGetOpenCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetOpenCommand("https://example.com")
	}
}

func BenchmarkGetOpenCommandMultipleArgs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetOpenCommand("https://example.com", "--new-window", "--private")
	}
}
