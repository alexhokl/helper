package git

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Checkout executes git checkout commnad
func Checkout(branchName string) (string, error) {
	args := []string{"checkout", branchName}
	return execute(args)
}

// GetOriginURL executes git remote command to retrieve origin URL
func GetOriginURL() (string, error) {
	args := []string{"remote", "get-url", "origin"}
	return execute(args)
}

// GetStatus executes git status command
func GetStatus() (string, error) {
	args := []string{"status", "-s"}
	return execute(args)
}

// Pull executes git pull command
func Pull() (string, error) {
	args := []string{"pull"}
	return execute(args)
}

// Fetch executes git fetch command
func Fetch() (string, error) {
	args := []string{"fetch"}
	return execute(args)
}

// GetCurrentBranchName executes git symbolic-ref to retrieve current branch name
func GetCurrentBranchName() (string, error) {
	args := []string{"symbolic-ref", "--short", "HEAD"}
	output, err := execute(args)
	if err != nil {
		return output, err
	}
	name := strings.ReplaceAll(output, "\n", "")
	return name, nil
}

// DeleteBranch executes git branch command to delete a branch
func DeleteBranch(branchName string) (string, error) {
	args := []string{"branch", "-D", branchName}
	return execute(args)
}

// Merge executes git merge command to merge from a branch
func Merge(branchName string) (string, error) {
	args := []string{"merge", "--no-edit", branchName}
	return execute(args)
}

// Difftool executes git difftool command
func Difftool(destinationBranchName, sourceBranchName string) error {
	args := []string{"difftool", fmt.Sprintf("origin/%s...origin/%s", destinationBranchName, sourceBranchName)}
	_, err := execute(args)
	return err
}

// DiffStat executes git diff to retrieve diff stat
func DiffStat(destinationBranchName, sourceBranchName string) (string, error) {
	args := []string{"diff", "--stat", fmt.Sprintf("origin/%s...origin/%s", destinationBranchName, sourceBranchName)}
	return execute(args)
}

// GetBranchCommitComments executes git log command to retrieve branch commit comments
func GetBranchCommitComments(sourceBranchName string, destinationBranchName string) (string, error) {
	branches := fmt.Sprintf("origin/%s..origin/%s", destinationBranchName, sourceBranchName)
	args := []string{"log", branches, "--no-merges", "--pretty=format:'%s %b'"}
	return execute(args)
}

// IsBranchExists executes git show-branch to check if a branch exists
func IsBranchExists(branchName string) (bool, error) {
	args := []string{"show-branch", branchName}
	_, err := execute(args)
	if err != nil {
		if strings.Contains(err.Error(), "128") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func HasStagedFiles() (bool, error) {
	args := []string{"diff", "--name-only", "--cached"}
	output, err := execute(args)
	if err != nil {
		return false, err
	}
	return len(output) > 0, nil
}

func DiffToStream(stagedOnly bool, writer io.Writer) error {
	args := []string{"--no-pager", "diff"}
	if stagedOnly {
		args = append(args, "--cached")
	}
	cmd := exec.Command("git", args...)
	cmd.Stdout = writer
	return cmd.Run()
}

func execute(args []string) (string, error) {
	byteOutput, err := exec.Command("git", args...).Output()
	return string(byteOutput), err
}
