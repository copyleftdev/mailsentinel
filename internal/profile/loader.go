package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/sirupsen/logrus"

	"github.com/mailsentinel/core/pkg/types"
)

// Loader handles loading and managing email classification profiles
type Loader struct {
	directory string
	registry  *types.ProfileRegistry
	logger    *logrus.Logger
	cache     map[string]*types.Profile
}

// NewLoader creates a new profile loader
func NewLoader(directory string, logger *logrus.Logger) *Loader {
	return &Loader{
		directory: directory,
		registry: &types.ProfileRegistry{
			Profiles:     make(map[string]*types.Profile),
			Dependencies: make(map[string][]string),
			LoadOrder:    make([]string, 0),
		},
		logger: logger,
		cache:  make(map[string]*types.Profile),
	}
}

// LoadAll loads all profiles from the directory and resolves dependencies
func (l *Loader) LoadAll() error {
	l.logger.WithField("directory", l.directory).Info("Loading all profiles")
	
	// Clear existing data
	l.registry.Profiles = make(map[string]*types.Profile)
	l.registry.Dependencies = make(map[string][]string)
	l.registry.LoadOrder = make([]string, 0)
	l.cache = make(map[string]*types.Profile)
	
	// Find all YAML files
	files, err := l.findProfileFiles()
	if err != nil {
		return fmt.Errorf("failed to find profile files: %w", err)
	}
	
	// Load profiles without inheritance first
	profiles := make(map[string]*types.Profile)
	for _, file := range files {
		profile, err := l.loadProfileFile(file)
		if err != nil {
			l.logger.WithError(err).WithField("file", file).Error("Failed to load profile")
			continue
		}
		profiles[profile.ID] = profile
	}
	
	// Build dependency graph
	if err := l.buildDependencyGraph(profiles); err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}
	
	// Resolve inheritance and dependencies
	if err := l.resolveInheritance(profiles); err != nil {
		return fmt.Errorf("failed to resolve inheritance: %w", err)
	}
	
	// Store in registry
	l.registry.Profiles = profiles
	
	l.logger.WithField("profile_count", len(profiles)).Info("Successfully loaded all profiles")
	return nil
}

// findProfileFiles finds all YAML profile files in the directory
func (l *Loader) findProfileFiles() ([]string, error) {
	var files []string
	
	err := filepath.Walk(l.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			files = append(files, path)
		}
		
		return nil
	})
	
	return files, err
}

// loadProfileFile loads a single profile from a YAML file
func (l *Loader) loadProfileFile(filename string) (*types.Profile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	
	var profile types.Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", filename, err)
	}
	
	// Set timestamps
	now := time.Now()
	profile.CreatedAt = now
	profile.UpdatedAt = now
	
	// Validate profile
	if err := l.validateProfile(&profile); err != nil {
		return nil, fmt.Errorf("profile validation failed for %s: %w", filename, err)
	}
	
	l.logger.WithFields(logrus.Fields{
		"profile_id": profile.ID,
		"version":    profile.Version,
		"file":       filename,
	}).Info("Loaded profile")
	
	return &profile, nil
}

// validateProfile validates a profile's structure and content
func (l *Loader) validateProfile(profile *types.Profile) error {
	if profile.ID == "" {
		return fmt.Errorf("profile ID is required")
	}
	
	if profile.Version == "" {
		return fmt.Errorf("profile version is required")
	}
	
	if profile.Model == "" {
		return fmt.Errorf("profile model is required")
	}
	
	if profile.System == "" {
		return fmt.Errorf("profile system prompt is required")
	}
	
	// Validate confidence range
	if len(profile.Response.Validation.ConfidenceRange) != 2 {
		return fmt.Errorf("confidence range must have exactly 2 values")
	}
	
	if profile.Response.Validation.ConfidenceRange[0] < 0 || profile.Response.Validation.ConfidenceRange[1] > 1 {
		return fmt.Errorf("confidence range must be between 0 and 1")
	}
	
	if profile.Response.Validation.ConfidenceRange[0] >= profile.Response.Validation.ConfidenceRange[1] {
		return fmt.Errorf("confidence range minimum must be less than maximum")
	}
	
	// Validate model parameters
	if profile.ModelParams.Temperature < 0 || profile.ModelParams.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	
	if profile.ModelParams.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}
	
	if profile.ModelParams.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout_seconds must be positive")
	}
	
	return nil
}

