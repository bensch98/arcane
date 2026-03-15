package tracker

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bensch98/arcane/internal/registry"
)

const TrackingFileName = ".arcane.json"

// Load reads and parses the tracking file.
func Load(path string) (*registry.TrackingFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tf registry.TrackingFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("invalid tracking file: %w", err)
	}
	return &tf, nil
}

// Save writes the tracking file to disk.
func Save(path string, tf *registry.TrackingFile) error {
	data, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// Track adds or updates an installed item in the tracking file.
func Track(path string, name string, version string, files []string) error {
	tf, err := Load(path)
	if err != nil {
		return err
	}

	// Remove existing entry for this name
	filtered := make([]registry.InstalledItem, 0, len(tf.Installed))
	for _, item := range tf.Installed {
		if item.Name != name {
			filtered = append(filtered, item)
		}
	}
	filtered = append(filtered, registry.InstalledItem{
		Name:    name,
		Version: version,
		Files:   files,
	})
	tf.Installed = filtered

	return Save(path, tf)
}

// Untrack removes an installed item from the tracking file.
func Untrack(path string, name string) error {
	tf, err := Load(path)
	if err != nil {
		return err
	}

	filtered := make([]registry.InstalledItem, 0, len(tf.Installed))
	for _, item := range tf.Installed {
		if item.Name != name {
			filtered = append(filtered, item)
		}
	}
	tf.Installed = filtered

	return Save(path, tf)
}

// FindInstalled returns the tracked item with the given name, or nil.
func FindInstalled(tf *registry.TrackingFile, name string) *registry.InstalledItem {
	for i := range tf.Installed {
		if tf.Installed[i].Name == name {
			return &tf.Installed[i]
		}
	}
	return nil
}
