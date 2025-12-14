package iohelper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadStringFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		wantContent string
		wantErr     bool
	}{
		{
			name:        "simple content",
			content:     "hello world",
			wantContent: "hello world",
			wantErr:     false,
		},
		{
			name:        "multiline content",
			content:     "line1\nline2\nline3",
			wantContent: "line1\nline2\nline3",
			wantErr:     false,
		},
		{
			name:        "empty file",
			content:     "",
			wantContent: "",
			wantErr:     false,
		},
		{
			name:        "content with special characters",
			content:     "hello\tworld\n\r\nspecial: @#$%^&*()",
			wantContent: "hello\tworld\n\r\nspecial: @#$%^&*()",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			path := filepath.Join(tmpDir, tt.name+".txt")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			got, err := ReadStringFromFile(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadStringFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantContent {
				t.Errorf("ReadStringFromFile() = %q, want %q", got, tt.wantContent)
			}
		})
	}
}

func TestReadStringFromFileErrors(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "empty path",
			path: "",
		},
		{
			name: "non-existent file",
			path: "/nonexistent/path/to/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ReadStringFromFile(tt.path)
			if err == nil {
				t.Errorf("ReadStringFromFile(%q) should return error", tt.path)
			}
		})
	}
}

func TestReadBytesFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with binary content
	binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	binaryPath := filepath.Join(tmpDir, "binary.bin")
	if err := os.WriteFile(binaryPath, binaryContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	got, err := ReadBytesFromFile(binaryPath)
	if err != nil {
		t.Fatalf("ReadBytesFromFile() error: %v", err)
	}

	if len(got) != len(binaryContent) {
		t.Errorf("ReadBytesFromFile() len = %d, want %d", len(got), len(binaryContent))
	}

	for i, b := range got {
		if b != binaryContent[i] {
			t.Errorf("ReadBytesFromFile() byte[%d] = %x, want %x", i, b, binaryContent[i])
		}
	}
}

func TestReadBytesFromFileEmptyPath(t *testing.T) {
	_, err := ReadBytesFromFile("")
	if err == nil {
		t.Error("ReadBytesFromFile(\"\") should return error")
	}
}

func TestReadBytesFromFileNotFound(t *testing.T) {
	_, err := ReadBytesFromFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ReadBytesFromFile() with non-existent file should return error")
	}
}

func TestReadFirstLineFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		wantLine string
	}{
		{
			name:     "single line",
			content:  "only one line",
			wantLine: "only one line",
		},
		{
			name:     "multiple lines",
			content:  "first line\nsecond line\nthird line",
			wantLine: "first line",
		},
		{
			name:     "empty file",
			content:  "",
			wantLine: "",
		},
		{
			name:     "line with trailing newline",
			content:  "first line\n",
			wantLine: "first line",
		},
		{
			name:     "windows line endings",
			content:  "first line\r\nsecond line",
			wantLine: "first line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name+".txt")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			got, err := ReadFirstLineFromFile(path)
			if err != nil {
				t.Fatalf("ReadFirstLineFromFile() error: %v", err)
			}
			if got != tt.wantLine {
				t.Errorf("ReadFirstLineFromFile() = %q, want %q", got, tt.wantLine)
			}
		})
	}
}

func TestReadFirstLineFromFileErrors(t *testing.T) {
	_, err := ReadFirstLineFromFile("")
	if err == nil {
		t.Error("ReadFirstLineFromFile(\"\") should return error")
	}

	_, err = ReadFirstLineFromFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ReadFirstLineFromFile() with non-existent file should return error")
	}
}

func TestReadFirstLineBytesFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with binary-like content in first line
	content := []byte("binary\x00data\nignored")
	path := filepath.Join(tmpDir, "binary_line.txt")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	got, err := ReadFirstLineBytesFromFile(path)
	if err != nil {
		t.Fatalf("ReadFirstLineBytesFromFile() error: %v", err)
	}

	expected := []byte("binary\x00data")
	if string(got) != string(expected) {
		t.Errorf("ReadFirstLineBytesFromFile() = %v, want %v", got, expected)
	}
}

func TestReadLinesFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		content   string
		wantLines []string
	}{
		{
			name:      "multiple lines",
			content:   "line1\nline2\nline3",
			wantLines: []string{"line1", "line2", "line3"},
		},
		{
			name:      "single line",
			content:   "only line",
			wantLines: []string{"only line"},
		},
		{
			name:      "empty file",
			content:   "",
			wantLines: nil,
		},
		{
			name:      "lines with trailing newline",
			content:   "line1\nline2\n",
			wantLines: []string{"line1", "line2"},
		},
		{
			name:      "empty lines in between",
			content:   "line1\n\nline3",
			wantLines: []string{"line1", "", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name+".txt")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			got, err := ReadLinesFromFile(path)
			if err != nil {
				t.Fatalf("ReadLinesFromFile() error: %v", err)
			}

			if len(got) != len(tt.wantLines) {
				t.Errorf("ReadLinesFromFile() len = %d, want %d", len(got), len(tt.wantLines))
				return
			}

			for i, line := range got {
				if line != tt.wantLines[i] {
					t.Errorf("ReadLinesFromFile() line[%d] = %q, want %q", i, line, tt.wantLines[i])
				}
			}
		})
	}
}

func TestReadLinesFromFileErrors(t *testing.T) {
	_, err := ReadLinesFromFile("")
	if err == nil {
		t.Error("ReadLinesFromFile(\"\") should return error")
	}

	_, err = ReadLinesFromFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ReadLinesFromFile() with non-existent file should return error")
	}
}

func TestIsFileExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	existingFile := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing file",
			path: existingFile,
			want: true,
		},
		{
			name: "non-existent file",
			path: filepath.Join(tmpDir, "nonexistent.txt"),
			want: false,
		},
		{
			name: "directory is not a file",
			path: tmpDir,
			want: false,
		},
		{
			name: "empty path",
			path: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFileExist(tt.path)
			if got != tt.want {
				t.Errorf("IsFileExist(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsDirectoryExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing directory",
			path: tmpDir,
			want: true,
		},
		{
			name: "existing subdirectory",
			path: subDir,
			want: true,
		},
		{
			name: "non-existent directory",
			path: filepath.Join(tmpDir, "nonexistent"),
			want: false,
		},
		{
			name: "file is not a directory",
			path: testFile,
			want: false,
		},
		{
			name: "empty path",
			path: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDirectoryExist(tt.path)
			if got != tt.want {
				t.Errorf("IsDirectoryExist(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestCreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create new directory",
			path:    filepath.Join(tmpDir, "newdir"),
			wantErr: false,
		},
		{
			name:    "create nested directories",
			path:    filepath.Join(tmpDir, "nested", "deep", "dir"),
			wantErr: false,
		},
		{
			name:    "existing directory",
			path:    tmpDir,
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateDirectory(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.path != "" {
				if !IsDirectoryExist(tt.path) {
					t.Errorf("CreateDirectory() directory was not created at %q", tt.path)
				}
			}
		})
	}
}

func TestCreateDirectoryIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "idempotent")

	// Create directory first time
	err := CreateDirectory(newDir)
	if err != nil {
		t.Fatalf("CreateDirectory() first call error: %v", err)
	}

	// Create directory second time (should succeed)
	err = CreateDirectory(newDir)
	if err != nil {
		t.Errorf("CreateDirectory() second call error: %v", err)
	}

	if !IsDirectoryExist(newDir) {
		t.Error("CreateDirectory() directory should exist after calls")
	}
}

func TestGenerateCRC32Checksum(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content []byte
	}{
		{
			name:    "simple content",
			content: []byte("hello world"),
		},
		{
			name:    "empty file",
			content: []byte{},
		},
		{
			name:    "binary content",
			content: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name+".bin")
			if err := os.WriteFile(path, tt.content, 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			checksum, err := GenerateCRC32Checksum(path)
			if err != nil {
				t.Fatalf("GenerateCRC32Checksum() error: %v", err)
			}

			// Calculate checksum again to verify consistency
			checksum2, err := GenerateCRC32Checksum(path)
			if err != nil {
				t.Fatalf("GenerateCRC32Checksum() second call error: %v", err)
			}

			if checksum != checksum2 {
				t.Errorf("GenerateCRC32Checksum() not consistent: %d != %d", checksum, checksum2)
			}
		})
	}
}

func TestGenerateCRC32ChecksumDifferentContent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two files with different content
	path1 := filepath.Join(tmpDir, "file1.txt")
	path2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(path1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}
	if err := os.WriteFile(path2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	checksum1, err := GenerateCRC32Checksum(path1)
	if err != nil {
		t.Fatalf("GenerateCRC32Checksum() file1 error: %v", err)
	}

	checksum2, err := GenerateCRC32Checksum(path2)
	if err != nil {
		t.Fatalf("GenerateCRC32Checksum() file2 error: %v", err)
	}

	if checksum1 == checksum2 {
		t.Error("GenerateCRC32Checksum() different content should produce different checksums")
	}
}

func TestGenerateCRC32ChecksumSameContent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two files with same content
	path1 := filepath.Join(tmpDir, "file1.txt")
	path2 := filepath.Join(tmpDir, "file2.txt")
	content := []byte("identical content")

	if err := os.WriteFile(path1, content, 0644); err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}
	if err := os.WriteFile(path2, content, 0644); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	checksum1, err := GenerateCRC32Checksum(path1)
	if err != nil {
		t.Fatalf("GenerateCRC32Checksum() file1 error: %v", err)
	}

	checksum2, err := GenerateCRC32Checksum(path2)
	if err != nil {
		t.Fatalf("GenerateCRC32Checksum() file2 error: %v", err)
	}

	if checksum1 != checksum2 {
		t.Errorf("GenerateCRC32Checksum() same content should produce same checksum: %d != %d", checksum1, checksum2)
	}
}

func TestGenerateCRC32ChecksumErrors(t *testing.T) {
	_, err := GenerateCRC32Checksum("")
	if err == nil {
		t.Error("GenerateCRC32Checksum(\"\") should return error")
	}

	_, err = GenerateCRC32Checksum("/nonexistent/file.txt")
	if err == nil {
		t.Error("GenerateCRC32Checksum() with non-existent file should return error")
	}
}

// Benchmark tests
func BenchmarkReadStringFromFile(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench.txt")
	content := []byte("benchmark content for testing file reading performance")
	if err := os.WriteFile(path, content, 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ReadStringFromFile(path)
	}
}

func BenchmarkReadLinesFromFile(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench_lines.txt")
	content := []byte("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	if err := os.WriteFile(path, content, 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ReadLinesFromFile(path)
	}
}

func BenchmarkIsFileExist(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench.txt")
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsFileExist(path)
	}
}

func BenchmarkGenerateCRC32Checksum(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench.bin")
	content := make([]byte, 1024) // 1KB file
	for i := range content {
		content[i] = byte(i % 256)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateCRC32Checksum(path)
	}
}
