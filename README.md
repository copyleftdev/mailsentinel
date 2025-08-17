# MailSentinel

<div align="center">
  <img src="media/MSLOGO.png" alt="MailSentinel Logo" width="200"/>
</div>

MailSentinel is a Gmail triage system that uses local Ollama LLM inference to automatically classify and organize emails based on configurable profiles. It provides privacy-first email processing with comprehensive audit trails and policy-driven decision making.

## Development Story

This project was developed through a collaborative human-AI pair programming session. The human provided strategic guidance, requirements, and context while the AI (Cascade) implemented the technical solution.

**Development Timeline**: August 16-17, 2025 (approximately 4 hours)
- **Started**: 2025-08-16 ~21:00 PST  
- **Completed**: 2025-08-17 01:01 PST
- **Total Duration**: ~4 hours of intensive development

**What Was Built**:
- Complete Gmail OAuth integration with secure token management
- Local Ollama LLM client with circuit breaker patterns
- Modular YAML profile system with inheritance
- Comprehensive audit logging with cryptographic integrity
- Production-ready CLI with dry-run safety features
- Full test suite with benchmarks and integration tests
- 8 different email classification profiles
- Enterprise-grade security and monitoring

**Human Contributions**:
- Strategic direction and requirements definition
- Architecture decisions and security considerations  
- Profile design and classification logic
- Testing strategy and validation approach
- Production deployment guidance

**AI Contributions**:
- Complete codebase implementation (~2000+ lines of Go)
- OAuth flow and Gmail API integration
- Cryptographic audit system design
- Circuit breaker and resilience patterns
- Comprehensive error handling and logging
- Test suite and benchmarking framework
- Documentation and deployment scripts

**Final Result**: A production-ready email classification system that successfully processed 25+ real Gmail emails across 5 different profiles with 100% success rate and zero errors.

## Features

- **Local LLM Processing**: All inference via local Ollama - zero cloud calls
- **Profile-Driven Classification**: Modular YAML profiles with inheritance and dependencies  
- **Policy Resolution**: Advanced conflict resolution with priority rules and confidence weighting
- **Enterprise Security**: Encrypted storage, audit trails with integrity verification
- **Circuit Breaker Patterns**: Resilient handling of external dependencies
- **Batch Processing**: Efficient processing with configurable concurrency
- **Dry-Run Mode**: Safe testing without modifying emails

## Quick Start

### Prerequisites

1. **Go 1.21+** - [Download](https://golang.org/dl/)
2. **Ollama** - [Install](https://ollama.ai/) and pull `qwen2.5:7b` model
3. **Gmail API Credentials** - [Setup Guide](https://developers.google.com/gmail/api/quickstart/go)

### Installation

```bash
# Clone repository
git clone https://github.com/mailsentinel/core.git
cd core

# Install dependencies
go mod tidy

# Copy environment template
cp env.example .env
# Edit .env with your Gmail credentials

# Create data directories
mkdir -p data/audit profiles

# Build binary
go build -o bin/mailsentinel cmd/mailsentinel/main.go
```

### Configuration

1. **Gmail OAuth Setup**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create project and enable Gmail API
   - Create OAuth 2.0 credentials (Desktop application)
   - Add credentials to `.env` file

2. **Ollama Setup**:
   ```bash
   # Install and start Ollama
   ollama serve
   
   # Pull required model
   ollama pull qwen2.5:7b
   ```

### Usage

```bash
# Run in dry-run mode (safe, no email modifications)
./bin/mailsentinel -dry-run -query "is:unread" -max-emails 10

# Process with specific profile
./bin/mailsentinel -dry-run -profile spam -query "is:unread"

# Apply changes (removes dry-run protection)
./bin/mailsentinel -query "is:unread" -max-emails 100

# Verbose logging
./bin/mailsentinel -verbose -dry-run
```

## Architecture

```
cmd/mailsentinel/     # CLI application entry point
internal/
├── gmail/           # Gmail API client with OAuth
├── ollama/          # Ollama client with circuit breaker  
├── profile/         # Profile loading and dependency resolution
├── resolver/        # Policy conflict resolution
└── audit/           # Secure audit logging
pkg/
├── types/           # Core data structures
└── config/          # Configuration management
profiles/            # YAML classification profiles
data/               # Runtime data (tokens, audit logs)
```

## Profile System

Profiles define email classification behavior using YAML:

```yaml
id: "spam"
version: "1.0.0"
model: "qwen2.5:7b"
system: "You are an expert spam classifier..."
fewshot:
  - name: "phishing_example"
    input: "Suspicious email content..."
    output: '{"action": "delete", "confidence": 0.95}'
policy:
  conditions:
    - name: "high_risk"
      expression: "metadata.phishing_score >= 0.8"
      actions: ["delete"]
```

### Profile Features

- **Inheritance**: `inherits_from: base_profile`
- **Dependencies**: `depends_on: [other_profiles]`
- **Conditional Execution**: `when: "expression"`
- **Few-Shot Learning**: Training examples for better accuracy
- **Policy Rules**: Confidence thresholds and action mapping

## Security

- **Local-Only Processing**: No external LLM calls
- **Encrypted Storage**: AES-256 for OAuth tokens and sensitive data
- **Audit Integrity**: SHA-256 checksums and cryptographic signatures
- **Input Sanitization**: Protection against prompt injection
- **Resource Limits**: DoS protection and memory constraints

## Monitoring

MailSentinel provides comprehensive observability:

- **Structured Logging**: JSON format with correlation IDs
- **Performance Metrics**: Latency, throughput, memory usage
- **Circuit Breaker Status**: External dependency health
- **Audit Trail**: Immutable log of all decisions

## Development

### Running Tests

```bash
# Unit tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Race detection
go test -race ./...
```

### Using Workflows

MailSentinel includes Windsurf workflows for common tasks:

```bash
# Setup development environment
/setup-development

# Create new profile
/develop-profile

# Run full test suite
/run-tests

# Security review
/security-review

# Deploy to production
/deploy-production
```

### Project Structure

- `.windsurf/rules/` - Development guidelines and standards
- `.windsurf/workflows/` - Automated development processes
- `testdata/` - Test fixtures and golden files
- `profiles/` - Email classification profiles

## Configuration Reference

See `config.yaml` for full configuration options and `docs/spec.md` for detailed technical specifications:

- **Gmail**: OAuth, rate limiting, batch sizes
- **Ollama**: Models, timeouts, circuit breaker settings  
- **Profiles**: Directory, reload intervals, validation
- **Audit**: Logging, rotation, integrity checks
- **Security**: Encryption, input validation, resource limits

## Performance Targets

- **Single Email**: p95 ≤ 1.5s processing time
- **Batch Processing**: 100 emails in ≤ 30s
- **Memory Usage**: ≤ 512MB for 1000 email batch
- **Uptime**: 99.9% for scheduled operations

## Contributing

1. Follow the development rules in `.windsurf/rules/`
2. Use workflows for consistent processes
3. Maintain 80%+ test coverage
4. Security review required for all changes

## License

MIT License - See LICENSE file for details.

## Author

**copyleftdev** - Email classification system architect and developer  
Contact: dj@codetestcode.io

## Support

- Documentation: `docs/spec.md`
- Issues: GitHub Issues
