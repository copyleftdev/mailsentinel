package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"

	"github.com/mailsentinel/core/pkg/config"
	"github.com/mailsentinel/core/pkg/types"
)

// Client represents an Ollama API client with circuit breaker
type Client struct {
	baseURL        string
	httpClient     *http.Client
	circuitBreaker *gobreaker.CircuitBreaker
	logger         *logrus.Logger
	config         *config.OllamaConfig
}

// GenerateRequest represents a request to Ollama's generate API
type GenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt,omitempty"`
	System   string                 `json:"system,omitempty"`
	Messages []Message              `json:"messages,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Stream   bool                   `json:"stream"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GenerateResponse represents Ollama's response
type GenerateResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// ModelInfo represents model information
type ModelInfo struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
}

// ListModelsResponse represents the response from /api/tags
type ListModelsResponse struct {
	Models []ModelInfo `json:"models"`
}

// NewClient creates a new Ollama client with circuit breaker
func NewClient(cfg *config.OllamaConfig, logger *logrus.Logger) *Client {
	// Configure circuit breaker
	cbSettings := gobreaker.Settings{
		Name:        "ollama-client",
		MaxRequests: cfg.CircuitBreaker.MaxRequests,
		Interval:    cfg.CircuitBreaker.Interval,
		Timeout:     cfg.CircuitBreaker.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= uint32(cfg.CircuitBreaker.ReadyToTrip)
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.WithFields(logrus.Fields{
				"circuit_breaker": name,
				"from_state":      from,
				"to_state":        to,
			}).Info("Circuit breaker state changed")
		},
	}

	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
		circuitBreaker: gobreaker.NewCircuitBreaker(cbSettings),
		logger:         logger,
		config:         cfg,
	}
}

// ClassifyEmailOld sends an email to Ollama for classification (old implementation)
func (c *Client) ClassifyEmailOld(ctx context.Context, email *types.Email, profile *types.Profile) (*types.ClassificationResponse, error) {
	startTime := time.Now()
	
	c.logger.WithFields(logrus.Fields{
		"email_id":   email.ID,
		"profile_id": profile.ID,
		"model":      profile.Model,
	}).Info("Classifying email with Ollama")

	// Build messages with few-shot examples
	messages := make([]Message, 0, len(profile.FewShot)+1)
	
	// Add few-shot examples
	for _, example := range profile.FewShot {
		messages = append(messages, Message{
			Role:    "user",
			Content: example.Input,
		})
		messages = append(messages, Message{
			Role:    "assistant", 
			Content: example.Output,
		})
	}
	
	// Add current email
	emailContent := fmt.Sprintf("Subject: %s\nFrom: %s\nBody: %s", 
		email.Subject, email.From, email.Body)
	messages = append(messages, Message{
		Role:    "user",
		Content: emailContent,
	})

	// Build request
	request := GenerateRequest{
		Model:    profile.Model,
		System:   profile.System,
		Messages: messages,
		Format:   "json",
		Options: map[string]interface{}{
			"temperature": profile.ModelParams.Temperature,
			"num_predict": profile.ModelParams.MaxTokens,
		},
		Stream: false,
	}

	// Execute with circuit breaker
	result, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.generate(ctx, &request)
	})
	
	if err != nil {
		return nil, fmt.Errorf("classification failed: %w", err)
	}
	
	response := result.(*GenerateResponse)
	
	// Parse JSON response
	var classificationResult types.ClassificationResponse
	if err := json.Unmarshal([]byte(response.Response), &classificationResult); err != nil {
		c.logger.WithError(err).WithField("response", response.Response).Error("Failed to parse classification response")
		return nil, fmt.Errorf("failed to parse classification response: %w", err)
	}
	
	// Set metadata
	classificationResult.ProfileID = profile.ID
	classificationResult.ProcessedAt = time.Now()
	
	// Log performance metrics
	duration := time.Since(startTime)
	c.logger.WithFields(logrus.Fields{
		"email_id":         email.ID,
		"profile_id":       profile.ID,
		"action":           classificationResult.Action,
		"confidence":       classificationResult.Confidence,
		"duration_ms":      duration.Milliseconds(),
		"total_duration":   response.TotalDuration,
		"eval_count":       response.EvalCount,
		"eval_duration":    response.EvalDuration,
	}).Info("Email classification completed")
	
	return &classificationResult, nil
}

