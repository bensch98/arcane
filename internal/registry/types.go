package registry

// Registry is the top-level structure of registry.json.
type Registry struct {
	Version int                `json:"version"`
	Tools   map[string]ToolDef `json:"tools"`
	Items   []Item             `json:"items"`
}

// ToolDef describes tool-level metadata (e.g. "claude").
type ToolDef struct {
	Types map[string]TypeDef `json:"types"`
}

// TypeDef describes how a type maps to the filesystem.
type TypeDef struct {
	TargetDir    string `json:"targetDir,omitempty"`
	SettingsFile string `json:"settingsFile,omitempty"`
	MergeKey     string `json:"mergeKey,omitempty"`
}

// Item is a single installable registry entry.
type Item struct {
	Name         string     `json:"name"`
	Tool         string     `json:"tool"`
	Type         string     `json:"type"`
	Description  string     `json:"description"`
	Tags         []string   `json:"tags"`
	Files        []FileRef  `json:"files"`
	Dependencies []string   `json:"dependencies,omitempty"`
	PostInstall  string     `json:"postInstall,omitempty"`
	HookMerge    *HookMerge `json:"hookMerge,omitempty"`
}

// FileRef maps a source path to a target path.
type FileRef struct {
	Src    string `json:"src"`
	Target string `json:"target"`
}

// HookMerge describes how to merge a hook into settings.json.
type HookMerge struct {
	Event string                 `json:"event"`
	Entry map[string]interface{} `json:"entry"`
}

// TrackingFile is the structure of .arcane.json.
type TrackingFile struct {
	Installed []InstalledItem `json:"installed"`
}

// InstalledItem tracks a single installed registry item.
type InstalledItem struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Files   []string `json:"files"`
}
