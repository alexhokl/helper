package git

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// mockExecutor is a mock implementation of CommandExecutor for testing
type mockExecutor struct {
	executeFunc         func(args []string) (string, error)
	executeToStreamFunc func(args []string, writer io.Writer) error
}

func (m *mockExecutor) Execute(args []string) (string, error) {
	if m.executeFunc != nil {
		return m.executeFunc(args)
	}
	return "", nil
}

func (m *mockExecutor) ExecuteToStream(args []string, writer io.Writer) error {
	if m.executeToStreamFunc != nil {
		return m.executeToStreamFunc(args, writer)
	}
	return nil
}

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

	// Test with an invalid git subcommand using the default executor
	ResetExecutor()
	defer ResetExecutor()

	args := []string{"this-is-not-a-valid-git-command"}
	_, err := executor.Execute(args)
	if err == nil {
		t.Error("Execute() with invalid command should return error")
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

// ============================================================================
// Mock-based tests for complete coverage
// ============================================================================

func TestCheckout_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		mockOutput string
		mockError  error
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "successful checkout",
			branchName: "main",
			mockOutput: "Switched to branch 'main'\n",
			mockError:  nil,
			wantOutput: "Switched to branch 'main'\n",
			wantErr:    false,
		},
		{
			name:       "checkout non-existent branch",
			branchName: "non-existent",
			mockOutput: "",
			mockError:  errors.New("error: pathspec 'non-existent' did not match any file(s) known to git"),
			wantOutput: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					if len(args) != 2 || args[0] != "checkout" || args[1] != tt.branchName {
						t.Errorf("unexpected args: %v", args)
					}
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			output, err := Checkout(tt.branchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Checkout() error = %v, wantErr %v", err, tt.wantErr)
			}
			if output != tt.wantOutput {
				t.Errorf("Checkout() output = %q, want %q", output, tt.wantOutput)
			}
		})
	}
}

func TestGetOriginURL_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		mockOutput string
		mockError  error
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "successful get origin URL",
			mockOutput: "git@github.com:user/repo.git\n",
			mockError:  nil,
			wantOutput: "git@github.com:user/repo.git\n",
			wantErr:    false,
		},
		{
			name:       "no origin configured",
			mockOutput: "",
			mockError:  errors.New("fatal: No such remote 'origin'"),
			wantOutput: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			output, err := GetOriginURL()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOriginURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if output != tt.wantOutput {
				t.Errorf("GetOriginURL() output = %q, want %q", output, tt.wantOutput)
			}
		})
	}
}

func TestGetStatus_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		mockOutput string
		mockError  error
		wantErr    bool
	}{
		{
			name:       "clean status",
			mockOutput: "",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "modified files",
			mockOutput: " M file1.go\n M file2.go\n",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "error not in git repo",
			mockOutput: "",
			mockError:  errors.New("fatal: not a git repository"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			output, err := GetStatus()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && output != tt.mockOutput {
				t.Errorf("GetStatus() output = %q, want %q", output, tt.mockOutput)
			}
		})
	}
}

func TestPull_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		mockOutput string
		mockError  error
		wantErr    bool
	}{
		{
			name:       "successful pull",
			mockOutput: "Already up to date.\n",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "pull with updates",
			mockOutput: "Updating abc123..def456\nFast-forward\n file.go | 10 ++++++++++\n",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "pull with conflicts",
			mockOutput: "",
			mockError:  errors.New("error: Your local changes would be overwritten by merge"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					if len(args) != 1 || args[0] != "pull" {
						t.Errorf("unexpected args: %v", args)
					}
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			output, err := Pull()
			if (err != nil) != tt.wantErr {
				t.Errorf("Pull() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && output != tt.mockOutput {
				t.Errorf("Pull() output = %q, want %q", output, tt.mockOutput)
			}
		})
	}
}

func TestFetch_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		mockOutput string
		mockError  error
		wantErr    bool
	}{
		{
			name:       "successful fetch",
			mockOutput: "",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "fetch with network error",
			mockOutput: "",
			mockError:  errors.New("fatal: unable to access: Could not resolve host"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			output, err := Fetch()
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && output != tt.mockOutput {
				t.Errorf("Fetch() output = %q, want %q", output, tt.mockOutput)
			}
		})
	}
}

func TestGetCurrentBranchName_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		mockOutput string
		mockError  error
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "successful get branch name",
			mockOutput: "main\n",
			mockError:  nil,
			wantOutput: "main",
			wantErr:    false,
		},
		{
			name:       "feature branch",
			mockOutput: "feature/test-branch\n",
			mockError:  nil,
			wantOutput: "feature/test-branch",
			wantErr:    false,
		},
		{
			name:       "detached HEAD",
			mockOutput: "",
			mockError:  errors.New("fatal: ref HEAD is not a symbolic ref"),
			wantOutput: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			output, err := GetCurrentBranchName()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentBranchName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if output != tt.wantOutput {
				t.Errorf("GetCurrentBranchName() output = %q, want %q", output, tt.wantOutput)
			}
		})
	}
}

