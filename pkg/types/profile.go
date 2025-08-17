package types

import (
	"time"
)

// Profile represents an email classification profile
type Profile struct {
	ID                    string                 `yaml:"id" json:"id"`
	Version               string                 `yaml:"version" json:"version"`
	InheritsFrom          string                 `yaml:"inherits_from,omitempty" json:"inherits_from,omitempty"`
	DependsOn             []string               `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	ConditionalExecution  *ConditionalExecution  `yaml:"conditional_execution,omitempty" json:"conditional_execution,omitempty"`
	Model                 string                 `yaml:"model" json:"model"`
	ModelParams           ModelParams            `yaml:"model_params" json:"model_params"`
	Response              ResponseConfig         `yaml:"response" json:"response"`
	System                string                 `yaml:"system" json:"system"`
	FewShot               []FewShotExample       `yaml:"fewshot" json:"fewshot"`
	Policy                PolicyConfig           `yaml:"policy" json:"policy"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

// ConditionalExecution defines when a profile should be executed
type ConditionalExecution struct {
	When   string `yaml:"when" json:"when"`
	Reason string `yaml:"reason" json:"reason"`
}

// ModelParams defines parameters for the LLM model
type ModelParams struct {
	Temperature    float64 `yaml:"temperature" json:"temperature"`
	MaxTokens      int     `yaml:"max_tokens" json:"max_tokens"`
	TimeoutSeconds int     `yaml:"timeout_seconds" json:"timeout_seconds"`
	TopP           float64 `yaml:"top_p,omitempty" json:"top_p,omitempty"`
	TopK           int     `yaml:"top_k,omitempty" json:"top_k,omitempty"`
}

// ResponseConfig defines the expected response format and validation
type ResponseConfig struct {
	Schema     string             `yaml:"schema" json:"schema"`
	Validation ValidationConfig   `yaml:"validation" json:"validation"`
}

// ValidationConfig defines validation rules for responses
type ValidationConfig struct {
	RequiredFields   []string  `yaml:"required_fields" json:"required_fields"`
	ConfidenceRange  [2]float64 `yaml:"confidence_range" json:"confidence_range"`
	AllowedActions   []string  `yaml:"allowed_actions,omitempty" json:"allowed_actions,omitempty"`
}

// FewShotExample represents a training example for the model
type FewShotExample struct {
	Name   string `yaml:"name" json:"name"`
	Input  string `yaml:"input" json:"input"`
	Output string `yaml:"output" json:"output"`
}

// PolicyConfig defines the decision-making policy
type PolicyConfig struct {
	Conditions []PolicyCondition `yaml:"conditions" json:"conditions"`
}

// PolicyCondition represents a single policy rule
type PolicyCondition struct {
	Name       string   `yaml:"name" json:"name"`
	Expression string   `yaml:"expression" json:"expression"`
	Actions    []string `yaml:"actions" json:"actions"`
	Priority   int      `yaml:"priority" json:"priority"`
}

// ProfileRegistry manages profile loading and dependencies
type ProfileRegistry struct {
	Profiles     map[string]*Profile `json:"profiles"`
	Dependencies map[string][]string `json:"dependencies"`
	LoadOrder    []string            `json:"load_order"`
}

// ResolverConfig defines how conflicts between profiles are resolved
type ResolverConfig struct {
	Version             string                    `yaml:"version" json:"version"`
	PriorityRules       []PriorityRule           `yaml:"priority_rules" json:"priority_rules"`
	ConfidenceWeighting ConfidenceWeighting      `yaml:"confidence_weighting" json:"confidence_weighting"`
	ConflictResolution  map[string]string        `yaml:"conflict_resolution" json:"conflict_resolution"`
}

// PriorityRule defines high-priority override conditions
type PriorityRule struct {
	Name            string  `yaml:"name" json:"name"`
	Condition       string  `yaml:"condition" json:"condition"`
	Action          string  `yaml:"action,omitempty" json:"action,omitempty"`
	ConfidenceBoost float64 `yaml:"confidence_boost,omitempty" json:"confidence_boost,omitempty"`
	Priority        int     `yaml:"priority" json:"priority"`
	Reason          string  `yaml:"reason" json:"reason"`
}

// ConfidenceWeighting defines how to weight different profile results
type ConfidenceWeighting struct {
	Method         string             `yaml:"method" json:"method"`
	ProfileWeights map[string]float64 `yaml:"profile_weights" json:"profile_weights"`
}
