package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bensch98/arcane/internal/registry"
	"github.com/bensch98/arcane/internal/ui"
)

// Rollback tracks mutations so they can be undone on error.
type Rollback struct {
	copiedFiles     []string
	settingsBackups map[string][]byte // path → original content
}

func NewRollback() *Rollback {
	return &Rollback{settingsBackups: make(map[string][]byte)}
}

func (rb *Rollback) TrackCopy(path string) {
	rb.copiedFiles = append(rb.copiedFiles, path)
}

func (rb *Rollback) TrackSettings(path string, original []byte) {
	if _, exists := rb.settingsBackups[path]; !exists {
		rb.settingsBackups[path] = original
	}
}

// Undo reverts all tracked mutations.
func (rb *Rollback) Undo() {
	for _, f := range rb.copiedFiles {
		os.Remove(f)
	}
	for path, content := range rb.settingsBackups {
		os.WriteFile(path, content, 0644)
	}
}

// CopyFile copies src to dst, creating directories as needed.
func CopyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// MergeHook merges a hook entry into a settings.json file (Claude Code).
func MergeHook(settingsPath string, hm *registry.HookMerge, force bool, rb *Rollback) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	// Read or create settings
	var settings map[string]interface{}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		settings = make(map[string]interface{})
	} else {
		rb.TrackSettings(settingsPath, data)
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("invalid settings JSON: %w", err)
		}
	}

	// Navigate to hooks[event]
	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
		settings["hooks"] = hooks
	}

	eventArr, _ := hooks[hm.Event].([]interface{})
	if eventArr == nil {
		eventArr = []interface{}{}
	}

	// Check for duplicate
	newEntryJSON, _ := json.Marshal(hm.Entry)
	isDuplicate := false
	duplicateIdx := -1

	for i, existing := range eventArr {
		existingMap, ok := existing.(map[string]interface{})
		if !ok {
			continue
		}
		if matchesEntry(existingMap, hm.Entry) {
			isDuplicate = true
			duplicateIdx = i
			break
		}
	}

	if isDuplicate {
		if force {
			eventArr[duplicateIdx] = hm.Entry
		}
		// else skip
	} else {
		var newEntry map[string]interface{}
		json.Unmarshal(newEntryJSON, &newEntry)
		eventArr = append(eventArr, newEntry)
	}

	hooks[hm.Event] = eventArr

	// Write back
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, append(out, '\n'), 0644)
}

// RemoveHook removes a matching hook entry from settings.json (Claude Code).
func RemoveHook(settingsPath string, hm *registry.HookMerge) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil // file doesn't exist, nothing to remove
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		return nil
	}

	eventArr, _ := hooks[hm.Event].([]interface{})
	if eventArr == nil {
		return nil
	}

	filtered := make([]interface{}, 0, len(eventArr))
	for _, existing := range eventArr {
		existingMap, ok := existing.(map[string]interface{})
		if !ok {
			filtered = append(filtered, existing)
			continue
		}
		if !matchesEntry(existingMap, hm.Entry) {
			filtered = append(filtered, existing)
		}
	}

	hooks[hm.Event] = filtered

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, append(out, '\n'), 0644)
}

// matchesEntry checks if an existing hook entry matches a new entry by matcher and hooks fields.
func matchesEntry(existing, new map[string]interface{}) bool {
	existingMatcher, _ := json.Marshal(existing["matcher"])
	newMatcher, _ := json.Marshal(new["matcher"])
	existingHooks, _ := json.Marshal(existing["hooks"])
	newHooks, _ := json.Marshal(new["hooks"])
	return string(existingMatcher) == string(newMatcher) && string(existingHooks) == string(newHooks)
}

// MergeConfig merges a config entry into a JSON config file (OpenCode).
// The path is a dot-separated key path, e.g. "formatter.prettier".
func MergeConfig(configPath string, cm *registry.ConfigMerge, force bool, rb *Rollback) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	// Read or create config
	var config map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		config = make(map[string]interface{})
	} else {
		rb.TrackSettings(configPath, data)
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("invalid config JSON: %w", err)
		}
	}

	// Navigate to the target path, creating intermediate objects as needed
	parts := strings.Split(cm.Path, ".")
	current := config
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last segment: set the value
			if _, exists := current[part]; exists && !force {
				// Key already exists, skip unless forced
				return nil
			}
			current[part] = cm.Value
		} else {
			// Intermediate segment: navigate or create
			next, ok := current[part].(map[string]interface{})
			if !ok {
				next = make(map[string]interface{})
				current[part] = next
			}
			current = next
		}
	}

	// Ensure $schema is present for opencode.json
	if filepath.Base(configPath) == "opencode.json" {
		if _, exists := config["$schema"]; !exists {
			config["$schema"] = "https://opencode.ai/config.json"
		}
	}

	// Write back
	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, append(out, '\n'), 0644)
}

