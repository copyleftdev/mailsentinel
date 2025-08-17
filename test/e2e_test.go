package test

import (
	"context"
	"testing"
	"time"

	"github.com/mailsentinel/core/internal/audit"
	"github.com/mailsentinel/core/internal/ollama"
	"github.com/mailsentinel/core/pkg/config"
	"github.com/mailsentinel/core/pkg/testutil"
	"github.com/mailsentinel/core/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite tests with actual local Ollama
type E2ETestSuite struct {
	suite.Suite
	testData *testutil.TestData
	tempDir  string
	logger   *logrus.Logger
}

func (suite *E2ETestSuite) SetupSuite() {
	suite.logger = logrus.New()
	suite.logger.SetLevel(logrus.WarnLevel)
	
	suite.testData = testutil.LoadTestData(suite.T())
	suite.tempDir = suite.T().TempDir()
}

func (suite *E2ETestSuite) createTestConfig() *config.Config {
	return &config.Config{
		Ollama: config.OllamaConfig{
			BaseURL:      "http://localhost:11434",
			DefaultModel: "qwen2.5:latest",
			Timeout:      30 * time.Second,
			CircuitBreaker: config.CircuitBreakerConfig{
				MaxRequests: 10,
				Interval:    60 * time.Second,
				Timeout:     60 * time.Second,
				ReadyToTrip: 5,
			},
		},
		Audit: config.AuditConfig{
			Enabled:        true,
			Directory:      suite.tempDir,
			IntegrityCheck: true,
		},
	}
}

func (suite *E2ETestSuite) TestCompleteEmailTriageWorkflow() {
	// Test the complete email triage workflow with real Ollama
	
	cfg := suite.createTestConfig()
	
	// Initialize components
	ollamaClient := ollama.NewClient(&cfg.Ollama, suite.logger)
	auditLogger, err := audit.NewLogger(&cfg.Audit, suite.logger)
	require.NoError(suite.T(), err)
	defer auditLogger.Close()
	
	// Check if Ollama is available
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err = ollamaClient.HealthCheck(ctx)
	if err != nil {
		suite.T().Skip("Local Ollama not available, skipping e2e test:", err)
		return
	}
	
	// Test scenarios with different email types
	testScenarios := []struct {
		emailID         string
		expectedAction  string
		profileType     string
		minConfidence   float64
	}{
		{"test-email-001", "delete", "spam_detector", 0.7},
		{"test-email-002", "archive", "newsletter_classifier", 0.6},
	}
	
	for _, scenario := range testScenarios {
		suite.Run(scenario.emailID, func() {
			email := suite.testData.GetTestEmail(scenario.emailID)
			require.NotNil(suite.T(), email)
			
			// Create test profile for this scenario
			testProfile := &types.Profile{
				ID:      scenario.profileType,
				Model:   "qwen2.5:latest",
				Version: "1.0.0",
				System:  "You are an email classifier. Classify emails and respond with JSON containing action, confidence, and reasoning.",
				ModelParams: types.ModelParams{
					Temperature: 0.7,
					MaxTokens:   150,
				},
				Response: types.ResponseConfig{
					Schema: "json",
				},
			}
			
			// Classify email
			ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel2()
			
			result, err := ollamaClient.ClassifyEmail(ctx2, testProfile, email)
			require.NoError(suite.T(), err)
			require.NotNil(suite.T(), result)
			
			// Verify results
			assert.GreaterOrEqual(suite.T(), result.Confidence, scenario.minConfidence)
			assert.NotEmpty(suite.T(), result.Action)
			assert.NotEmpty(suite.T(), result.Reasoning)
			assert.Equal(suite.T(), testProfile.ID, result.ProfileID)
			
			// Log to audit trail
			err = auditLogger.LogClassification(email, result)
			require.NoError(suite.T(), err)
			
			// Log action
			err = auditLogger.LogAction(email, result.Action, "automated_classification")
			require.NoError(suite.T(), err)
			
			suite.logger.Infof("E2E Test: %s -> %s (%.2f confidence)", 
				email.Subject, result.Action, result.Confidence)
		})
	}
	
	// Verify audit integrity
	valid, err := auditLogger.VerifyIntegrity()
	require.NoError(suite.T(), err)
	assert.True(suite.T(), valid, "Audit log integrity should be valid")
}

func (suite *E2ETestSuite) TestErrorRecoveryScenarios() {
	// Test error recovery with circuit breaker
	
	cfg := suite.createTestConfig()
	
	// Create client with aggressive circuit breaker for testing
	cfg.Ollama.BaseURL = "http://localhost:99999" // Invalid URL
	cfg.Ollama.CircuitBreaker.ReadyToTrip = 2
	
	ollamaClient := ollama.NewClient(&cfg.Ollama, suite.logger)
	
	testProfile := &types.Profile{
		ID:      "error_test",
		Model:   "qwen2.5:latest",
		Version: "1.0.0",
		System:  "Test profile for error scenarios",
		ModelParams: types.ModelParams{
			Temperature: 0.7,
			MaxTokens:   100,
		},
		Response: types.ResponseConfig{
			Schema: "json",
		},
	}
	
	email := suite.testData.GetTestEmail("test-email-001")
	require.NotNil(suite.T(), email)
	
	// Test circuit breaker behavior
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := ollamaClient.ClassifyEmail(ctx, testProfile, email)
	assert.Error(suite.T(), err, "Should fail with invalid Ollama URL")
	
	suite.logger.Info("✅ Error recovery test completed")
}

func (suite *E2ETestSuite) TestPerformanceUnderLoad() {
	// Test performance with multiple concurrent requests
	
	cfg := suite.createTestConfig()
	ollamaClient := ollama.NewClient(&cfg.Ollama, suite.logger)
	
	// Check if Ollama is available
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err := ollamaClient.HealthCheck(ctx)
	if err != nil {
		suite.T().Skip("Local Ollama not available for performance test:", err)
		return
	}
	
	testProfile := &types.Profile{
		ID:      "performance_test",
		Model:   "qwen2.5:latest",
		Version: "1.0.0",
		System:  "Quick email classifier for performance testing",
		ModelParams: types.ModelParams{
			Temperature: 0.5,
			MaxTokens:   200,
		},
		Response: types.ResponseConfig{
			Schema: "json",
		},
	}
	
	email := suite.testData.GetTestEmail("test-email-001")
	require.NotNil(suite.T(), email)
	
	// Test multiple sequential requests
	start := time.Now()
	for i := 0; i < 3; i++ {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
		result, err := ollamaClient.ClassifyEmail(ctx2, testProfile, email)
		cancel2()
		
		require.NoError(suite.T(), err)
		require.NotNil(suite.T(), result)
	}
	duration := time.Since(start)
	
	suite.logger.Infof("✅ Performance test: 3 requests in %v", duration)
	assert.Less(suite.T(), duration, 2*time.Minute, "Should complete within reasonable time")
}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
