package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestRegistry(t *testing.T, dir string) string {
	t.Helper()
	content := `{
  "version": 1,
  "tools": {"claude": {"types": {"command": {"targetDir": ".claude/commands"}}}},
  "items": [
    {"name": "a", "tool": "claude", "type": "command", "description": "Item A", "tags": ["git"], "files": [], "dependencies": ["b"]},
    {"name": "b", "tool": "claude", "type": "script", "description": "Item B", "tags": ["util"], "files": [], "dependencies": ["c"]},
    {"name": "c", "tool": "claude", "type": "script", "description": "Item C", "tags": [], "files": [], "dependencies": []},
    {"name": "d", "tool": "claude", "type": "hook", "description": "Item D", "tags": ["hook"], "files": [], "dependencies": []}
  ]
}`
	path := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := writeTestRegistry(t, dir)

	reg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if reg.Version != 1 {
		t.Errorf("expected version 1, got %d", reg.Version)
	}
	if len(reg.Items) != 4 {
		t.Errorf("expected 4 items, got %d", len(reg.Items))
	}
}

func TestFindItem(t *testing.T) {
	dir := t.TempDir()
	path := writeTestRegistry(t, dir)
	reg, _ := Load(path)

	if item := reg.FindItem("a"); item == nil {
		t.Error("expected to find item 'a'")
	}
	if item := reg.FindItem("nonexistent"); item != nil {
		t.Error("expected nil for nonexistent item")
	}
}

func TestResolveDeps(t *testing.T) {
	dir := t.TempDir()
	path := writeTestRegistry(t, dir)
	reg, _ := Load(path)

	// a → b → c
	result, err := reg.ResolveDeps("a")
	if err != nil {
		t.Fatalf("ResolveDeps failed: %v", err)
	}
	expected := []string{"c", "b", "a"}
	if len(result) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
	for i, name := range expected {
		if result[i] != name {
			t.Errorf("position %d: expected %s, got %s", i, name, result[i])
		}
	}
}

func TestResolveDepsNoDeps(t *testing.T) {
	dir := t.TempDir()
	path := writeTestRegistry(t, dir)
	reg, _ := Load(path)

	result, err := reg.ResolveDeps("d")
	if err != nil {
		t.Fatalf("ResolveDeps failed: %v", err)
	}
	if len(result) != 1 || result[0] != "d" {
		t.Errorf("expected [d], got %v", result)
	}
}

func TestResolveDepsCircular(t *testing.T) {
	dir := t.TempDir()
	content := `{
  "version": 1,
  "tools": {},
  "items": [
    {"name": "x", "tool": "claude", "type": "command", "description": "", "tags": [], "files": [], "dependencies": ["y"]},
    {"name": "y", "tool": "claude", "type": "command", "description": "", "tags": [], "files": [], "dependencies": ["x"]}
  ]
}`
	path := filepath.Join(dir, "registry.json")
	os.WriteFile(path, []byte(content), 0644)
	reg, _ := Load(path)

	_, err := reg.ResolveDeps("x")
	if err == nil {
		t.Error("expected circular dependency error")
	}
}

func TestResolveDepsMissing(t *testing.T) {
	dir := t.TempDir()
	content := `{
  "version": 1,
  "tools": {},
  "items": [
    {"name": "x", "tool": "claude", "type": "command", "description": "", "tags": [], "files": [], "dependencies": ["missing"]}
  ]
}`
	path := filepath.Join(dir, "registry.json")
	os.WriteFile(path, []byte(content), 0644)
	reg, _ := Load(path)

	_, err := reg.ResolveDeps("x")
	if err == nil {
		t.Error("expected missing dependency error")
	}
}
