package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RunCommand executes a system command and returns its trimmed stdout output.
// Stderr is also captured and included in the error message if the command fails.
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command '%s %s' failed: %w\nStderr: %s", name, strings.Join(args, " "), err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RunGitCommand is a convenience wrapper around RunCommand for git commands.
func RunGitCommand(args ...string) (string, error) {
	return RunCommand("git", args...)
}

// RunCommandInDir executes a system command in a specific directory.
func RunCommandInDir(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command '%s %s' in dir '%s' failed: %w\nStderr: %s", name, strings.Join(args, " "), dir, err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// RunGitCommandInDir is a convenience wrapper for running git commands in a specific directory.
func RunGitCommandInDir(dir string, args ...string) (string, error) {
    return RunCommandInDir(dir, "git", args...)
}