// generate sends a request to Ollama's generate API
func (c *Client) generate(ctx context.Context, request *GenerateRequest) (*GenerateResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	url := fmt.Sprintf("%s/api/generate", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	var response GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &response, nil
}

// ListModels retrieves available models from Ollama
func (c *Client) ListModels(ctx context.Context) ([]ModelInfo, error) {
	url := fmt.Sprintf("%s/api/tags", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	var response ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return response.Models, nil
}

// HealthCheck verifies Ollama connectivity and model availability
func (c *Client) HealthCheck(ctx context.Context) error {
	// Check if service is responding
	models, err := c.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("Ollama health check failed: %w", err)
	}
	
	// Check if default model is available
	defaultModel := c.config.DefaultModel
	for _, model := range models {
		if model.Name == defaultModel {
			c.logger.WithField("model", defaultModel).Info("Default model is available")
			return nil
		}
	}
	
	return fmt.Errorf("default model %s not found in available models", defaultModel)
}

// GetCircuitBreakerState returns the current circuit breaker state
func (c *Client) GetCircuitBreakerState() gobreaker.State {
	return c.circuitBreaker.State()
}

// GetCircuitBreakerCounts returns the current circuit breaker counts
func (c *Client) GetCircuitBreakerCounts() gobreaker.Counts {
	return c.circuitBreaker.Counts()
}

// ClassifyEmail classifies an email using the specified profile
func (c *Client) ClassifyEmail(ctx context.Context, profile *types.Profile, email *types.Email) (*types.ClassificationResponse, error) {
	// Build the prompt from profile and email
	prompt := c.buildClassificationPrompt(profile, email)
	
	// Create generate request
	request := GenerateRequest{
		Model:  profile.Model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": profile.ModelParams.Temperature,
			"num_predict": profile.ModelParams.MaxTokens,
		},
	}
	
	// Make the request through circuit breaker
	result, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.generate(ctx, &request)
	})
	
	if err != nil {
		return nil, fmt.Errorf("classification request failed: %w", err)
	}
	
	response := result.(*GenerateResponse)
	
	// Parse the response into classification result
	classification, err := c.parseClassificationResponse(response.Response, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse classification response: %w", err)
	}
	
	return classification, nil
}

// buildClassificationPrompt constructs the prompt for email classification
func (c *Client) buildClassificationPrompt(profile *types.Profile, email *types.Email) string {
	var prompt strings.Builder
	
	// Add system prompt with strict JSON enforcement
	prompt.WriteString("System: ")
	prompt.WriteString(profile.System)
	prompt.WriteString(" You must respond with valid JSON only, no markdown, no explanations, no code blocks.")
	prompt.WriteString("\n\n")
	
	// Add few-shot examples if available
	for _, example := range profile.FewShot {
		prompt.WriteString("Example: ")
		prompt.WriteString(example.Name)
		prompt.WriteString("\n")
		prompt.WriteString("Input: ")
		prompt.WriteString(example.Input)
		prompt.WriteString("\n")
		prompt.WriteString("Output: ")
		prompt.WriteString(example.Output)
		prompt.WriteString("\n\n")
	}
	
	// Add the email to classify
	prompt.WriteString("Classify this email:\n")
	prompt.WriteString("Subject: ")
	prompt.WriteString(email.Subject)
	prompt.WriteString("\n")
	prompt.WriteString("From: ")
	prompt.WriteString(email.From)
	prompt.WriteString("\n")
	prompt.WriteString("To: ")
	prompt.WriteString(strings.Join(email.To, ", "))
	prompt.WriteString("\n")
	prompt.WriteString("Body: ")
	prompt.WriteString(email.Body)
	prompt.WriteString("\n\n")
	
	// Add strict response format instruction
	prompt.WriteString("\n\nIMPORTANT: You MUST respond with ONLY valid JSON in this exact format:\n")
	prompt.WriteString(`{"action": "string", "confidence": number, "reasoning": "string"}`)
	prompt.WriteString("\n\nDo NOT include any markdown formatting, explanations, or additional text.")
	prompt.WriteString("\nDo NOT wrap the JSON in code blocks or backticks.")
	prompt.WriteString("\nRespond with raw JSON only.")
	
	return prompt.String()
}

// parseClassificationResponse parses the LLM response into a classification result
func (c *Client) parseClassificationResponse(response string, profile *types.Profile) (*types.ClassificationResponse, error) {
	// Try to extract JSON from the response
	var result map[string]interface{}
	
	// First try to extract from markdown code blocks
	jsonStr := ""
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json")
		if start != -1 {
			start += 7 // Skip "```json"
			end := strings.Index(response[start:], "```")
			if end != -1 {
				jsonStr = strings.TrimSpace(response[start : start+end])
			} else {
				// Handle case where closing ``` is missing (truncated response)
				jsonStr = strings.TrimSpace(response[start:])
			}
		}
	}
	
	// If no markdown block found, find JSON in the response
	if jsonStr == "" {
		start := strings.Index(response, "{")
		end := strings.LastIndex(response, "}")
		
		if start == -1 || end == -1 || start >= end {
			return nil, fmt.Errorf("no valid JSON found in response: %s", response)
		}
		
		jsonStr = response[start : end+1]
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	
	// Extract required fields
	action, ok := result["action"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'action' field in response")
	}
	
	confidence, ok := result["confidence"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'confidence' field in response")
	}
	
	reasoning, ok := result["reasoning"].(string)
	if !ok {
		reasoning = "No reasoning provided"
	}
	
	// Validate confidence range
	if confidence < 0.0 || confidence > 1.0 {
		return nil, fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", confidence)
	}
	
	// Create classification response
	classification := &types.ClassificationResponse{
		ProfileID:   profile.ID,
		Action:      action,
		Confidence:  confidence,
		Reasoning:   reasoning,
		ProcessedAt: time.Now(),
	}
	
	// Add metadata if present
	if metadata, exists := result["metadata"]; exists {
		if metadataMap, ok := metadata.(map[string]interface{}); ok {
			classification.Metadata = metadataMap
		}
	}
	
	// Add labels if present
	if labels, exists := result["labels"]; exists {
		if labelsList, ok := labels.([]interface{}); ok {
			var stringLabels []string
			for _, label := range labelsList {
				if labelStr, ok := label.(string); ok {
					stringLabels = append(stringLabels, labelStr)
				}
			}
			classification.Labels = stringLabels
		}
	}
	
	return classification, nil
}
