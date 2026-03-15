package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

// MergeHook merges a hook entry into a settings.json file.
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

// RemoveHook removes a matching hook entry from settings.json.
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

// InstallItem installs a single item (file copy or hook merge).
func InstallItem(item *registry.Item, registryDir, targetRoot string, force, dryRun bool, rb *Rollback) ([]string, error) {
	var installedFiles []string

	if item.Type == "hook" {
		if item.HookMerge == nil {
			return nil, fmt.Errorf("hook item '%s' has no hookMerge field", item.Name)
		}
		settingsPath := filepath.Join(targetRoot, ".claude", "settings.json")
		if dryRun {
			fmt.Printf("    %s %s (event: %s)\n", ui.Dim("would merge hook into:"), settingsPath, item.HookMerge.Event)
			return nil, nil
		}
		if err := MergeHook(settingsPath, item.HookMerge, force, rb); err != nil {
			return nil, fmt.Errorf("hook merge failed: %w", err)
		}
		fmt.Printf("    %s %s → %s\n", ui.Green("merged hook:"), item.HookMerge.Event, settingsPath)
		return nil, nil
	}

	// File-based items
	for _, f := range item.Files {
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
		for _, f := range item.Files {
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
