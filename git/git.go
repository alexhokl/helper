package git

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// CommandExecutor defines the interface for executing git commands
type CommandExecutor interface {
	Execute(args []string) (string, error)
	ExecuteToStream(args []string, writer io.Writer) error
}

// defaultExecutor is the default implementation that runs real git commands
type defaultExecutor struct{}

func (e *defaultExecutor) Execute(args []string) (string, error) {
	byteOutput, err := exec.Command("git", args...).Output()
	return string(byteOutput), err
}

func (e *defaultExecutor) ExecuteToStream(args []string, writer io.Writer) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = writer
	return cmd.Run()
}

// executor is the current command executor (can be replaced for testing)
var executor CommandExecutor = &defaultExecutor{}

// SetExecutor sets a custom command executor (useful for testing)
func SetExecutor(e CommandExecutor) {
	executor = e
}

// ResetExecutor resets to the default command executor
func ResetExecutor() {
	executor = &defaultExecutor{}
}

// Checkout executes git checkout commnad
func Checkout(branchName string) (string, error) {
	args := []string{"checkout", branchName}
	return executor.Execute(args)
}

// GetOriginURL executes git remote command to retrieve origin URL
func GetOriginURL() (string, error) {
	args := []string{"remote", "get-url", "origin"}
	return executor.Execute(args)
}

// GetStatus executes git status command
func GetStatus() (string, error) {
	args := []string{"status", "-s"}
	return executor.Execute(args)
}

// Pull executes git pull command
func Pull() (string, error) {
	args := []string{"pull"}
	return executor.Execute(args)
}

// Fetch executes git fetch command
func Fetch() (string, error) {
	args := []string{"fetch"}
	return executor.Execute(args)
}

// GetCurrentBranchName executes git symbolic-ref to retrieve current branch name
func GetCurrentBranchName() (string, error) {
	args := []string{"symbolic-ref", "--short", "HEAD"}
	output, err := executor.Execute(args)
	if err != nil {
		return output, err
	}
	name := strings.ReplaceAll(output, "\n", "")
	return name, nil
}

// DeleteBranch executes git branch command to delete a branch
func DeleteBranch(branchName string) (string, error) {
	args := []string{"branch", "-D", branchName}
	return executor.Execute(args)
}

// Merge executes git merge command to merge from a branch
func Merge(branchName string) (string, error) {
	args := []string{"merge", "--no-edit", branchName}
	return executor.Execute(args)
}

// Difftool executes git difftool command
func Difftool(destinationBranchName, sourceBranchName string) error {
	args := []string{"difftool", fmt.Sprintf("origin/%s...origin/%s", destinationBranchName, sourceBranchName)}
	_, err := executor.Execute(args)
	return err
}

// DiffStat executes git diff to retrieve diff stat
func DiffStat(destinationBranchName, sourceBranchName string) (string, error) {
	args := []string{"diff", "--stat", fmt.Sprintf("origin/%s...origin/%s", destinationBranchName, sourceBranchName)}
	return executor.Execute(args)
}

// GetBranchCommitComments executes git log command to retrieve branch commit comments
func GetBranchCommitComments(sourceBranchName string, destinationBranchName string) (string, error) {
	branches := fmt.Sprintf("origin/%s..origin/%s", destinationBranchName, sourceBranchName)
	args := []string{"log", branches, "--no-merges", "--pretty=format:'%s %b'"}
	return executor.Execute(args)
}

// IsBranchExists executes git show-branch to check if a branch exists
func IsBranchExists(branchName string) (bool, error) {
	args := []string{"show-branch", branchName}
	_, err := executor.Execute(args)
	if err != nil {
		if strings.Contains(err.Error(), "128") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// HasStagedFiles checks if there are staged files in the repository
func HasStagedFiles() (bool, error) {
	args := []string{"diff", "--name-only", "--cached"}
	output, err := executor.Execute(args)
	if err != nil {
		return false, err
	}
	return len(output) > 0, nil
}

// DiffToStream executes git diff and writes output to the provided writer
func DiffToStream(stagedOnly bool, writer io.Writer) error {
	args := []string{"--no-pager", "diff"}
	if stagedOnly {
		args = append(args, "--cached")
	}
	return executor.ExecuteToStream(args, writer)
}
