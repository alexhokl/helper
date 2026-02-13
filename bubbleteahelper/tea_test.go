package bubbleteahelper

import (
	"os"
	"path/filepath"
	"strings"
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
		_ = os.Chmod(readOnlyDir, 0755)
	})

	logPath := filepath.Join(readOnlyDir, "test.log")

	err := SetupLogFile(logPath, "test")
	if err == nil {
		t.Error("SetupLogFile() in read-only directory should return error")
	}
}

func TestSetupLogFileErrorMessage(t *testing.T) {
	// Test that error message contains useful information
	invalidPath := "/nonexistent/directory/path/test.log"

	err := SetupLogFile(invalidPath, "test")
	if err == nil {
		t.Fatal("SetupLogFile() with invalid path should return error")
	}

	if !strings.Contains(err.Error(), "failed to open log file") {
		t.Errorf("Error message should contain 'failed to open log file', got: %v", err)
	}
}

func TestSetupLogFileSpecialCharactersInPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Test with special characters in prefix
	err := SetupLogFile(logPath, "test-prefix_with.special:chars")
	if err != nil {
		t.Fatalf("SetupLogFile() with special characters in prefix error: %v", err)
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("SetupLogFile() did not create log file with special prefix")
	}
}

func TestSetupLogFileLongPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Test with a long prefix
	longPrefix := strings.Repeat("a", 1000)
	err := SetupLogFile(logPath, longPrefix)
	if err != nil {
		t.Fatalf("SetupLogFile() with long prefix error: %v", err)
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("SetupLogFile() did not create log file with long prefix")
	}
}

func TestSetupLogFileMultipleCalls(t *testing.T) {
	tmpDir := t.TempDir()

	// Test multiple calls to SetupLogFile with different files
	for i := 0; i < 5; i++ {
		logPath := filepath.Join(tmpDir, filepath.Base(t.Name())+string(rune('a'+i))+".log")
		err := SetupLogFile(logPath, "test")
		if err != nil {
			t.Fatalf("SetupLogFile() call %d error: %v", i, err)
		}

		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Errorf("SetupLogFile() call %d did not create log file", i)
		}
	}
}

func TestSetupLogFileEmptyPath(t *testing.T) {
	// Test with empty path
	err := SetupLogFile("", "test")
	if err == nil {
		t.Error("SetupLogFile() with empty path should return error")
	}
}

// Benchmark tests

func BenchmarkSetupLogFile(b *testing.B) {
	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logPath := filepath.Join(tmpDir, "bench.log")
		_ = SetupLogFile(logPath, "bench")
	}
}

func BenchmarkSetupLogFileWithLongPrefix(b *testing.B) {
	tmpDir := b.TempDir()
	longPrefix := strings.Repeat("prefix", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logPath := filepath.Join(tmpDir, "bench.log")
		_ = SetupLogFile(logPath, longPrefix)
	}
}
