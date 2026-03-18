package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestRegistry(t *testing.T, dir string) string {
	t.Helper()
	content := `{
  "version": 2,
  "tools": {
    "claude": {
      "types": {
        "command": {"targetDir": ".claude/commands"},
        "hook": {"settingsFile": ".claude/settings.json", "mergeKey": "hooks"}
      }
    },
    "opencode": {
      "types": {
        "command": {"targetDir": ".opencode/commands"},
        "plugin": {"targetDir": ".opencode/plugins"},
        "formatter": {"configFile": "opencode.json", "mergeKey": "formatter"}
      }
    }
  },
  "items": [
    {"name": "a", "tool": ["claude", "opencode"], "type": "command", "description": "Item A", "tags": ["git"], "files": [
      {"src": "items/a.md", "target": ".claude/commands/a.md", "tool": "claude"},
      {"src": "items/a.md", "target": ".opencode/commands/a.md", "tool": "opencode"}
    ], "dependencies": ["b"]},
    {"name": "b", "tool": ["claude", "opencode"], "type": "script", "description": "Item B", "tags": ["util"], "files": [], "dependencies": ["c"]},
    {"name": "c", "tool": "claude", "type": "script", "description": "Item C", "tags": [], "files": [], "dependencies": []},
    {"name": "d", "tool": "claude", "type": "hook", "description": "Item D", "tags": ["hook"], "files": [], "dependencies": []},
    {"name": "e", "tool": "opencode", "type": "plugin", "description": "Item E", "tags": ["plugin"], "files": [], "dependencies": []}
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
	if reg.Version != 2 {
		t.Errorf("expected version 2, got %d", reg.Version)
	}
	if len(reg.Items) != 5 {
		t.Errorf("expected 5 items, got %d", len(reg.Items))
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

func TestStringOrSlice(t *testing.T) {
	dir := t.TempDir()
	path := writeTestRegistry(t, dir)
	reg, _ := Load(path)

	// Multi-tool item
	a := reg.FindItem("a")
	if a == nil {
		t.Fatal("expected to find item 'a'")
	}
	if !a.Tool.Contains("claude") {
		t.Error("expected item 'a' to contain tool 'claude'")
	}
	if !a.Tool.Contains("opencode") {
		t.Error("expected item 'a' to contain tool 'opencode'")
	}

	// Single-tool item (string in JSON)
	c := reg.FindItem("c")
	if c == nil {
		t.Fatal("expected to find item 'c'")
	}
	if !c.Tool.Contains("claude") {
		t.Error("expected item 'c' to contain tool 'claude'")
	}
	if c.Tool.Contains("opencode") {
		t.Error("expected item 'c' to NOT contain tool 'opencode'")
	}
}

func TestFindItemForTool(t *testing.T) {
	dir := t.TempDir()
	path := writeTestRegistry(t, dir)
	reg, _ := Load(path)

	// Multi-tool item should be found for both
	if item := reg.FindItemForTool("a", "claude"); item == nil {
		t.Error("expected to find item 'a' for claude")
	}
	if item := reg.FindItemForTool("a", "opencode"); item == nil {
		t.Error("expected to find item 'a' for opencode")
	}

	// Claude-only item
	if item := reg.FindItemForTool("c", "claude"); item == nil {
		t.Error("expected to find item 'c' for claude")
	}
	if item := reg.FindItemForTool("c", "opencode"); item != nil {
		t.Error("expected nil for item 'c' with opencode")
	}

	// OpenCode-only item
	if item := reg.FindItemForTool("e", "opencode"); item == nil {
		t.Error("expected to find item 'e' for opencode")
	}
	if item := reg.FindItemForTool("e", "claude"); item != nil {
		t.Error("expected nil for item 'e' with claude")
	}
}

func TestItemsForTool(t *testing.T) {
	dir := t.TempDir()
	path := writeTestRegistry(t, dir)
	reg, _ := Load(path)

	claude := reg.ItemsForTool("claude")
	// a(claude+opencode), b(claude+opencode), c(claude), d(claude) = 4
	if len(claude) != 4 {
		t.Errorf("expected 4 claude items, got %d", len(claude))
	}

	opencode := reg.ItemsForTool("opencode")
	// a(claude+opencode), b(claude+opencode), e(opencode) = 3
	if len(opencode) != 3 {
		t.Errorf("expected 3 opencode items, got %d", len(opencode))
	}
}

func TestValidTypesForTool(t *testing.T) {
	dir := t.TempDir()
	path := writeTestRegistry(t, dir)
	reg, _ := Load(path)

	claudeTypes := reg.ValidTypesForTool("claude")
	if !claudeTypes["command"] || !claudeTypes["hook"] {
		t.Error("expected claude to have command and hook types")
	}

	ocTypes := reg.ValidTypesForTool("opencode")
	if !ocTypes["command"] || !ocTypes["plugin"] || !ocTypes["formatter"] {
		t.Error("expected opencode to have command, plugin, and formatter types")
	}
}

func TestFilesForTool(t *testing.T) {
	dir := t.TempDir()
	path := writeTestRegistry(t, dir)
	reg, _ := Load(path)

	a := reg.FindItem("a")
	if a == nil {
		t.Fatal("expected to find item 'a'")
	}

	claudeFiles := FilesForTool(a, "claude")
	if len(claudeFiles) != 1 {
		t.Fatalf("expected 1 claude file, got %d", len(claudeFiles))
	}
	if claudeFiles[0].Target != ".claude/commands/a.md" {
		t.Errorf("expected .claude/commands/a.md, got %s", claudeFiles[0].Target)
	}

	ocFiles := FilesForTool(a, "opencode")
	if len(ocFiles) != 1 {
		t.Fatalf("expected 1 opencode file, got %d", len(ocFiles))
	}
	if ocFiles[0].Target != ".opencode/commands/a.md" {
		t.Errorf("expected .opencode/commands/a.md, got %s", ocFiles[0].Target)
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
  "version": 2,
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
  "version": 2,
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

func TestSettingsFileForTool(t *testing.T) {
	dir := t.TempDir()
	path := writeTestRegistry(t, dir)
	reg, _ := Load(path)

	sf := reg.SettingsFileForTool("claude", "hook")
	if sf != ".claude/settings.json" {
		t.Errorf("expected .claude/settings.json, got %s", sf)
	}

	cf := reg.SettingsFileForTool("opencode", "formatter")
	if cf != "opencode.json" {
		t.Errorf("expected opencode.json, got %s", cf)
	}

	empty := reg.SettingsFileForTool("nonexistent", "hook")
	if empty != "" {
		t.Errorf("expected empty string for nonexistent tool, got %s", empty)
	}
}
