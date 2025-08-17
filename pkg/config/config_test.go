package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test Gmail defaults
	assert.Equal(t, []string{
		"https://www.googleapis.com/auth/gmail.readonly",
		"https://www.googleapis.com/auth/gmail.modify",
	}, cfg.Gmail.Scopes)
	assert.Equal(t, 100, cfg.Gmail.BatchSize)
	assert.Equal(t, 250, cfg.Gmail.RateLimit)
	assert.Equal(t, 30*time.Second, cfg.Gmail.Timeout)

	// Test Ollama defaults
	assert.Equal(t, "http://127.0.0.1:11434", cfg.Ollama.BaseURL)
	assert.Equal(t, "qwen2.5:7b", cfg.Ollama.DefaultModel)
	assert.Equal(t, 30*time.Second, cfg.Ollama.Timeout)
	assert.Equal(t, 5, cfg.Ollama.CircuitBreaker.ReadyToTrip)

	// Test security defaults
	assert.True(t, cfg.Security.TokenEncryption)
	assert.True(t, cfg.Security.InputSanitization)
	assert.Equal(t, int64(10*1024*1024), cfg.Security.MaxEmailSize)

	// Test audit defaults
	assert.True(t, cfg.Audit.Enabled)
	assert.True(t, cfg.Audit.IntegrityCheck)
	assert.Equal(t, int64(100*1024*1024), cfg.Audit.MaxFileSize)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:   "valid_config",
			config: validTestConfig(),
			wantErr: false,
		},
		{
			name: "missing_gmail_client_id",
			config: func() *Config {
				cfg := validTestConfig()
				cfg.Gmail.ClientID = ""
				return cfg
			}(),
			wantErr: true,
			errMsg:  "gmail.client_id is required",
		},
		{
			name: "missing_gmail_client_secret",
			config: func() *Config {
				cfg := validTestConfig()
				cfg.Gmail.ClientSecret = ""
				return cfg
			}(),
			wantErr: true,
			errMsg:  "gmail.client_secret is required",
		},
		{
			name: "missing_ollama_base_url",
			config: func() *Config {
				cfg := validTestConfig()
				cfg.Ollama.BaseURL = ""
				return cfg
			}(),
			wantErr: true,
			errMsg:  "ollama.base_url is required",
		},
		{
			name: "missing_ollama_model",
			config: func() *Config {
				cfg := validTestConfig()
				cfg.Ollama.DefaultModel = ""
				return cfg
			}(),
			wantErr: true,
			errMsg:  "ollama.default_model is required",
		},
		{
			name: "missing_profiles_directory",
			config: func() *Config {
				cfg := validTestConfig()
				cfg.Profiles.Directory = ""
				return cfg
			}(),
			wantErr: true,
			errMsg:  "profiles.directory is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Test loading non-existent file (should return defaults)
	cfg, err := LoadConfig("non-existent.yaml")
	require.NoError(t, err)
	assert.Equal(t, "qwen2.5:7b", cfg.Ollama.DefaultModel)

	// Test loading valid config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	
	configContent := `
gmail:
  client_id: "test_client_id"
  client_secret: "test_client_secret"
  batch_size: 50
ollama:
  base_url: "http://localhost:11434"
  default_model: "test_model"
profiles:
  directory: "test_profiles"
`
	
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err = LoadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, "test_client_id", cfg.Gmail.ClientID)
	assert.Equal(t, "test_client_secret", cfg.Gmail.ClientSecret)
	assert.Equal(t, 50, cfg.Gmail.BatchSize)
	assert.Equal(t, "test_model", cfg.Ollama.DefaultModel)
	assert.Equal(t, "test_profiles", cfg.Profiles.Directory)
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "save-test.yaml")

	cfg := validTestConfig()
	cfg.Gmail.BatchSize = 200
	cfg.Ollama.DefaultModel = "custom_model"

	err := cfg.SaveConfig(configPath)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Load and verify content
	loadedCfg, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, 200, loadedCfg.Gmail.BatchSize)
	assert.Equal(t, "custom_model", loadedCfg.Ollama.DefaultModel)
}

func TestCircuitBreakerConfig(t *testing.T) {
	cfg := DefaultConfig()
	cb := cfg.Ollama.CircuitBreaker

	assert.Equal(t, uint32(10), cb.MaxRequests)
	assert.Equal(t, 60*time.Second, cb.Interval)
	assert.Equal(t, 60*time.Second, cb.Timeout)
	assert.Equal(t, 5, cb.ReadyToTrip)
}

// validTestConfig returns a valid configuration for testing
func validTestConfig() *Config {
	cfg := DefaultConfig()
	cfg.Gmail.ClientID = "test_client_id"
	cfg.Gmail.ClientSecret = "test_client_secret"
	cfg.Ollama.BaseURL = "http://127.0.0.1:11434"
	cfg.Ollama.DefaultModel = "qwen2.5:7b"
	cfg.Profiles.Directory = "profiles"
	return cfg
}