// buildDependencyGraph builds the dependency graph for profiles
func (l *Loader) buildDependencyGraph(profiles map[string]*types.Profile) error {
	// Build dependency map
	for id, profile := range profiles {
		var deps []string
		
		// Add inheritance dependency
		if profile.InheritsFrom != "" {
			deps = append(deps, profile.InheritsFrom)
		}
		
		// Add explicit dependencies
		deps = append(deps, profile.DependsOn...)
		
		l.registry.Dependencies[id] = deps
	}
	
	// Topological sort to determine load order
	loadOrder, err := l.topologicalSort(profiles)
	if err != nil {
		return err
	}
	
	l.registry.LoadOrder = loadOrder
	return nil
}

// topologicalSort performs topological sorting to determine profile load order
func (l *Loader) topologicalSort(profiles map[string]*types.Profile) ([]string, error) {
	// Kahn's algorithm for topological sorting
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)
	
	// Initialize in-degree and adjacency list
	for id := range profiles {
		inDegree[id] = 0
		adjList[id] = make([]string, 0)
	}
	
	// Build graph
	for id, deps := range l.registry.Dependencies {
		for _, dep := range deps {
			if _, exists := profiles[dep]; !exists {
				return nil, fmt.Errorf("dependency %s not found for profile %s", dep, id)
			}
			adjList[dep] = append(adjList[dep], id)
			inDegree[id]++
		}
	}
	
	// Find nodes with no incoming edges
	queue := make([]string, 0)
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	
	// Process queue
	var result []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		
		// Remove edges from current node
		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}
	
	// Check for cycles
	if len(result) != len(profiles) {
		return nil, fmt.Errorf("circular dependency detected in profiles")
	}
	
	return result, nil
}

// resolveInheritance resolves profile inheritance in dependency order
func (l *Loader) resolveInheritance(profiles map[string]*types.Profile) error {
	for _, id := range l.registry.LoadOrder {
		profile := profiles[id]
		
		if profile.InheritsFrom != "" {
			parent, exists := profiles[profile.InheritsFrom]
			if !exists {
				return fmt.Errorf("parent profile %s not found for %s", profile.InheritsFrom, id)
			}
			
			// Merge with parent
			if err := l.mergeWithParent(profile, parent); err != nil {
				return fmt.Errorf("failed to merge profile %s with parent %s: %w", id, profile.InheritsFrom, err)
			}
			
			l.logger.WithFields(logrus.Fields{
				"profile_id": id,
				"parent_id":  profile.InheritsFrom,
			}).Info("Resolved profile inheritance")
		}
	}
	
	return nil
}

// mergeWithParent merges a profile with its parent profile
func (l *Loader) mergeWithParent(child, parent *types.Profile) error {
	// Merge system prompt (append to parent)
	if child.System == "" {
		child.System = parent.System
	} else if parent.System != "" {
		child.System = parent.System + "\n\n" + child.System
	}
	
	// Merge model parameters (child overrides parent)
	if child.ModelParams.Temperature == 0 {
		child.ModelParams.Temperature = parent.ModelParams.Temperature
	}
	if child.ModelParams.MaxTokens == 0 {
		child.ModelParams.MaxTokens = parent.ModelParams.MaxTokens
	}
	if child.ModelParams.TimeoutSeconds == 0 {
		child.ModelParams.TimeoutSeconds = parent.ModelParams.TimeoutSeconds
	}
	
	// Merge few-shot examples (parent first, then child)
	if len(parent.FewShot) > 0 {
		child.FewShot = append(parent.FewShot, child.FewShot...)
	}
	
	// Merge policy conditions (parent first, then child)
	if len(parent.Policy.Conditions) > 0 {
		child.Policy.Conditions = append(parent.Policy.Conditions, child.Policy.Conditions...)
	}
	
	// Merge response validation (child overrides parent)
	if len(child.Response.Validation.RequiredFields) == 0 {
		child.Response.Validation.RequiredFields = parent.Response.Validation.RequiredFields
	}
	if child.Response.Validation.ConfidenceRange[0] == 0 && child.Response.Validation.ConfidenceRange[1] == 0 {
		child.Response.Validation.ConfidenceRange = parent.Response.Validation.ConfidenceRange
	}
	
	return nil
}

// GetProfile retrieves a profile by ID
func (l *Loader) GetProfile(id string) (*types.Profile, error) {
	profile, exists := l.registry.Profiles[id]
	if !exists {
		return nil, fmt.Errorf("profile %s not found", id)
	}
	
	return profile, nil
}

// ListProfiles returns all loaded profile IDs
func (l *Loader) ListProfiles() []string {
	var ids []string
	for id := range l.registry.Profiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// GetRegistry returns the profile registry
func (l *Loader) GetRegistry() *types.ProfileRegistry {
	return l.registry
}

// Reload reloads all profiles from disk
func (l *Loader) Reload() error {
	l.logger.Info("Reloading profiles from disk")
	return l.LoadAll()
}
