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
	if err := Track(path, "test-item", "abc123", []string{"file1.md", "file2.sh"}); err != nil {
		t.Fatalf("Track failed: %v", err)
	}

	tf, _ := Load(path)
	if len(tf.Installed) != 1 {
		t.Fatalf("expected 1 installed item, got %d", len(tf.Installed))
	}
	if tf.Installed[0].Name != "test-item" {
		t.Errorf("expected name 'test-item', got '%s'", tf.Installed[0].Name)
	}

	// Track again (update)
	Track(path, "test-item", "def456", []string{"file1.md"})
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

func TestFindInstalled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".arcane.json")
	os.WriteFile(path, []byte(`{"installed":[]}`), 0644)

	Track(path, "item-a", "v1", []string{})
	Track(path, "item-b", "v2", []string{})

	tf, _ := Load(path)

	if found := FindInstalled(tf, "item-a"); found == nil {
		t.Error("expected to find item-a")
	}
	if found := FindInstalled(tf, "nonexistent"); found != nil {
		t.Error("expected nil for nonexistent item")
	}
}
