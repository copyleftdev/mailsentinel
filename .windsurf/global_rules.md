# Global MailSentinel Development Rules

## Enterprise Standards
- **Security first**: All code must pass security review before production
- **Privacy by design**: No external LLM calls, localhost-only processing
- **Audit everything**: Comprehensive logging with integrity verification
- **Performance targets**: p95 ≤ 1.5s per email, ≤ 512MB memory usage
- **Error handling**: Graceful degradation, circuit breakers, exponential backoff

## Code Quality
- Go 1.21+ with proper error handling and context propagation
- Structured logging with correlation IDs for all operations
- Interface-driven design for testability and modularity
- Resource cleanup with defer statements and context cancellation
- Comprehensive unit tests with 80%+ coverage requirement

## Gmail Integration
- Minimal OAuth scopes (gmail.readonly, gmail.modify)
- Batch operations to respect API quotas (250 units/user/second)
- Dry-run mode by default, explicit --apply flag for modifications
- Idempotent operations with message ID tracking
- Use MailSentinel/* namespace for all labels

## Ollama Integration  
- Localhost-only communication (127.0.0.1:11434)
- JSON format for all structured responses
- Circuit breaker pattern for connection failures
- Input sanitization before sending to LLM
- Response validation against profile schemas

## Profile Development
- YAML-based with strict schema validation
- Semantic versioning for all profile updates
- Comprehensive few-shot examples covering edge cases
- Policy expressions with confidence thresholds
- Inheritance and dependency management support
