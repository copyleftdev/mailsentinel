package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mailsentinel/core/pkg/types"
)

func TestNewLoader(t *testing.T) {
	logger := logrus.New()
	loader := NewLoader("test_profiles", logger)

	assert.Equal(t, "test_profiles", loader.directory)
	assert.NotNil(t, loader.registry)
	assert.NotNil(t, loader.logger)
	assert.NotNil(t, loader.cache)
}

func TestLoadProfileFile(t *testing.T) {
	tempDir := t.TempDir()
	logger := logrus.New()
	loader := NewLoader(tempDir, logger)

	// Create test profile
	profileContent := `
id: "test_profile"
version: "1.0.0"
model: "qwen2.5:7b"
system: "Test system prompt"
model_params:
  temperature: 0.1
  max_tokens: 1000
  timeout_seconds: 30
response:
  schema: "{}"
  validation:
    required_fields: ["action"]
    confidence_range: [0.0, 1.0]
fewshot:
  - name: "test_example"
    input: "test input"
    output: "test output"
policy:
  conditions:
    - name: "test_condition"
      expression: "confidence > 0.5"
      actions: ["archive"]
      priority: 100
`

	profilePath := filepath.Join(tempDir, "test.yaml")
	err := os.WriteFile(profilePath, []byte(profileContent), 0644)
	require.NoError(t, err)

	profile, err := loader.loadProfileFile(profilePath)
	require.NoError(t, err)

	assert.Equal(t, "test_profile", profile.ID)
	assert.Equal(t, "1.0.0", profile.Version)
	assert.Equal(t, "qwen2.5:7b", profile.Model)
	assert.Equal(t, "Test system prompt", profile.System)
	assert.Equal(t, 0.1, profile.ModelParams.Temperature)
	assert.Equal(t, 1000, profile.ModelParams.MaxTokens)
	assert.Equal(t, 30, profile.ModelParams.TimeoutSeconds)
	assert.Len(t, profile.FewShot, 1)
	assert.Equal(t, "test_example", profile.FewShot[0].Name)
	assert.Len(t, profile.Policy.Conditions, 1)
	assert.Equal(t, "test_condition", profile.Policy.Conditions[0].Name)
}

func TestValidateProfile(t *testing.T) {
	logger := logrus.New()
	loader := NewLoader("", logger)

	tests := []struct {
		name    string
		profile *types.Profile
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_profile",
			profile: validTestProfile(),
			wantErr: false,
		},
		{
			name: "missing_id",
			profile: func() *types.Profile {
				p := validTestProfile()
				p.ID = ""
				return p
			}(),
			wantErr: true,
			errMsg:  "profile ID is required",
		},
		{
			name: "missing_version",
			profile: func() *types.Profile {
				p := validTestProfile()
				p.Version = ""
				return p
			}(),
			wantErr: true,
			errMsg:  "profile version is required",
		},
		{
			name: "missing_model",
			profile: func() *types.Profile {
				p := validTestProfile()
				p.Model = ""
				return p
			}(),
			wantErr: true,
			errMsg:  "profile model is required",
		},
		{
			name: "missing_system",
			profile: func() *types.Profile {
				p := validTestProfile()
				p.System = ""
				return p
			}(),
			wantErr: true,
			errMsg:  "profile system prompt is required",
		},
		{
			name: "invalid_confidence_range",
			profile: func() *types.Profile {
				p := validTestProfile()
				p.Response.Validation.ConfidenceRange = [2]float64{0.5, 0.3}
				return p
			}(),
			wantErr: true,
			errMsg:  "confidence range minimum must be less than maximum",
		},
		{
			name: "invalid_temperature",
			profile: func() *types.Profile {
				p := validTestProfile()
				p.ModelParams.Temperature = 3.0
				return p
			}(),
			wantErr: true,
			errMsg:  "temperature must be between 0 and 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.validateProfile(tt.profile)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTopologicalSort(t *testing.T) {
	logger := logrus.New()
	loader := NewLoader("", logger)

	// Create profiles with dependencies
	profiles := map[string]*types.Profile{
		"base": {
			ID:      "base",
			Version: "1.0.0",
		},
		"child1": {
			ID:           "child1",
			Version:      "1.0.0",
			InheritsFrom: "base",
		},
		"child2": {
			ID:        "child2",
			Version:   "1.0.0",
			DependsOn: []string{"child1"},
		},
	}

	// Set up dependencies
	loader.registry.Dependencies = map[string][]string{
		"base":   {},
		"child1": {"base"},
		"child2": {"child1"},
	}

	order, err := loader.topologicalSort(profiles)
	require.NoError(t, err)

	// Verify order: base should come before child1, child1 before child2
	baseIdx := findIndex(order, "base")
	child1Idx := findIndex(order, "child1")
	child2Idx := findIndex(order, "child2")

	assert.True(t, baseIdx < child1Idx, "base should come before child1")
	assert.True(t, child1Idx < child2Idx, "child1 should come before child2")
}

func TestTopologicalSortCircularDependency(t *testing.T) {
	logger := logrus.New()
	loader := NewLoader("", logger)

	// Create profiles with circular dependency
	profiles := map[string]*types.Profile{
		"a": {ID: "a", Version: "1.0.0"},
		"b": {ID: "b", Version: "1.0.0"},
	}

	// Set up circular dependencies
	loader.registry.Dependencies = map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}

	_, err := loader.topologicalSort(profiles)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency detected")
}

