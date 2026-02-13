package cli

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

// Tests for OpenInBrowser

func TestOpenInBrowser_InvalidURL(t *testing.T) {
	// Test with an invalid URL scheme that should fail
	// On most systems, trying to open a non-existent file path should fail
	err := OpenInBrowser("")
	// The behavior depends on the OS - some may return error, some may not
	// We mainly want to ensure the function doesn't panic
	_ = err
}

func TestOpenInBrowser_ErrorMessageFormat(t *testing.T) {
	// Test with a path that is very likely to fail on any OS
	// Use a URL that the system command will reject
	err := OpenInBrowser("file:///nonexistent/path/that/should/not/exist/ever/12345.xyz")

	// If there's an error, verify the error message format
	if err != nil {
		errStr := err.Error()
		if !strings.Contains(errStr, "unable to complete command") {
			t.Errorf("OpenInBrowser() error should contain 'unable to complete command', got: %v", errStr)
		}
	}
	// Note: On some systems (like macOS), open command may succeed even for non-existent files
}

func TestOpenInBrowser_ValidURL(t *testing.T) {
	// Skip this test in CI or headless environments
	if os.Getenv("CI") != "" {
		t.Skip("Skipping browser test in CI environment")
	}

	// We can't really test opening a browser without side effects,
	// but we can verify the function doesn't panic with a valid URL
	// This test is more for documentation purposes
	t.Skip("Skipping actual browser opening test")
}

// Additional GetOpenCommand tests for documentation

func TestGetOpenCommand_EmptyArgs(t *testing.T) {
	cmd, args := GetOpenCommand()

	if cmd == "" {
		t.Error("GetOpenCommand() should return non-empty command")
	}

	switch runtime.GOOS {
	case "windows":
		// Windows adds /C and start even with no user args
		if len(args) != 2 {
			t.Errorf("GetOpenCommand() on windows with no args: len(args) = %d, want 2", len(args))
		}
	case "darwin", "linux":
		if len(args) != 0 {
			t.Errorf("GetOpenCommand() on %s with no args: len(args) = %d, want 0", runtime.GOOS, len(args))
		}
	}
}

