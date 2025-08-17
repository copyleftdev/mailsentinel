package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mailsentinel/core/pkg/types"
	"github.com/stretchr/testify/require"
)

// TestData holds all test fixtures and golden files
type TestData struct {
	Emails              []types.Email
	GmailResponses      map[string]interface{}
	OllamaResponses     map[string]interface{}
	ClassificationGold  map[string]interface{}
	PolicyResolutionGold map[string]interface{}
	AuditLogs           []interface{}
}

// LoadTestData loads all test fixtures from the testdata directory
func LoadTestData(t *testing.T) *TestData {
	testDataDir := getTestDataDir(t)
	
	data := &TestData{}
	
	// Load email fixtures
	data.Emails = loadJSONFile[[]types.Email](t, filepath.Join(testDataDir, "fixtures", "emails.json"))
	
	// Load API response mocks
	data.GmailResponses = loadJSONFile[map[string]interface{}](t, filepath.Join(testDataDir, "fixtures", "gmail_responses.json"))
	data.OllamaResponses = loadJSONFile[map[string]interface{}](t, filepath.Join(testDataDir, "fixtures", "ollama_responses.json"))
	
	// Load golden files
	data.ClassificationGold = loadJSONFile[map[string]interface{}](t, filepath.Join(testDataDir, "golden", "classification_outputs.json"))
	data.PolicyResolutionGold = loadJSONFile[map[string]interface{}](t, filepath.Join(testDataDir, "golden", "policy_resolutions.json"))
	
	// Load audit logs
	data.AuditLogs = loadJSONFile[[]interface{}](t, filepath.Join(testDataDir, "fixtures", "audit_logs.json"))
	
	return data
}

// GetTestEmail returns a test email by ID
func (td *TestData) GetTestEmail(id string) *types.Email {
	for _, email := range td.Emails {
		if email.ID == id {
			return &email
		}
	}
	return nil
}

// GetExpectedClassification returns expected classification for an email
func (td *TestData) GetExpectedClassification(emailID string) map[string]interface{} {
	for _, value := range td.ClassificationGold {
		if data, ok := value.(map[string]interface{}); ok {
			if input, exists := data["input"].(map[string]interface{}); exists {
				if input["email_id"] == emailID {
					return data["expected_output"].(map[string]interface{})
				}
			}
		}
	}
	return nil
}

// MockOllamaServer creates a mock Ollama server with test responses
func (td *TestData) MockOllamaServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch r.URL.Path {
		case "/api/tags":
			response := td.OllamaResponses["models_list_response"]
			json.NewEncoder(w).Encode(response)
		case "/api/generate":
			// Parse request to determine which response to send
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			
			prompt := req["prompt"].(string)
			var response interface{}
			
			// Simple pattern matching to return appropriate response
			if containsAny(prompt, []string{"amaz0n", "phishing", "suspicious"}) {
				response = td.OllamaResponses["classification_responses"].(map[string]interface{})["phishing_email"]
			} else if containsAny(prompt, []string{"$5000", "get-rich", "opportunity"}) {
				response = td.OllamaResponses["classification_responses"].(map[string]interface{})["spam_email"]
			} else if containsAny(prompt, []string{"meeting", "colleague", "conference"}) {
				response = td.OllamaResponses["classification_responses"].(map[string]interface{})["legitimate_email"]
			} else {
				response = td.OllamaResponses["classification_responses"].(map[string]interface{})["legitimate_email"]
			}
			
			json.NewEncoder(w).Encode(response)
		case "/":
			response := td.OllamaResponses["health_check_response"]
			json.NewEncoder(w).Encode(response)
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "endpoint not found"})
		}
	}))
}

// MockGmailServer creates a mock Gmail API server
func (td *TestData) MockGmailServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch {
		case r.URL.Path == "/gmail/v1/users/me/messages":
			response := td.GmailResponses["messages_list_response"]
			json.NewEncoder(w).Encode(response)
		case r.URL.Path == "/gmail/v1/users/me/messages/test-email-001":
			response := td.GmailResponses["message_get_response"]
			json.NewEncoder(w).Encode(response)
		case r.URL.Path == "/gmail/v1/users/me/labels":
			response := td.GmailResponses["labels_list_response"]
			json.NewEncoder(w).Encode(response)
		case r.URL.Path == "/gmail/v1/users/me/profile":
			response := td.GmailResponses["profile_response"]
			json.NewEncoder(w).Encode(response)
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "endpoint not found"})
		}
	}))
}

// AssertClassificationResult validates a classification result against golden data
func (td *TestData) AssertClassificationResult(t *testing.T, emailID string, actual *types.ClassificationResponse) {
	expected := td.GetExpectedClassification(emailID)
	require.NotNil(t, expected, "No expected classification found for email %s", emailID)
	
	require.Equal(t, expected["action"], actual.Action, "Action mismatch for email %s", emailID)
	require.InDelta(t, expected["confidence"], actual.Confidence, 0.05, "Confidence mismatch for email %s", emailID)
	require.NotEmpty(t, actual.Reasoning, "Reasoning should not be empty for email %s", emailID)
}

// CreateTempConfig creates a temporary configuration file for testing
func CreateTempConfig(t *testing.T, configType string) string {
	testDataDir := getTestDataDir(t)
	configPath := filepath.Join(testDataDir, "mocks", "config_templates.yaml")
	
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	
	tempFile, err := os.CreateTemp("", "mailsentinel-test-config-*.yaml")
	require.NoError(t, err)
	
	_, err = tempFile.Write(data)
	require.NoError(t, err)
	tempFile.Close()
	
	return tempFile.Name()
}

// CleanupTempFiles removes temporary test files
func CleanupTempFiles(t *testing.T, files ...string) {
	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to cleanup temp file %s: %v", file, err)
		}
	}
}

// Helper functions

func getTestDataDir(t *testing.T) string {
	// Try to find testdata directory relative to current working directory
	candidates := []string{
		"testdata",
		"../testdata",
		"../../testdata",
		"../../../testdata",
	}
	
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(candidate)
			return abs
		}
	}
	
	t.Fatal("Could not find testdata directory")
	return ""
}

func loadJSONFile[T any](t *testing.T, path string) T {
	var result T
	
	data, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read file %s", path)
	
	err = json.Unmarshal(data, &result)
	require.NoError(t, err, "Failed to parse JSON from %s", path)
	
	return result
}

func containsAny(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if len(text) >= len(pattern) {
			for i := 0; i <= len(text)-len(pattern); i++ {
				if text[i:i+len(pattern)] == pattern {
					return true
				}
			}
		}
	}
	return false
}
