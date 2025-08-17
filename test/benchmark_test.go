package test

import (
	"context"
	"testing"
	"time"

	"github.com/mailsentinel/core/internal/ollama"
	"github.com/mailsentinel/core/internal/profile"
	"github.com/mailsentinel/core/pkg/config"
	"github.com/mailsentinel/core/pkg/testutil"
	"github.com/mailsentinel/core/pkg/types"
	"github.com/sirupsen/logrus"
)

// BenchmarkEmailClassification measures classification performance
func BenchmarkEmailClassification(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	testData := testutil.LoadTestData(&testing.T{})
	ollamaServer := testData.MockOllamaServer(&testing.T{})
	defer ollamaServer.Close()
	
	cfg := &config.OllamaConfig{
		BaseURL:      ollamaServer.URL,
		DefaultModel: "qwen2.5:7b",
		Timeout:      30 * time.Second,
		CircuitBreaker: config.CircuitBreakerConfig{
			MaxRequests: 100,
			Interval:    60 * time.Second,
			Timeout:     60 * time.Second,
			ReadyToTrip: 10,
		},
	}
	
	ollamaClient := ollama.NewClient(cfg, logrus.New())

	// Create test profile
	testProfile := &types.Profile{
		ID:    "benchmark",
		Model: "qwen2.5:7b",
		ModelParams: types.ModelParams{
			Temperature: 0.7,
			MaxTokens:   100,
		},
		Response: types.ResponseConfig{
			Schema: "json",
		},
	}
	
	email := testData.GetTestEmail("test-email-001")
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := ollamaClient.ClassifyEmail(context.Background(), testProfile, email)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBatchProcessing measures batch email processing performance
func BenchmarkBatchProcessing(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	testData := testutil.LoadTestData(&testing.T{})
	ollamaServer := testData.MockOllamaServer(&testing.T{})
	defer ollamaServer.Close()
	
	cfg := &config.OllamaConfig{
		BaseURL:      ollamaServer.URL,
		DefaultModel: "qwen2.5:7b",
		Timeout:      30 * time.Second,
		CircuitBreaker: config.CircuitBreakerConfig{
			MaxRequests: 1000,
			Interval:    60 * time.Second,
			Timeout:     60 * time.Second,
			ReadyToTrip: 50,
		},
	}
	
	ollamaClient := ollama.NewClient(cfg, logrus.New())
	testProfile := &types.Profile{
		ID:    "batch",
		Model: "qwen2.5:7b",
		ModelParams: types.ModelParams{
			Temperature: 0.5,
			MaxTokens:   50,
		},
		Response: types.ResponseConfig{},
	}
	
	// Create batch of test emails
	batchSize := 10
	emails := make([]*types.Email, batchSize)
	for i := 0; i < batchSize; i++ {
		emails[i] = testData.GetTestEmail("test-email-001")
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		for _, email := range emails {
			_, err := ollamaClient.ClassifyEmail(context.Background(), testProfile, email)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkProfileLoading measures profile loading performance
func BenchmarkProfileLoading(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	b.ResetTimer()
	
	loader := profile.NewLoader("./profiles", logger)
	err := loader.LoadAll()
	if err != nil {
		b.Fatal(err)
	}
}

// BenchmarkMemoryUsage measures memory usage during classification
func BenchmarkMemoryUsage(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	testData := testutil.LoadTestData(&testing.T{})
	ollamaServer := testData.MockOllamaServer(&testing.T{})
	defer ollamaServer.Close()
	
	cfg := &config.OllamaConfig{
		BaseURL:      ollamaServer.URL,
		DefaultModel: "qwen2.5:7b",
		Timeout:      30 * time.Second,
	}
	
	client := ollama.NewClient(cfg, logger)
	
	// Create large email for memory testing
	largeEmail := &types.Email{
		ID:      "large-email",
		Subject: "Large email for memory testing",
		From:    "test@example.com",
		To:      []string{"recipient@example.com"},
		Body:    generateLargeEmailBody(10000), // 10KB body
	}
	
	profile := &types.Profile{
		ID:     "memory-test",
		Model:  "qwen2.5:7b",
		System: "Process large emails efficiently.",
		ModelParams: types.ModelParams{
			Temperature: 0.7,
			MaxTokens:   100,
		},
		Response: types.ResponseConfig{Schema: "json"},
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := client.ClassifyEmail(context.Background(), profile, largeEmail)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func generateLargeEmailBody(size int) string {
	content := "This is a test email body with repeated content. "
	result := ""
	for len(result) < size {
		result += content
	}
	return result[:size]
}
