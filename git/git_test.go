package git

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// isInGitRepo checks if we're running tests in a git repository
func isInGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

// skipIfNotGitRepo skips the test if not running in a git repository
func skipIfNotGitRepo(t *testing.T) {
	t.Helper()
	if !isInGitRepo() {
		t.Skip("skipping test: not in a git repository")
	}
}

func TestGetCurrentBranchName(t *testing.T) {
	skipIfNotGitRepo(t)

	branchName, err := GetCurrentBranchName()
	if err != nil {
		t.Fatalf("GetCurrentBranchName() returned error: %v", err)
	}

	// Branch name should not contain newlines
	if strings.Contains(branchName, "\n") {
		t.Errorf("GetCurrentBranchName() returned branch name with newline: %q", branchName)
	}

	// Branch name should not be empty
	if branchName == "" {
		t.Error("GetCurrentBranchName() returned empty branch name")
	}
}

func TestGetOriginURL(t *testing.T) {
	skipIfNotGitRepo(t)

	url, err := GetOriginURL()
	if err != nil {
		// This might fail if there's no origin remote configured
		t.Skipf("GetOriginURL() returned error (may not have origin remote): %v", err)
	}

	// URL should not be empty if no error
	if url == "" {
		t.Error("GetOriginURL() returned empty URL without error")
	}
}

func TestGetStatus(t *testing.T) {
	skipIfNotGitRepo(t)

	// GetStatus should not return an error in a valid git repo
	_, err := GetStatus()
	if err != nil {
		t.Fatalf("GetStatus() returned error: %v", err)
	}
}

func TestFetch(t *testing.T) {
	skipIfNotGitRepo(t)

	// Skip if no remote configured
	_, err := GetOriginURL()
	if err != nil {
		t.Skip("skipping TestFetch: no origin remote configured")
	}

	// Fetch should work without error (may take time on slow network)
	_, err = Fetch()
	if err != nil {
		t.Logf("Fetch() returned error (may be expected with no network): %v", err)
	}
}

func TestIsBranchExists(t *testing.T) {
	skipIfNotGitRepo(t)

	// Get current branch name first
	currentBranch, err := GetCurrentBranchName()
	if err != nil {
		t.Fatalf("GetCurrentBranchName() failed: %v", err)
	}

	tests := []struct {
		name       string
		branchName string
		wantExists bool
	}{
		{
			name:       "current branch exists",
			branchName: currentBranch,
			wantExists: true,
		},
		{
			name:       "non-existent branch",
			branchName: "this-branch-definitely-does-not-exist-12345",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := IsBranchExists(tt.branchName)
			if err != nil {
				t.Fatalf("IsBranchExists(%q) returned unexpected error: %v", tt.branchName, err)
			}
			if exists != tt.wantExists {
				t.Errorf("IsBranchExists(%q) = %v, want %v", tt.branchName, exists, tt.wantExists)
			}
		})
	}
}

func TestHasStagedFiles(t *testing.T) {
	skipIfNotGitRepo(t)

	// This test just verifies the function runs without error
	// The result depends on the current state of the repo
	_, err := HasStagedFiles()
	if err != nil {
		t.Fatalf("HasStagedFiles() returned error: %v", err)
	}
}

func TestDiffToStream(t *testing.T) {
	skipIfNotGitRepo(t)

	tests := []struct {
		name       string
		stagedOnly bool
	}{
		{
			name:       "all changes",
			stagedOnly: false,
		},
		{
			name:       "staged only",
			stagedOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := DiffToStream(tt.stagedOnly, &buf)
			if err != nil {
				t.Fatalf("DiffToStream(%v, writer) returned error: %v", tt.stagedOnly, err)
			}
			// Output may be empty if no changes, but should not error
		})
	}
}

func TestDiffStat(t *testing.T) {
	skipIfNotGitRepo(t)

	// Get current branch to use in test
	currentBranch, err := GetCurrentBranchName()
	if err != nil {
		t.Fatalf("GetCurrentBranchName() failed: %v", err)
	}

	// Test diff stat between the same branch (should produce empty diff)
	_, err = DiffStat(currentBranch, currentBranch)
	if err != nil {
		// May fail if no origin remote
		t.Logf("DiffStat() returned error (may be expected without origin): %v", err)
	}
}

func TestGetBranchCommitComments(t *testing.T) {
	skipIfNotGitRepo(t)

	// Get current branch to use in test
	currentBranch, err := GetCurrentBranchName()
	if err != nil {
		t.Fatalf("GetCurrentBranchName() failed: %v", err)
	}

	// Test getting commits between the same branch (should be empty)
	_, err = GetBranchCommitComments(currentBranch, currentBranch)
	if err != nil {
		// May fail if no origin remote
		t.Logf("GetBranchCommitComments() returned error (may be expected without origin): %v", err)
	}
}

// TestExecuteInvalidCommand tests that invalid git commands return errors
func TestExecuteInvalidCommand(t *testing.T) {
	skipIfNotGitRepo(t)

	// Test with an invalid git subcommand
	args := []string{"this-is-not-a-valid-git-command"}
	_, err := execute(args)
	if err == nil {
		t.Error("execute() with invalid command should return error")
	}
}

// TestCheckout tests checkout functionality
// Note: This test only runs in CI or when HELPER_TEST_CHECKOUT=1 is set
// to avoid accidentally changing user's working directory
func TestCheckout(t *testing.T) {
	skipIfNotGitRepo(t)

	if os.Getenv("HELPER_TEST_CHECKOUT") != "1" {
		t.Skip("skipping TestCheckout: set HELPER_TEST_CHECKOUT=1 to enable")
	}

	// Get current branch to restore later
	originalBranch, err := GetCurrentBranchName()
	if err != nil {
		t.Fatalf("GetCurrentBranchName() failed: %v", err)
	}

	// Checkout to the same branch (safe operation)
	_, err = Checkout(originalBranch)
	if err != nil {
		t.Errorf("Checkout(%q) returned error: %v", originalBranch, err)
	}
}

// Benchmark tests
func BenchmarkGetCurrentBranchName(b *testing.B) {
	if !isInGitRepo() {
		b.Skip("skipping benchmark: not in a git repository")
	}

	for i := 0; i < b.N; i++ {
		_, _ = GetCurrentBranchName()
	}
}

func BenchmarkGetStatus(b *testing.B) {
	if !isInGitRepo() {
		b.Skip("skipping benchmark: not in a git repository")
	}

	for i := 0; i < b.N; i++ {
		_, _ = GetStatus()
	}
}

func BenchmarkIsBranchExists(b *testing.B) {
	if !isInGitRepo() {
		b.Skip("skipping benchmark: not in a git repository")
	}

	branchName, _ := GetCurrentBranchName()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = IsBranchExists(branchName)
	}
}