func TestGetOpenCommand_SpecialCharactersInURL(t *testing.T) {
	specialURL := "https://example.com/path?query=value&foo=bar#anchor"
	cmd, args := GetOpenCommand(specialURL)

	if cmd == "" {
		t.Error("GetOpenCommand() should return non-empty command")
	}

	// Verify the URL is included in args
	found := false
	for _, arg := range args {
		if arg == specialURL {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetOpenCommand() args should contain the URL, got: %v", args)
	}
}

func TestGetOpenCommand_FileURL(t *testing.T) {
	fileURL := "file:///home/user/document.pdf"
	cmd, args := GetOpenCommand(fileURL)

	if cmd == "" {
		t.Error("GetOpenCommand() should return non-empty command")
	}

	// Verify the file URL is included in args
	found := false
	for _, arg := range args {
		if arg == fileURL {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetOpenCommand() args should contain the file URL, got: %v", args)
	}
}

// Additional ConfigureViper tests

func TestConfigureViperWithVerboseAndExistingConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")
	configContent := []byte("verbose_key: verbose_value\n")
	if err := os.WriteFile(configFile, configContent, 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	viper.Reset()
	defer viper.Reset()

	// Test with verbose=true to cover the verbose output path
	ConfigureViper(configFile, "test-app", true, "TEST")

	// Verify the config file was read
	if viper.ConfigFileUsed() != configFile {
		t.Errorf("ConfigureViper() config file = %q, want %q", viper.ConfigFileUsed(), configFile)
	}
}

func TestConfigureViperWithNonExistentConfig(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	// Test with a config file that doesn't exist
	// This should not panic, just not find a config
	ConfigureViper("/nonexistent/path/config.yaml", "test-app", false, "TEST")

	// Viper should have been configured but no config loaded
	// This is a valid use case
}

func TestConfigureViperEnvPrefixWithHyphens(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	// Test that app name with hyphens gets converted to underscores
	// when environmentVariablePrefix is empty
	ConfigureViper("", "my-cool-app", false, "")

	// Set an environment variable with the expected prefix
	t.Setenv("MY_COOL_APP_TEST_VAR", "test_value")

	// Note: We can't easily verify the prefix was set correctly without
	// accessing viper internals, but we verify it doesn't panic
}

func TestConfigureViperWithHomeDir(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	// Test the path where configFilePath is empty (uses home directory)
	// This exercises the os.UserHomeDir() path
	ConfigureViper("", "test-app-home", false, "TEST_HOME")

	// Should not panic and should configure viper with home directory path
}

// Additional BindFlagsAndEnvToViper tests

type nestedConfig struct {
	Name    string `mapstructure:"name" structs:"name" env:"NAME"`
	Count   int    `mapstructure:"count" structs:"count" env:"COUNT"`
	Enabled bool   `mapstructure:"enabled" structs:"enabled" env:"ENABLED"`
}

func TestBindFlagsAndEnvToViperWithMultipleTypes(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	cmd := &cobra.Command{Use: "test"}

	opts := nestedConfig{}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name parameter")
	cmd.Flags().IntVar(&opts.Count, "count", 0, "Count parameter")
	cmd.Flags().BoolVar(&opts.Enabled, "enabled", false, "Enabled parameter")

	err := BindFlagsAndEnvToViper(cmd, opts)
	if err != nil {
		t.Fatalf("BindFlagsAndEnvToViper() returned error: %v", err)
	}

	// Test environment variable binding for different types
	t.Setenv("NAME", "test-name")
	t.Setenv("COUNT", "42")
	t.Setenv("ENABLED", "true")

	if viper.GetString("name") != "test-name" {
		t.Errorf("name = %q, want %q", viper.GetString("name"), "test-name")
	}
	if viper.GetInt("count") != 42 {
		t.Errorf("count = %d, want %d", viper.GetInt("count"), 42)
	}
	if viper.GetBool("enabled") != true {
		t.Errorf("enabled = %v, want %v", viper.GetBool("enabled"), true)
	}
}

func TestBindFlagsAndEnvToViperFlagOverridesEnv(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	cmd := &cobra.Command{Use: "test"}

	opts := testConfig{}
	cmd.Flags().StringVar(&opts.Param1, "param1", "", "Parameter 1")
	cmd.Flags().IntVar(&opts.Param2, "param2", 0, "Parameter 2")

	err := BindFlagsAndEnvToViper(cmd, opts)
	if err != nil {
		t.Fatalf("BindFlagsAndEnvToViper() returned error: %v", err)
	}

	// Set environment variable
	t.Setenv("PARAM1", "env_value")

	// Set flag value (should override env)
	if err := cmd.Flags().Set("param1", "flag_value"); err != nil {
		t.Fatalf("Failed to set flag: %v", err)
	}

	// Flag should take precedence
	if viper.GetString("param1") != "flag_value" {
		t.Errorf("param1 = %q, want %q (flag should override env)", viper.GetString("param1"), "flag_value")
	}
}

// Test LogUnableToMarkFlagAsRequired with different error types

func TestLogUnableToMarkFlagAsRequired_DifferentErrors(t *testing.T) {
	testCases := []struct {
		name     string
		flagName string
		err      error
	}{
		{
			name:     "simple error",
			flagName: "test-flag",
			err:      errors.New("simple error"),
		},
		{
			name:     "wrapped error",
			flagName: "another-flag",
			err:      errors.New("wrapped: inner error"),
		},
		{
			name:     "empty flag name",
			flagName: "",
			err:      errors.New("some error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("LogUnableToMarkFlagAsRequired() panicked: %v", r)
				}
			}()

			LogUnableToMarkFlagAsRequired(tc.flagName, tc.err)
		})
	}
}
