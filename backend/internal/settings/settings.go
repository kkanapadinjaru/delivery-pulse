package settings

import (
	"encoding/json"
	"os"
	"sync"
)

// Settings represents the application configuration managed via the UI.
type Settings struct {
	Teams         []string `json:"teams"`
	WorkItemTypes []string `json:"workItemTypes"`
	AreaPaths     []string `json:"areaPaths"`
	Activities    []string `json:"activities"`
}

// DefaultWorkItemTypes is the default set of work item types to query.
var DefaultWorkItemTypes = []string{"Bug", "Task"}

// DefaultAreaPaths is the default set of area paths to scope work items to.
var DefaultAreaPaths = []string{
	`Solvas\AssetMgmt\Compliance\Fabric`,
	`Solvas\AssetMgmt\Compliance\BC`,
	`Solvas\AssetMgmt\Portfolio\Fabric`,
}

// DefaultActivities is the default set of activity types to filter by.
var DefaultActivities = []string{"Development", "Testing", "Requirements"}

// Store handles loading and saving settings from a JSON file.
type Store struct {
	filePath string
	mu       sync.RWMutex
	current  Settings
}

// NewStore creates a settings store that persists to the given file path.
// It loads existing settings from the file if present.
func NewStore(filePath string) *Store {
	s := &Store{
		filePath: filePath,
		current: Settings{
			Teams:         []string{},
			WorkItemTypes: DefaultWorkItemTypes,
			AreaPaths:     DefaultAreaPaths,
			Activities:    DefaultActivities,
		},
	}
	s.load()
	return s
}

// Get returns the current settings.
func (s *Store) Get() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Update saves new settings to the store and persists to disk.
func (s *Store) Update(settings Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure non-nil slices
	if settings.Teams == nil {
		settings.Teams = []string{}
	}
	if settings.WorkItemTypes == nil {
		settings.WorkItemTypes = DefaultWorkItemTypes
	}
	if settings.AreaPaths == nil {
		settings.AreaPaths = DefaultAreaPaths
	}
	if settings.Activities == nil {
		settings.Activities = DefaultActivities
	}

	s.current = settings
	return s.save()
}

func (s *Store) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return // file doesn't exist yet, use defaults
	}

	var loaded Settings
	if err := json.Unmarshal(data, &loaded); err != nil {
		return // corrupt file, use defaults
	}

	if loaded.Teams != nil {
		s.current.Teams = loaded.Teams
	}
	if loaded.WorkItemTypes != nil && len(loaded.WorkItemTypes) > 0 {
		s.current.WorkItemTypes = loaded.WorkItemTypes
	}
	if loaded.AreaPaths != nil && len(loaded.AreaPaths) > 0 {
		s.current.AreaPaths = loaded.AreaPaths
	}
	if loaded.Activities != nil && len(loaded.Activities) > 0 {
		s.current.Activities = loaded.Activities
	}
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.current, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}