func TestMergeWithParent(t *testing.T) {
	logger := logrus.New()
	loader := NewLoader("", logger)

	parent := &types.Profile{
		System: "Parent system prompt",
		ModelParams: types.ModelParams{
			Temperature:    0.2,
			MaxTokens:      500,
			TimeoutSeconds: 20,
		},
		FewShot: []types.FewShotExample{
			{Name: "parent_example", Input: "parent input", Output: "parent output"},
		},
		Policy: types.PolicyConfig{
			Conditions: []types.PolicyCondition{
				{Name: "parent_condition", Expression: "confidence > 0.8"},
			},
		},
		Response: types.ResponseConfig{
			Validation: types.ValidationConfig{
				RequiredFields:  []string{"action", "confidence"},
				ConfidenceRange: [2]float64{0.0, 1.0},
			},
		},
	}

	child := &types.Profile{
		System: "Child system prompt",
		ModelParams: types.ModelParams{
			Temperature: 0.1, // Override parent
			// MaxTokens and TimeoutSeconds should inherit from parent
		},
		FewShot: []types.FewShotExample{
			{Name: "child_example", Input: "child input", Output: "child output"},
		},
		Policy: types.PolicyConfig{
			Conditions: []types.PolicyCondition{
				{Name: "child_condition", Expression: "confidence > 0.5"},
			},
		},
	}

	err := loader.mergeWithParent(child, parent)
	require.NoError(t, err)

	// Verify system prompt is merged
	expectedSystem := "Parent system prompt\n\nChild system prompt"
	assert.Equal(t, expectedSystem, child.System)

	// Verify model params are merged (child overrides, parent fills gaps)
	assert.Equal(t, 0.1, child.ModelParams.Temperature)     // Child override
	assert.Equal(t, 500, child.ModelParams.MaxTokens)       // Inherited from parent
	assert.Equal(t, 20, child.ModelParams.TimeoutSeconds)   // Inherited from parent

	// Verify few-shot examples are merged (parent first)
	assert.Len(t, child.FewShot, 2)
	assert.Equal(t, "parent_example", child.FewShot[0].Name)
	assert.Equal(t, "child_example", child.FewShot[1].Name)

	// Verify policy conditions are merged (parent first)
	assert.Len(t, child.Policy.Conditions, 2)
	assert.Equal(t, "parent_condition", child.Policy.Conditions[0].Name)
	assert.Equal(t, "child_condition", child.Policy.Conditions[1].Name)

	// Verify validation inherits from parent when child is empty
	assert.Equal(t, []string{"action", "confidence"}, child.Response.Validation.RequiredFields)
	assert.Equal(t, [2]float64{0.0, 1.0}, child.Response.Validation.ConfidenceRange)
}

func TestGetProfile(t *testing.T) {
	logger := logrus.New()
	loader := NewLoader("", logger)

	// Add test profile to registry
	testProfile := validTestProfile()
	loader.registry.Profiles = map[string]*types.Profile{
		"test": testProfile,
	}

	// Test getting existing profile
	profile, err := loader.GetProfile("test")
	require.NoError(t, err)
	assert.Equal(t, testProfile, profile)

	// Test getting non-existent profile
	_, err = loader.GetProfile("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "profile nonexistent not found")
}

func TestListProfiles(t *testing.T) {
	logger := logrus.New()
	loader := NewLoader("", logger)

	// Add test profiles to registry
	loader.registry.Profiles = map[string]*types.Profile{
		"spam":     {ID: "spam"},
		"meetings": {ID: "meetings"},
		"alerts":   {ID: "alerts"},
	}

	profiles := loader.ListProfiles()
	assert.Len(t, profiles, 3)
	assert.Contains(t, profiles, "spam")
	assert.Contains(t, profiles, "meetings")
	assert.Contains(t, profiles, "alerts")

	// Verify they're sorted
	assert.Equal(t, []string{"alerts", "meetings", "spam"}, profiles)
}

// Helper functions

func validTestProfile() *types.Profile {
	return &types.Profile{
		ID:      "test",
		Version: "1.0.0",
		Model:   "qwen2.5:7b",
		System:  "Test system prompt",
		ModelParams: types.ModelParams{
			Temperature:    0.1,
			MaxTokens:      1000,
			TimeoutSeconds: 30,
		},
		Response: types.ResponseConfig{
			Validation: types.ValidationConfig{
				RequiredFields:  []string{"action"},
				ConfidenceRange: [2]float64{0.0, 1.0},
			},
		},
	}
}

func findIndex(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}
