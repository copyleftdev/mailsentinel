# MailSentinel Core Development Rules

## Project Context
This is MailSentinel, an enterprise-grade Gmail triage system using local Ollama LLM inference. The system processes emails through configurable profiles with dependency management, policy resolution, and comprehensive audit trails.

## Architecture Principles
- **Local-only processing**: All LLM inference via local Ollama, zero cloud calls
- **Batch processing**: Handle emails in configurable batches with concurrency control
- **Circuit breaker patterns**: Resilient external dependency handling
- **Profile-driven**: Modular YAML profiles with inheritance and dependencies
- **Enterprise-ready**: Comprehensive monitoring, backup, and disaster recovery

## Code Standards
- Go 1.21+ with modules enabled
- Structured logging with correlation IDs
- Comprehensive error handling with typed errors
- Interface-driven design for testability
- Resource limits and graceful degradation
- Idempotent operations for Gmail modifications

## Security Requirements
- Input sanitization for all email content and YAML profiles
- Encrypted storage for OAuth tokens and sensitive data
- Audit log integrity with checksums
- No external network calls except Gmail API and localhost Ollama
- Resource isolation and DoS protection

## Performance Targets
- Single email processing: p95 ≤ 1.5s
- Batch processing (100 emails): p95 ≤ 30s
- Memory usage: ≤ 512MB for 1000 email batch
- 99.9% uptime for scheduled operations