// RemoveConfig removes a config entry from a JSON config file (OpenCode).
func RemoveConfig(configPath string, cm *registry.ConfigMerge) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil // file doesn't exist, nothing to remove
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	// Navigate to the parent and delete the last key
	parts := strings.Split(cm.Path, ".")
	current := config
	for i, part := range parts {
		if i == len(parts)-1 {
			delete(current, part)
		} else {
			next, ok := current[part].(map[string]interface{})
			if !ok {
				return nil // path doesn't exist
			}
			current = next
		}
	}

	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, append(out, '\n'), 0644)
}

// copyItemFiles copies file-based items (commands, scripts, skills, plugins).
func copyItemFiles(item *registry.Item, files []registry.FileRef, registryDir, targetRoot string, force, dryRun bool, rb *Rollback) ([]string, error) {
	var installedFiles []string

	for _, f := range files {
		srcPath := filepath.Join(registryDir, f.Src)
		targetPath := filepath.Join(targetRoot, f.Target)

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			fmt.Printf("    %s %s\n", ui.Red("source not found:"), srcPath)
			continue
		}

		if _, err := os.Stat(targetPath); err == nil && !force {
			fmt.Printf("    %s %s (use --force to overwrite)\n", ui.Yellow("exists:"), f.Target)
			continue
		}

		if dryRun {
			fmt.Printf("    %s %s\n", ui.Dim("would copy:"), f.Target)
			continue
		}

		if err := CopyFile(srcPath, targetPath); err != nil {
			return installedFiles, fmt.Errorf("copy failed: %w", err)
		}
		rb.TrackCopy(targetPath)
		fmt.Printf("    %s %s\n", ui.Green("copied:"), f.Target)
		installedFiles = append(installedFiles, f.Target)
	}

	// postInstall
	if item.PostInstall == "chmod +x" && !dryRun {
		for _, f := range files {
			targetPath := filepath.Join(targetRoot, f.Target)
			if _, err := os.Stat(targetPath); err == nil {
				os.Chmod(targetPath, 0755)
				fmt.Printf("    %s\n", ui.Dim("chmod +x "+f.Target))
			}
		}
	} else if item.PostInstall != "" && item.PostInstall != "chmod +x" && !dryRun {
		fmt.Printf("    %s %s (manual step)\n", ui.Yellow("postInstall:"), item.PostInstall)
	}

	return installedFiles, nil
}

// InstallItem installs a single item for a specific tool (file copy, hook merge, or config merge).
func InstallItem(item *registry.Item, reg *registry.Registry, registryDir, targetRoot, targetTool string, force, dryRun bool, rb *Rollback) ([]string, error) {
	// Filter files for this tool
	files := registry.FilesForTool(item, targetTool)

	switch item.Type {
	case "hook":
		if item.HookMerge == nil {
			return nil, fmt.Errorf("hook item '%s' has no hookMerge field", item.Name)
		}
		settingsFile := reg.SettingsFileForTool(targetTool, "hook")
		if settingsFile == "" {
			settingsFile = ".claude/settings.json" // fallback for backward compat
		}
		settingsPath := filepath.Join(targetRoot, settingsFile)
		if dryRun {
			fmt.Printf("    %s %s (event: %s)\n", ui.Dim("would merge hook into:"), settingsPath, item.HookMerge.Event)
			return nil, nil
		}
		if err := MergeHook(settingsPath, item.HookMerge, force, rb); err != nil {
			return nil, fmt.Errorf("hook merge failed: %w", err)
		}
		fmt.Printf("    %s %s → %s\n", ui.Green("merged hook:"), item.HookMerge.Event, settingsPath)
		return nil, nil

	case "formatter", "config-merge":
		if item.ConfigMerge == nil {
			return nil, fmt.Errorf("config item '%s' has no configMerge field", item.Name)
		}
		configFile := reg.SettingsFileForTool(targetTool, item.Type)
		if configFile == "" {
			configFile = "opencode.json" // fallback
		}
		configPath := filepath.Join(targetRoot, configFile)
		if dryRun {
			fmt.Printf("    %s %s (path: %s)\n", ui.Dim("would merge config into:"), configPath, item.ConfigMerge.Path)
			return nil, nil
		}
		if err := MergeConfig(configPath, item.ConfigMerge, force, rb); err != nil {
			return nil, fmt.Errorf("config merge failed: %w", err)
		}
		fmt.Printf("    %s %s → %s\n", ui.Green("merged config:"), item.ConfigMerge.Path, configPath)
		return nil, nil

	default:
		// File-based items: command, script, skill, plugin
		return copyItemFiles(item, files, registryDir, targetRoot, force, dryRun, rb)
	}
}
