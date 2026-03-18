package git

import (
	"os/exec"
	"strings"
)

// IsRepo checks if a directory is inside a git work tree.
func IsRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

// Pull runs git pull --quiet in the given directory.
func Pull(dir string) error {
	cmd := exec.Command("git", "-C", dir, "pull", "--quiet")
	return cmd.Run()
}

// Clone clones a git repository into the given directory.
func Clone(url, dest string) error {
	cmd := exec.Command("git", "clone", "--quiet", url, dest)
	return cmd.Run()
}

// RevParseShort returns the short HEAD SHA, or "" on error.
func RevParseShort(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
