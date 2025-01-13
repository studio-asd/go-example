package git

import (
	"context"
	"os/exec"
	"strings"
)

// RepositoryRoot returns the path of root repository using git command. The function returns error if the git command is not exist.
func RepositoryRoot() (string, error) {
	_, err := exec.LookPath("git")
	if err != nil {
		return "", err
	}
	cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(out), "\n", ""), nil
}
