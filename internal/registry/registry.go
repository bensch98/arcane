package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/bensch98/arcane/internal/git"
)

const DefaultRegistryURL = "https://github.com/bensch98/arcane.git"

// Load reads and parses a registry.json file.
func Load(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read registry: %w", err)
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("invalid registry JSON: %w", err)
	}
	return &reg, nil
}

// FindItem returns the item with the given name, or nil.
func (r *Registry) FindItem(name string) *Item {
	for i := range r.Items {
		if r.Items[i].Name == name {
			return &r.Items[i]
		}
	}
	return nil
}

// FindItemForTool returns the item with the given name only if it supports the given tool.
func (r *Registry) FindItemForTool(name, tool string) *Item {
	item := r.FindItem(name)
	if item == nil {
		return nil
	}
	if !item.Tool.Contains(tool) {
		return nil
	}
	return item
}

// ItemsForTool returns all items that support the given tool.
func (r *Registry) ItemsForTool(tool string) []Item {
	var result []Item
	for _, item := range r.Items {
		if item.Tool.Contains(tool) {
			result = append(result, item)
		}
	}
	return result
}

// ValidTypesForTool returns the set of valid type names for a given tool.
func (r *Registry) ValidTypesForTool(tool string) map[string]bool {
	types := make(map[string]bool)
	td, ok := r.Tools[tool]
	if !ok {
		return types
	}
	for typeName := range td.Types {
		types[typeName] = true
	}
	return types
}

// ToolDirs returns the config directory names that indicate a tool is configured
// in the project. Used for auto-detection.
func ToolDirs() map[string]string {
	return map[string]string{
		"claude":   ".claude",
		"opencode": ".opencode",
	}
}

// DetectTools returns which tools are configured in the given project root
// by checking for their config directories.
func DetectTools(projectRoot string) []string {
	var tools []string
	for tool, dir := range ToolDirs() {
		if _, err := os.Stat(filepath.Join(projectRoot, dir)); err == nil {
			tools = append(tools, tool)
		}
	}
	return tools
}

// FilesForTool filters an item's file refs for a specific tool.
// Returns file refs that either have no tool restriction or match the given tool.
func FilesForTool(item *Item, tool string) []FileRef {
	var result []FileRef
	for _, f := range item.Files {
		if f.Tool == "" || f.Tool == tool {
			result = append(result, f)
		}
	}
	return result
}

// SettingsFileForTool returns the settings/config file path for a given item type and tool.
func (r *Registry) SettingsFileForTool(tool, itemType string) string {
	td, ok := r.Tools[tool]
	if !ok {
		return ""
	}
	typeDef, ok := td.Types[itemType]
	if !ok {
		return ""
	}
	if typeDef.SettingsFile != "" {
		return typeDef.SettingsFile
	}
	return typeDef.ConfigFile
}

// ResolveDeps returns items in topological (dependency-first) order.
func (r *Registry) ResolveDeps(name string) ([]string, error) {
	visited := map[string]bool{}
	pending := map[string]bool{}
	var result []string

	var visit func(string) error
	visit = func(n string) error {
		if visited[n] {
			return nil
		}
		if pending[n] {
			return fmt.Errorf("circular dependency detected: %s", n)
		}
		pending[n] = true

		item := r.FindItem(n)
		if item == nil {
			return fmt.Errorf("dependency '%s' not found in registry", n)
		}

		for _, dep := range item.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		delete(pending, n)
		visited[n] = true
		result = append(result, n)
		return nil
	}

	if err := visit(name); err != nil {
		return nil, err
	}
	return result, nil
}

// ItemsByType returns all items matching the given type.
func (r *Registry) ItemsByType(typ string) []Item {
	var result []Item
	for _, item := range r.Items {
		if item.Type == typ {
			result = append(result, item)
		}
	}
	return result
}

// ResolveMultipleDeps resolves dependencies for multiple items, deduplicating.
func (r *Registry) ResolveMultipleDeps(names []string) ([]string, error) {
	seen := map[string]bool{}
	var result []string
	for _, name := range names {
		deps, err := r.ResolveDeps(name)
		if err != nil {
			return nil, err
		}
		for _, d := range deps {
			if !seen[d] {
				seen[d] = true
				result = append(result, d)
			}
		}
	}
	return result, nil
}

// FindRegistryDir locates the registry directory.
// Priority: $ARCANE_REGISTRY, then directory containing the binary, then ~/repos/arcane.
func FindRegistryDir() (string, error) {
	// 1. Environment variable
	if dir := os.Getenv("ARCANE_REGISTRY"); dir != "" {
		if _, err := os.Stat(filepath.Join(dir, "registry.json")); err == nil {
			return dir, nil
		}
	}

	// 2. Directory containing the binary
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		// If binary is in arcane-go/, check parent
		if _, err := os.Stat(filepath.Join(dir, "registry.json")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if _, err := os.Stat(filepath.Join(parent, "registry.json")); err == nil {
			return parent, nil
		}
	}

	// 3. Default fallback
	home, _ := os.UserHomeDir()
	def := filepath.Join(home, "repos", "arcane")
	if _, err := os.Stat(filepath.Join(def, "registry.json")); err == nil {
		return def, nil
	}

	// 4. Check cache directory
	cacheDir := CacheDir()
	if _, err := os.Stat(filepath.Join(cacheDir, "registry.json")); err == nil {
		return cacheDir, nil
	}

	return "", fmt.Errorf("registry not found. Run 'arcane registry fetch' or set ARCANE_REGISTRY env var")
}

// CacheDir returns the platform-specific cache directory for the registry.
func CacheDir() string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "arcane", "registry")
	}
	// Use XDG_DATA_HOME on Linux if set, otherwise ~/.local/share
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, "arcane", "registry")
	}
	return filepath.Join(home, ".local", "share", "arcane", "registry")
}

// EnsureRegistry finds or fetches the registry, returning its directory.
// If no local registry is found, it clones the default registry to the cache directory.
func EnsureRegistry() (string, bool, error) {
	dir, err := FindRegistryDir()
	if err == nil {
		return dir, false, nil
	}

	// Auto-clone to cache dir
	cacheDir := CacheDir()
	if err := os.MkdirAll(filepath.Dir(cacheDir), 0755); err != nil {
		return "", false, fmt.Errorf("cannot create cache directory: %w", err)
	}

	if err := git.Clone(DefaultRegistryURL, cacheDir); err != nil {
		return "", false, fmt.Errorf("failed to fetch registry from %s: %w", DefaultRegistryURL, err)
	}

	return cacheDir, true, nil
}
