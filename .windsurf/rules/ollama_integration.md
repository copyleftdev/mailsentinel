# Ollama Integration Rules

## Ollama Client Standards
- **Endpoint**: Always use http://127.0.0.1:11434 (localhost only)
- **Models**: Default to qwen2.5:7b, support model switching per profile
- **JSON mode**: Always use `format:"json"` for structured outputs
- **Timeouts**: 30s default, configurable per profile
- **Circuit breaker**: 5 failure threshold, 60s recovery timeout

## LLM Request Structure
```go
type ClassifyRequest struct {
    Model       string        // "qwen2.5:7b"
    System      string        // Profile system prompt
    Messages    []Message     // Few-shot examples + current email
    Format      string        // "json"
    Options     ModelOptions  // Temperature, max_tokens, etc.
}
```

## Error Handling
- **Connection failures**: Circuit breaker activation, fallback to heuristics
- **Timeout handling**: Exponential backoff (1s, 2s, 4s, 8s)
- **Invalid JSON**: Schema validation, retry once, then action=none
- **Model unavailable**: Graceful degradation, user notification
- **Rate limiting**: Respect Ollama limits, queue management

## Response Validation
- **JSON schema**: Strict validation against profile schema
- **Confidence bounds**: Ensure 0.0 ≤ confidence ≤ 1.0
- **Required fields**: Validate all mandatory response fields
- **Sanitization**: Clean any potentially harmful content from responses
- **Performance tracking**: Log execution time, token usage

## Model Management
- **Health checks**: Regular connectivity and model availability checks
- **Version tracking**: Log model versions in audit trails
- **Performance monitoring**: Track latency, success rates, error patterns
- **Resource monitoring**: Monitor Ollama memory/CPU usage

## Security Considerations
- **Local only**: Never send requests to external LLM services
- **Input sanitization**: Clean email content before sending to Ollama
- **Output validation**: Verify responses don't contain instructions or code
- **Prompt injection**: Resist attempts to manipulate system behavior
- **Resource limits**: Prevent DoS through request throttling
