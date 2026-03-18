package installer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bensch98/arcane/internal/registry"
)

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "sub", "dst.txt")
	os.WriteFile(src, []byte("hello"), 0644)

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("cannot read dst: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(data))
	}
}

func TestRollback(t *testing.T) {
	dir := t.TempDir()

	// Create a file to be rolled back
	f := filepath.Join(dir, "file.txt")
	os.WriteFile(f, []byte("content"), 0644)

	rb := NewRollback()
	rb.TrackCopy(f)

	rb.Undo()

	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Error("expected file to be removed after rollback")
	}
}

func TestRollbackSettings(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "settings.json")
	original := []byte(`{"hooks":{}}`)
	os.WriteFile(f, original, 0644)

	rb := NewRollback()
	rb.TrackSettings(f, original)

	// Modify the file
	os.WriteFile(f, []byte(`{"hooks":{"Stop":[]}}`), 0644)

	rb.Undo()

	data, _ := os.ReadFile(f)
	if string(data) != string(original) {
		t.Errorf("expected settings to be restored, got '%s'", string(data))
	}
}

func TestMergeHook(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, ".claude", "settings.json")

	hm := &registry.HookMerge{
		Event: "Stop",
		Entry: map[string]interface{}{
			"matcher": "",
			"hooks": []interface{}{
				map[string]interface{}{
					"type":    "command",
					"command": "echo done",
				},
			},
		},
	}

	rb := NewRollback()

	// First merge — should create the file
	if err := MergeHook(settingsPath, hm, false, rb); err != nil {
		t.Fatalf("MergeHook failed: %v", err)
	}

	data, _ := os.ReadFile(settingsPath)
	var settings map[string]interface{}
	json.Unmarshal(data, &settings)

	hooks := settings["hooks"].(map[string]interface{})
	stopArr := hooks["Stop"].([]interface{})
	if len(stopArr) != 1 {
		t.Fatalf("expected 1 hook entry, got %d", len(stopArr))
	}

	// Second merge (same entry) — should not duplicate
	MergeHook(settingsPath, hm, false, rb)
	data, _ = os.ReadFile(settingsPath)
	json.Unmarshal(data, &settings)
	hooks = settings["hooks"].(map[string]interface{})
	stopArr = hooks["Stop"].([]interface{})
	if len(stopArr) != 1 {
		t.Errorf("expected no duplicate, got %d entries", len(stopArr))
	}
}

func TestRemoveHook(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")

	hm := &registry.HookMerge{
		Event: "Stop",
		Entry: map[string]interface{}{
			"matcher": "",
			"hooks": []interface{}{
				map[string]interface{}{
					"type":    "command",
					"command": "echo done",
				},
			},
		},
	}

	rb := NewRollback()
	MergeHook(settingsPath, hm, false, rb)

	if err := RemoveHook(settingsPath, hm); err != nil {
		t.Fatalf("RemoveHook failed: %v", err)
	}

	data, _ := os.ReadFile(settingsPath)
	var settings map[string]interface{}
	json.Unmarshal(data, &settings)
	hooks := settings["hooks"].(map[string]interface{})
	stopArr := hooks["Stop"].([]interface{})
	if len(stopArr) != 0 {
		t.Errorf("expected 0 entries after removal, got %d", len(stopArr))
	}
}

func TestMergeConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	cm := &registry.ConfigMerge{
		Path: "formatter.prettier",
		Value: map[string]interface{}{
			"command":    []interface{}{"npx", "prettier", "--write", "$FILE"},
			"extensions": []interface{}{".js", ".ts"},
		},
	}

	rb := NewRollback()

	// First merge — should create the file
	if err := MergeConfig(configPath, cm, false, rb); err != nil {
		t.Fatalf("MergeConfig failed: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	var config map[string]interface{}
	json.Unmarshal(data, &config)

	// Check $schema was added
	if config["$schema"] != "https://opencode.ai/config.json" {
		t.Error("expected $schema to be set for opencode.json")
	}

	// Check nested path was created
	formatter, ok := config["formatter"].(map[string]interface{})
	if !ok {
		t.Fatal("expected formatter key to be a map")
	}
	prettier, ok := formatter["prettier"].(map[string]interface{})
	if !ok {
		t.Fatal("expected formatter.prettier key to be a map")
	}
	cmd, ok := prettier["command"].([]interface{})
	if !ok {
		t.Fatal("expected formatter.prettier.command to be an array")
	}
	if len(cmd) != 4 {
		t.Errorf("expected 4 command entries, got %d", len(cmd))
	}

	// Second merge (same entry, no force) — should skip
	if err := MergeConfig(configPath, cm, false, rb); err != nil {
		t.Fatalf("MergeConfig second call failed: %v", err)
	}
}

func TestRemoveConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "opencode.json")

	cm := &registry.ConfigMerge{
		Path:  "formatter.prettier",
		Value: map[string]interface{}{"disabled": true},
	}

	rb := NewRollback()
	MergeConfig(configPath, cm, false, rb)

	if err := RemoveConfig(configPath, cm); err != nil {
		t.Fatalf("RemoveConfig failed: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	var config map[string]interface{}
	json.Unmarshal(data, &config)
	formatter, ok := config["formatter"].(map[string]interface{})
	if !ok {
		t.Fatal("expected formatter key to still exist")
	}
	if _, exists := formatter["prettier"]; exists {
		t.Error("expected prettier key to be removed")
	}
}
