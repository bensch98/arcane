package registry

import "encoding/json"

// Registry is the top-level structure of registry.json.
type Registry struct {
	Version int                `json:"version"`
	Tools   map[string]ToolDef `json:"tools"`
	Items   []Item             `json:"items"`
}

// ToolDef describes tool-level metadata (e.g. "claude", "opencode").
type ToolDef struct {
	Types map[string]TypeDef `json:"types"`
}

// TypeDef describes how a type maps to the filesystem.
type TypeDef struct {
	TargetDir    string `json:"targetDir,omitempty"`
	SettingsFile string `json:"settingsFile,omitempty"` // Claude hooks: merge into this file
	ConfigFile   string `json:"configFile,omitempty"`   // OpenCode: merge into this config file
	MergeKey     string `json:"mergeKey,omitempty"`
}

// Item is a single installable registry entry.
type Item struct {
	Name         string        `json:"name"`
	Tool         StringOrSlice `json:"tool"`
	Type         string        `json:"type"`
	Description  string        `json:"description"`
	Tags         []string      `json:"tags"`
	Files        []FileRef     `json:"files"`
	Dependencies []string      `json:"dependencies,omitempty"`
	PostInstall  string        `json:"postInstall,omitempty"`
	HookMerge    *HookMerge    `json:"hookMerge,omitempty"`
	ConfigMerge  *ConfigMerge  `json:"configMerge,omitempty"`
}

// FileRef maps a source path to a target path.
// If Tool is set, this file ref only applies to that specific tool.
type FileRef struct {
	Src    string `json:"src"`
	Target string `json:"target"`
	Tool   string `json:"tool,omitempty"` // optional: restrict this file to a specific tool
}

// HookMerge describes how to merge a hook into a settings file (Claude Code).
type HookMerge struct {
	Event string                 `json:"event"`
	Entry map[string]interface{} `json:"entry"`
}

// ConfigMerge describes how to merge a config entry into a JSON config file (OpenCode).
type ConfigMerge struct {
	Path  string      `json:"path"`  // dot-separated JSON path, e.g. "formatter.prettier"
	Value interface{} `json:"value"` // value to set/merge at that path
}

// StringOrSlice is a custom type that accepts either a single string or an array
// of strings in JSON. This enables backward compatibility: "tool": "claude"
// works the same as "tool": ["claude"].
type StringOrSlice []string

func (s *StringOrSlice) UnmarshalJSON(data []byte) error {
	// Try single string first
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = StringOrSlice{single}
		return nil
	}
	// Try array of strings
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*s = StringOrSlice(arr)
	return nil
}

func (s StringOrSlice) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal(s[0])
	}
	return json.Marshal([]string(s))
}

// Contains returns true if the slice contains the given value.
func (s StringOrSlice) Contains(val string) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}
	return false
}

// String returns a display-friendly representation.
func (s StringOrSlice) String() string {
	if len(s) == 1 {
		return s[0]
	}
	result := ""
	for i, v := range s {
		if i > 0 {
			result += ", "
		}
		result += v
	}
	return result
}

// TrackingFile is the structure of .arcane.json.
type TrackingFile struct {
	Installed []InstalledItem `json:"installed"`
}

// InstalledItem tracks a single installed registry item.
type InstalledItem struct {
	Name    string   `json:"name"`
	Tool    string   `json:"tool,omitempty"` // which tool this was installed for
	Version string   `json:"version"`
	Files   []string `json:"files"`
}
