package test

import (
	"testing"
	"time"

	"github.com/mailsentinel/core/pkg/config"
	"github.com/mailsentinel/core/pkg/testutil"
	"github.com/mailsentinel/core/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTestDataLoading verifies test data can be loaded correctly
func TestTestDataLoading(t *testing.T) {
	testData := testutil.LoadTestData(t)
	
	// Verify emails loaded
	assert.NotEmpty(t, testData.Emails)
	
	// Verify specific test email exists
	email := testData.GetTestEmail("test-email-001")
	require.NotNil(t, email)
	assert.Equal(t, "test-email-001", email.ID)
	assert.NotEmpty(t, email.Subject)
	assert.NotEmpty(t, email.From)
	assert.NotEmpty(t, email.Body)
}

// TestMockServers verifies mock servers start and respond correctly
func TestMockServers(t *testing.T) {
	testData := testutil.LoadTestData(t)
	
	// Test Ollama mock server
	ollamaServer := testData.MockOllamaServer(t)
	defer ollamaServer.Close()
	
	assert.NotEmpty(t, ollamaServer.URL)
	
	// Test Gmail mock server
	gmailServer := testData.MockGmailServer(t)
	defer gmailServer.Close()
	
	assert.NotEmpty(t, gmailServer.URL)
}

// TestConfigDefaults verifies configuration defaults work
func TestConfigDefaults(t *testing.T) {
	cfg := config.DefaultConfig()
	
	require.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.Gmail.Scopes)
	assert.Greater(t, cfg.Gmail.BatchSize, 0)
	assert.Greater(t, cfg.Ollama.Timeout, time.Duration(0))
}

// TestEmailStructure verifies email structure matches expectations
func TestEmailStructure(t *testing.T) {
	email := &types.Email{
		ID:      "test-123",
		Subject: "Test Email",
		From:    "test@example.com",
		To:      []string{"recipient@example.com"},
		Body:    "Test email body",
		Date:    time.Now(),
		Labels:  []string{"INBOX"},
		Headers: map[string]string{"Message-ID": "test-123"},
	}
	
	assert.Equal(t, "test-123", email.ID)
	assert.Equal(t, "Test Email", email.Subject)
	assert.Len(t, email.To, 1)
	assert.Contains(t, email.Labels, "INBOX")
}

// TestClassificationResponse verifies classification response structure
func TestClassificationResponse(t *testing.T) {
	response := &types.ClassificationResponse{
		ProfileID:   "test-profile",
		Action:      "archive",
		Confidence:  0.85,
		Reasoning:   "Test reasoning",
		ProcessedAt: time.Now(),
	}
	
	assert.Equal(t, "test-profile", response.ProfileID)
	assert.Equal(t, "archive", response.Action)
	assert.Greater(t, response.Confidence, 0.0)
	assert.LessOrEqual(t, response.Confidence, 1.0)
	assert.NotEmpty(t, response.Reasoning)
}
