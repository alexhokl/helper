package bubbleteahelper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupLogFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	err := SetupLogFile(logPath, "test")
	if err != nil {
		t.Fatalf("SetupLogFile() error: %v", err)
	}

	// Verify the log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("SetupLogFile() did not create log file")
	}
}

func TestSetupLogFileEmptyPrefix(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	err := SetupLogFile(logPath, "")
	if err != nil {
		t.Fatalf("SetupLogFile() with empty prefix error: %v", err)
	}

	// Verify the log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("SetupLogFile() did not create log file")
	}
}

func TestSetupLogFileInvalidPath(t *testing.T) {
	// Test with an invalid path (directory that doesn't exist)
	invalidPath := "/nonexistent/directory/path/test.log"

	err := SetupLogFile(invalidPath, "test")
	if err == nil {
		t.Error("SetupLogFile() with invalid path should return error")
	}
}

func TestSetupLogFileNestedDirectory(t *testing.T) {
	// Create temp directory with nested structure
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "dir")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	logPath := filepath.Join(nestedDir, "test.log")

	err := SetupLogFile(logPath, "nested-test")
	if err != nil {
		t.Fatalf("SetupLogFile() with nested directory error: %v", err)
	}

	// Verify the log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("SetupLogFile() did not create log file in nested directory")
	}
}

func TestSetupLogFileOverwrite(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create an existing file
	existingContent := []byte("existing content")
	if err := os.WriteFile(logPath, existingContent, 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Call SetupLogFile which should overwrite/append to the file
	err := SetupLogFile(logPath, "test")
	if err != nil {
		t.Fatalf("SetupLogFile() error: %v", err)
	}

	// Verify the log file still exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("SetupLogFile() removed existing log file")
	}
}

func TestSetupLogFilePermissions(t *testing.T) {
	// Skip on non-Unix systems where permission handling differs
	if os.Getenv("CI") != "" {
		t.Skip("Skipping permission test in CI environment")
	}

	// Create temp directory
	tmpDir := t.TempDir()

	// Create a read-only directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}

	// Ensure we can clean up
	t.Cleanup(func() {
		os.Chmod(readOnlyDir, 0755)
	})

	logPath := filepath.Join(readOnlyDir, "test.log")

	err := SetupLogFile(logPath, "test")
	if err == nil {
		t.Error("SetupLogFile() in read-only directory should return error")
	}
}
