package test

import (
	"context"
	"testing"
	"time"

	"github.com/mailsentinel/core/internal/ollama"
	"github.com/mailsentinel/core/pkg/config"
	"github.com/mailsentinel/core/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalOllamaConnection(t *testing.T) {
	// Test basic connection to local Ollama
	cfg := &config.OllamaConfig{
		BaseURL:      "http://localhost:11434",
		DefaultModel: "qwen2.5:latest",
		Timeout:      30 * time.Second,
		CircuitBreaker: config.CircuitBreakerConfig{
			MaxRequests: 10,
			Interval:    60 * time.Second,
			Timeout:     60 * time.Second,
			ReadyToTrip: 5,
		},
	}
	
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	
	client := ollama.NewClient(cfg, logger)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err := client.HealthCheck(ctx)
	if err != nil {
		t.Skip("Local Ollama not available:", err)
		return
	}
	
	assert.NoError(t, err)
	t.Log("✅ Local Ollama connection successful")
}

func TestLocalOllamaClassification(t *testing.T) {
	// Test email classification with local Ollama
	cfg := &config.OllamaConfig{
		BaseURL:      "http://localhost:11434",
		DefaultModel: "qwen2.5:latest",
		Timeout:      30 * time.Second,
		CircuitBreaker: config.CircuitBreakerConfig{
			MaxRequests: 10,
			Interval:    60 * time.Second,
			Timeout:     60 * time.Second,
			ReadyToTrip: 5,
		},
	}
	
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	
	client := ollama.NewClient(cfg, logger)
	
	// Check if Ollama is available
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err := client.HealthCheck(ctx)
	if err != nil {
		t.Skip("Local Ollama not available:", err)
		return
	}
	
	// Create test profile
	profile := &types.Profile{
		ID:      "test_spam",
		Model:   "qwen2.5:latest",
		Version: "1.0.0",
		System:  "You are an email classifier. Classify emails as spam, legitimate, or phishing. Respond with JSON only.",
		ModelParams: types.ModelParams{
			Temperature: 0.7,
			MaxTokens:   150,
		},
		Response: types.ResponseConfig{
			Schema: "json",
		},
	}
	
	// Create test email
	email := &types.Email{
		ID:      "test-001",
		Subject: "URGENT: Claim your $1000 prize NOW!",
		From:    "winner@suspicious-site.com",
		Body:    "Congratulations! You have won $1000! Click this link immediately to claim your prize before it expires in 24 hours!",
		Date:    time.Now(),
	}
	
	// Perform classification
	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()
	
	result, err := client.ClassifyEmail(ctx2, profile, email)
	require.NoError(t, err)
	require.NotNil(t, result)
	
	// Verify result
	assert.NotEmpty(t, result.Action)
	assert.GreaterOrEqual(t, result.Confidence, 0.0)
	assert.LessOrEqual(t, result.Confidence, 1.0)
	assert.NotEmpty(t, result.Reasoning)
	assert.Equal(t, profile.ID, result.ProfileID)
	
	t.Logf("✅ Classification successful: action=%s, confidence=%.2f, reasoning=%s", 
		result.Action, result.Confidence, result.Reasoning)
}
