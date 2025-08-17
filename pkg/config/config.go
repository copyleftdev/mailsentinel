package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main application configuration
type Config struct {
	Gmail    GmailConfig    `yaml:"gmail" json:"gmail"`
	Ollama   OllamaConfig   `yaml:"ollama" json:"ollama"`
	Profiles ProfilesConfig `yaml:"profiles" json:"profiles"`
	Audit    AuditConfig    `yaml:"audit" json:"audit"`
	Security SecurityConfig `yaml:"security" json:"security"`
	Server   ServerConfig   `yaml:"server" json:"server"`
}

// GmailConfig contains Gmail API configuration
type GmailConfig struct {
	ClientID       string        `yaml:"client_id" json:"client_id"`
	ClientSecret   string        `yaml:"client_secret" json:"client_secret"`
	TokenFile      string        `yaml:"token_file" json:"token_file"`
	Scopes         []string      `yaml:"scopes" json:"scopes"`
	BatchSize      int           `yaml:"batch_size" json:"batch_size"`
	RateLimit      int           `yaml:"rate_limit" json:"rate_limit"`
	Timeout        time.Duration `yaml:"timeout" json:"timeout"`
	RetryAttempts  int           `yaml:"retry_attempts" json:"retry_attempts"`
	RetryDelay     time.Duration `yaml:"retry_delay" json:"retry_delay"`
}

// OllamaConfig contains Ollama client configuration
type OllamaConfig struct {
	BaseURL           string        `yaml:"base_url" json:"base_url"`
	DefaultModel      string        `yaml:"default_model" json:"default_model"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout"`
	MaxRetries        int           `yaml:"max_retries" json:"max_retries"`
	CircuitBreaker    CircuitBreakerConfig `yaml:"circuit_breaker" json:"circuit_breaker"`
	RequestTimeout    time.Duration `yaml:"request_timeout" json:"request_timeout"`
	HealthCheckPeriod time.Duration `yaml:"health_check_period" json:"health_check_period"`
}

// CircuitBreakerConfig defines circuit breaker parameters
type CircuitBreakerConfig struct {
	MaxRequests     uint32        `yaml:"max_requests" json:"max_requests"`
	Interval        time.Duration `yaml:"interval" json:"interval"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	ReadyToTrip     int           `yaml:"ready_to_trip" json:"ready_to_trip"`
}

// ProfilesConfig contains profile system configuration
type ProfilesConfig struct {
	Directory       string        `yaml:"directory" json:"directory"`
	ResolverConfig  string        `yaml:"resolver_config" json:"resolver_config"`
	ReloadInterval  time.Duration `yaml:"reload_interval" json:"reload_interval"`
	ValidateOnLoad  bool          `yaml:"validate_on_load" json:"validate_on_load"`
	CacheEnabled    bool          `yaml:"cache_enabled" json:"cache_enabled"`
}

// AuditConfig contains audit logging configuration
type AuditConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	Directory       string        `yaml:"directory" json:"directory"`
	MaxFileSize     int64         `yaml:"max_file_size" json:"max_file_size"`
	MaxFiles        int           `yaml:"max_files" json:"max_files"`
	RotationPeriod  time.Duration `yaml:"rotation_period" json:"rotation_period"`
	IntegrityCheck  bool          `yaml:"integrity_check" json:"integrity_check"`
	EncryptionKey   string        `yaml:"encryption_key" json:"encryption_key"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	EncryptionKey     string `yaml:"encryption_key" json:"encryption_key"`
	TokenEncryption   bool   `yaml:"token_encryption" json:"token_encryption"`
	InputSanitization bool   `yaml:"input_sanitization" json:"input_sanitization"`
	MaxEmailSize      int64  `yaml:"max_email_size" json:"max_email_size"`
	MaxBatchSize      int    `yaml:"max_batch_size" json:"max_batch_size"`
}

// ServerConfig contains server configuration
type ServerConfig struct {
	Port            int           `yaml:"port" json:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout"`
	MaxHeaderBytes  int           `yaml:"max_header_bytes" json:"max_header_bytes"`
	EnableProfiling bool          `yaml:"enable_profiling" json:"enable_profiling"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Gmail: GmailConfig{
			Scopes:        []string{"https://www.googleapis.com/auth/gmail.readonly", "https://www.googleapis.com/auth/gmail.modify"},
			BatchSize:     100,
			RateLimit:     250,
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
			RetryDelay:    1 * time.Second,
			TokenFile:     "data/gmail_token.json",
		},
		Ollama: OllamaConfig{
			BaseURL:           "http://127.0.0.1:11434",
			DefaultModel:      "qwen2.5:7b",
			Timeout:           30 * time.Second,
			MaxRetries:        3,
			RequestTimeout:    30 * time.Second,
			HealthCheckPeriod: 60 * time.Second,
			CircuitBreaker: CircuitBreakerConfig{
				MaxRequests:  10,
				Interval:     60 * time.Second,
				Timeout:      60 * time.Second,
				ReadyToTrip:  5,
			},
		},
		Profiles: ProfilesConfig{
			Directory:       "profiles",
			ResolverConfig:  "profiles/resolver.yaml",
			ReloadInterval:  5 * time.Minute,
			ValidateOnLoad:  true,
			CacheEnabled:    true,
		},
		Audit: AuditConfig{
			Enabled:         true,
			Directory:       "data/audit",
			MaxFileSize:     100 * 1024 * 1024, // 100MB
			MaxFiles:        10,
			RotationPeriod:  24 * time.Hour,
			IntegrityCheck:  true,
		},
		Security: SecurityConfig{
			TokenEncryption:   true,
			InputSanitization: true,
			MaxEmailSize:      10 * 1024 * 1024, // 10MB
			MaxBatchSize:      1000,
		},
		Server: ServerConfig{
			Port:            8080,
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			MaxHeaderBytes:  1 << 20, // 1MB
			EnableProfiling: false,
		},
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()
	
	if path == "" {
		return config, nil
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Use defaults if file doesn't exist
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Expand environment variables in the YAML content
	expandedData := os.ExpandEnv(string(data))
	
	if err := yaml.Unmarshal([]byte(expandedData), config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return config, nil
}

// SaveConfig saves configuration to a YAML file
func (c *Config) SaveConfig(path string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Gmail.ClientID == "" {
		return fmt.Errorf("gmail.client_id is required")
	}
	
	if c.Gmail.ClientSecret == "" {
		return fmt.Errorf("gmail.client_secret is required")
	}
	
	if c.Ollama.BaseURL == "" {
		return fmt.Errorf("ollama.base_url is required")
	}
	
	if c.Ollama.DefaultModel == "" {
		return fmt.Errorf("ollama.default_model is required")
	}
	
	if c.Profiles.Directory == "" {
		return fmt.Errorf("profiles.directory is required")
	}
	
	return nil
}
