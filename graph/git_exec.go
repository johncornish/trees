package graph

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// ExecGitChecker implements GitChecker by shelling out to git.
type ExecGitChecker struct{}

func (c *ExecGitChecker) HasFileChangedSince(commit, filePath string) (bool, error) {
	dir := filepath.Dir(filePath)
	cmd := exec.Command("git", "log", "--oneline", commit+"..HEAD", "--", filePath)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}
