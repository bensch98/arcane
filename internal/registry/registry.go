package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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

	return "", fmt.Errorf("registry not found. Set ARCANE_REGISTRY env var to your registry path")
}

