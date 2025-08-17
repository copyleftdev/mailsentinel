# MailSentinel Test Data

This directory contains comprehensive test fixtures, golden files, and mock data for testing and development of the MailSentinel email triage system.

## Directory Structure

```
testdata/
├── fixtures/           # Test input data and configurations
├── golden/            # Expected outputs for validation
├── mocks/             # Mock responses and configurations
└── README.md          # This file
```

## Fixtures

### `fixtures/emails.json`
Sample email data covering various classification scenarios:
- **Phishing emails** - Suspicious domains, urgent language, malicious links
- **Spam emails** - Get-rich-quick schemes, promotional content
- **Legitimate emails** - Business communications, meeting requests
- **Important emails** - Client communications, project proposals
- **Newsletter emails** - Subscription content with unsubscribe links

### `fixtures/gmail_responses.json`
Mock Gmail API responses including:
- Message list responses with pagination
- Individual message details with headers and body
- Label management responses
- User profile information

### `fixtures/ollama_responses.json`
Mock Ollama API responses for:
- Model listing and availability
- Classification responses for different email types
- Health check responses
- Error scenarios (model not found, invalid requests)

### `fixtures/profiles.yaml`
Test profile configurations:
- **spam_basic** - Simple spam detection
- **phishing_advanced** - Advanced threat detection with metadata
- **newsletter_classifier** - Newsletter and promotional content classification

### `fixtures/audit_logs.json`
Sample audit log entries showing:
- Email classification events
- System startup events
- Policy conflict resolutions
- Cryptographic hashes and signatures

## Golden Files

### `golden/classification_outputs.json`
Expected classification results for test emails, including:
- Action recommendations (delete, archive, keep, prioritize)
- Confidence scores
- Reasoning explanations
- Metadata scores (spam_score, phishing_score)

### `golden/policy_resolutions.json`
Expected policy resolution outcomes for:
- Conflict resolution between competing classifications
- Weighted average calculations
- Priority rule applications

## Mocks

### `mocks/oauth_tokens.json`
Mock OAuth2 tokens for testing:
- Valid tokens with proper expiration
- Expired tokens for refresh testing
- Invalid tokens for error handling

### `mocks/config_templates.yaml`
Configuration templates for different test scenarios:
- **test_config_minimal** - Basic configuration for unit tests
- **test_config_full** - Complete configuration with all features enabled

## Usage in Tests

### Loading Test Data

```go
// Load email fixtures
func loadTestEmails() ([]types.Email, error) {
    data, err := os.ReadFile("testdata/fixtures/emails.json")
    if err != nil {
        return nil, err
    }
    
    var emails []types.Email
    err = json.Unmarshal(data, &emails)
    return emails, err
}

// Load expected outputs
func loadGoldenOutputs() (map[string]interface{}, error) {
    data, err := os.ReadFile("testdata/golden/classification_outputs.json")
    if err != nil {
        return nil, err
    }
    
    var outputs map[string]interface{}
    err = json.Unmarshal(data, &outputs)
    return outputs, err
}
```

### Integration Testing

```go
func TestEmailClassification(t *testing.T) {
    emails := loadTestEmails()
    expected := loadGoldenOutputs()
    
    for _, email := range emails {
        result := classifyEmail(email)
        expectedResult := expected[email.ID]
        
        assert.Equal(t, expectedResult.Action, result.Action)
        assert.InDelta(t, expectedResult.Confidence, result.Confidence, 0.05)
    }
}
```

### Mock Server Testing

```go
func TestWithMockOllama(t *testing.T) {
    // Start mock server with fixture responses
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        response := loadOllamaResponse(r.URL.Path)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()
    
    // Configure client to use mock server
    client := ollama.NewClient(&config.OllamaConfig{
        BaseURL: server.URL,
        DefaultModel: "qwen2.5:7b",
    })
    
    // Run tests with mock responses
}
```

## Data Validation

All test data follows these principles:

1. **Realistic Content** - Based on actual spam/phishing patterns
2. **Diverse Scenarios** - Covers edge cases and common patterns
3. **Consistent Format** - Matches production data structures
4. **Verifiable Results** - Golden files provide expected outcomes
5. **Security Conscious** - No real credentials or sensitive data

## Updating Test Data

When adding new test cases:

1. Add email samples to `fixtures/emails.json`
2. Create corresponding expected outputs in `golden/`
3. Update mock responses if needed
4. Document any new test scenarios
5. Ensure all data is sanitized and safe

## Performance Testing

The fixtures include emails of various sizes and complexities to test:
- Processing time under different loads
- Memory usage with large email bodies
- Classification accuracy across different content types
- Error handling with malformed data

This comprehensive test data enables thorough validation of all MailSentinel components while maintaining security and realistic testing scenarios.