func TestDeleteBranch_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		mockOutput string
		mockError  error
		wantErr    bool
	}{
		{
			name:       "successful delete",
			branchName: "feature-branch",
			mockOutput: "Deleted branch feature-branch (was abc1234).\n",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "branch does not exist",
			branchName: "non-existent",
			mockOutput: "",
			mockError:  errors.New("error: branch 'non-existent' not found"),
			wantErr:    true,
		},
		{
			name:       "cannot delete current branch",
			branchName: "main",
			mockOutput: "",
			mockError:  errors.New("error: Cannot delete branch 'main' checked out"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					if len(args) != 3 || args[0] != "branch" || args[1] != "-D" || args[2] != tt.branchName {
						t.Errorf("unexpected args: %v", args)
					}
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			output, err := DeleteBranch(tt.branchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && output != tt.mockOutput {
				t.Errorf("DeleteBranch() output = %q, want %q", output, tt.mockOutput)
			}
		})
	}
}

func TestMerge_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		mockOutput string
		mockError  error
		wantErr    bool
	}{
		{
			name:       "successful merge",
			branchName: "feature-branch",
			mockOutput: "Updating abc123..def456\nFast-forward\n",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "merge conflicts",
			branchName: "conflicting-branch",
			mockOutput: "",
			mockError:  errors.New("CONFLICT (content): Merge conflict in file.go"),
			wantErr:    true,
		},
		{
			name:       "already up to date",
			branchName: "main",
			mockOutput: "Already up to date.\n",
			mockError:  nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					if len(args) != 3 || args[0] != "merge" || args[1] != "--no-edit" || args[2] != tt.branchName {
						t.Errorf("unexpected args: %v", args)
					}
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			output, err := Merge(tt.branchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Merge() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && output != tt.mockOutput {
				t.Errorf("Merge() output = %q, want %q", output, tt.mockOutput)
			}
		})
	}
}

