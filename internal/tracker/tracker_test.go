package tracker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrackAndUntrack(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".arcane.json")
	os.WriteFile(path, []byte(`{"installed":[]}`), 0644)

	// Track
	if err := Track(path, "test-item", "claude", "abc123", []string{"file1.md", "file2.sh"}); err != nil {
		t.Fatalf("Track failed: %v", err)
	}

	tf, _ := Load(path)
	if len(tf.Installed) != 1 {
		t.Fatalf("expected 1 installed item, got %d", len(tf.Installed))
	}
	if tf.Installed[0].Name != "test-item" {
		t.Errorf("expected name 'test-item', got '%s'", tf.Installed[0].Name)
	}
	if tf.Installed[0].Tool != "claude" {
		t.Errorf("expected tool 'claude', got '%s'", tf.Installed[0].Tool)
	}

	// Track again (update)
	Track(path, "test-item", "claude", "def456", []string{"file1.md"})
	tf, _ = Load(path)
	if len(tf.Installed) != 1 {
		t.Errorf("expected 1 installed item after update, got %d", len(tf.Installed))
	}
	if tf.Installed[0].Version != "def456" {
		t.Errorf("expected version 'def456', got '%s'", tf.Installed[0].Version)
	}

	// Untrack
	if err := Untrack(path, "test-item"); err != nil {
		t.Fatalf("Untrack failed: %v", err)
	}
	tf, _ = Load(path)
	if len(tf.Installed) != 0 {
		t.Errorf("expected 0 installed items after untrack, got %d", len(tf.Installed))
	}
}

func TestTrackMultiTool(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".arcane.json")
	os.WriteFile(path, []byte(`{"installed":[]}`), 0644)

	// Track same item for two different tools
	Track(path, "commit-message", "claude", "abc123", []string{".claude/commands/commit-message.md"})
	Track(path, "commit-message", "opencode", "abc123", []string{".opencode/commands/commit-message.md"})

	tf, _ := Load(path)
	if len(tf.Installed) != 2 {
		t.Fatalf("expected 2 installed items (one per tool), got %d", len(tf.Installed))
	}

	// Verify both entries exist
	hasClaude := false
	hasOpencode := false
	for _, item := range tf.Installed {
		if item.Name == "commit-message" && item.Tool == "claude" {
			hasClaude = true
		}
		if item.Name == "commit-message" && item.Tool == "opencode" {
			hasOpencode = true
		}
	}
	if !hasClaude {
		t.Error("expected claude entry")
	}
	if !hasOpencode {
		t.Error("expected opencode entry")
	}
}

func TestFindInstalled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".arcane.json")
	os.WriteFile(path, []byte(`{"installed":[]}`), 0644)

	Track(path, "item-a", "claude", "v1", []string{})
	Track(path, "item-b", "opencode", "v2", []string{})

	tf, _ := Load(path)

	if found := FindInstalled(tf, "item-a"); found == nil {
		t.Error("expected to find item-a")
	}
	if found := FindInstalled(tf, "nonexistent"); found != nil {
		t.Error("expected nil for nonexistent item")
	}
}