func TestDifftool_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		destBranch string
		srcBranch  string
		mockError  error
		wantErr    bool
	}{
		{
			name:       "successful difftool",
			destBranch: "main",
			srcBranch:  "feature",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "difftool error",
			destBranch: "main",
			srcBranch:  "non-existent",
			mockError:  errors.New("fatal: bad revision 'origin/non-existent'"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					expectedArg := "origin/" + tt.destBranch + "...origin/" + tt.srcBranch
					if len(args) != 2 || args[0] != "difftool" || args[1] != expectedArg {
						t.Errorf("unexpected args: %v, expected difftool %s", args, expectedArg)
					}
					return "", tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			err := Difftool(tt.destBranch, tt.srcBranch)
			if (err != nil) != tt.wantErr {
				t.Errorf("Difftool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiffStat_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		destBranch string
		srcBranch  string
		mockOutput string
		mockError  error
		wantErr    bool
	}{
		{
			name:       "successful diff stat",
			destBranch: "main",
			srcBranch:  "feature",
			mockOutput: " file1.go | 10 ++++++++++\n file2.go |  5 +++++\n 2 files changed, 15 insertions(+)\n",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "no differences",
			destBranch: "main",
			srcBranch:  "main",
			mockOutput: "",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "invalid branch",
			destBranch: "main",
			srcBranch:  "non-existent",
			mockOutput: "",
			mockError:  errors.New("fatal: bad revision"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			output, err := DiffStat(tt.destBranch, tt.srcBranch)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffStat() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && output != tt.mockOutput {
				t.Errorf("DiffStat() output = %q, want %q", output, tt.mockOutput)
			}
		})
	}
}

func TestGetBranchCommitComments_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		srcBranch  string
		destBranch string
		mockOutput string
		mockError  error
		wantErr    bool
	}{
		{
			name:       "successful get comments",
			srcBranch:  "feature",
			destBranch: "main",
			mockOutput: "'Add new feature'\n'Fix bug in feature'\n",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "no commits",
			srcBranch:  "main",
			destBranch: "main",
			mockOutput: "",
			mockError:  nil,
			wantErr:    false,
		},
		{
			name:       "invalid branch",
			srcBranch:  "non-existent",
			destBranch: "main",
			mockOutput: "",
			mockError:  errors.New("fatal: bad revision"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			output, err := GetBranchCommitComments(tt.srcBranch, tt.destBranch)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBranchCommitComments() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && output != tt.mockOutput {
				t.Errorf("GetBranchCommitComments() output = %q, want %q", output, tt.mockOutput)
			}
		})
	}
}

func TestIsBranchExists_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		mockOutput string
		mockError  error
		wantExists bool
		wantErr    bool
	}{
		{
			name:       "branch exists",
			branchName: "main",
			mockOutput: "[main] Initial commit\n",
			mockError:  nil,
			wantExists: true,
			wantErr:    false,
		},
		{
			name:       "branch does not exist (exit code 128)",
			branchName: "non-existent",
			mockOutput: "",
			mockError:  errors.New("exit status 128"),
			wantExists: false,
			wantErr:    false,
		},
		{
			name:       "other error",
			branchName: "test",
			mockOutput: "",
			mockError:  errors.New("fatal: some other error"),
			wantExists: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			exists, err := IsBranchExists(tt.branchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsBranchExists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if exists != tt.wantExists {
				t.Errorf("IsBranchExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestHasStagedFiles_WithMock(t *testing.T) {
	tests := []struct {
		name       string
		mockOutput string
		mockError  error
		wantResult bool
		wantErr    bool
	}{
		{
			name:       "has staged files",
			mockOutput: "file1.go\nfile2.go\n",
			mockError:  nil,
			wantResult: true,
			wantErr:    false,
		},
		{
			name:       "no staged files",
			mockOutput: "",
			mockError:  nil,
			wantResult: false,
			wantErr:    false,
		},
		{
			name:       "error",
			mockOutput: "",
			mockError:  errors.New("fatal: not a git repository"),
			wantResult: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeFunc: func(args []string) (string, error) {
					return tt.mockOutput, tt.mockError
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			result, err := HasStagedFiles()
			if (err != nil) != tt.wantErr {
				t.Errorf("HasStagedFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
			if result != tt.wantResult {
				t.Errorf("HasStagedFiles() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestDiffToStream_WithMock(t *testing.T) {
	tests := []struct {
		name        string
		stagedOnly  bool
		mockOutput  string
		mockError   error
		wantErr     bool
		expectCache bool // whether --cached should be in args
	}{
		{
			name:        "all changes",
			stagedOnly:  false,
			mockOutput:  "diff --git a/file.go b/file.go\n",
			mockError:   nil,
			wantErr:     false,
			expectCache: false,
		},
		{
			name:        "staged only",
			stagedOnly:  true,
			mockOutput:  "diff --git a/staged.go b/staged.go\n",
			mockError:   nil,
			wantErr:     false,
			expectCache: true,
		},
		{
			name:        "error",
			stagedOnly:  false,
			mockOutput:  "",
			mockError:   errors.New("git error"),
			wantErr:     true,
			expectCache: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				executeToStreamFunc: func(args []string, writer io.Writer) error {
					// Verify args
					if args[0] != "--no-pager" || args[1] != "diff" {
						t.Errorf("unexpected args prefix: %v", args)
					}
					hasCached := len(args) > 2 && args[2] == "--cached"
					if hasCached != tt.expectCache {
						t.Errorf("expected --cached=%v, got %v", tt.expectCache, hasCached)
					}

					if tt.mockError != nil {
						return tt.mockError
					}
					_, err := writer.Write([]byte(tt.mockOutput))
					return err
				},
			}
			SetExecutor(mock)
			defer ResetExecutor()

			var buf bytes.Buffer
			err := DiffToStream(tt.stagedOnly, &buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffToStream() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetExecutor_AndResetExecutor(t *testing.T) {
	// Test that SetExecutor and ResetExecutor work correctly
	originalExecutor := executor

	mock := &mockExecutor{
		executeFunc: func(args []string) (string, error) {
			return "mock output", nil
		},
	}

	SetExecutor(mock)
	if executor != mock {
		t.Error("SetExecutor did not set the executor")
	}

	ResetExecutor()
	// After reset, executor should be a defaultExecutor
	if _, ok := executor.(*defaultExecutor); !ok {
		t.Error("ResetExecutor did not reset to defaultExecutor")
	}

	// Restore original
	executor = originalExecutor
}

func TestDefaultExecutor_Execute(t *testing.T) {
	skipIfNotGitRepo(t)

	exec := &defaultExecutor{}

	// Test with a simple git command
	output, err := exec.Execute([]string{"rev-parse", "--is-inside-work-tree"})
	if err != nil {
		t.Errorf("defaultExecutor.Execute() error = %v", err)
	}
	if !strings.Contains(output, "true") {
		t.Errorf("defaultExecutor.Execute() output = %q, expected 'true'", output)
	}
}

func TestDefaultExecutor_ExecuteToStream(t *testing.T) {
	skipIfNotGitRepo(t)

	exec := &defaultExecutor{}
	var buf bytes.Buffer

	// Test with a simple git command
	err := exec.ExecuteToStream([]string{"rev-parse", "--is-inside-work-tree"}, &buf)
	if err != nil {
		t.Errorf("defaultExecutor.ExecuteToStream() error = %v", err)
	}
	if !strings.Contains(buf.String(), "true") {
		t.Errorf("defaultExecutor.ExecuteToStream() output = %q, expected 'true'", buf.String())
	}
}
