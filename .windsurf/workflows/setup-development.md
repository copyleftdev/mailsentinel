# Setup Development Environment

**Description**: Initialize a complete MailSentinel development environment with all dependencies, configuration, and testing setup.

## Steps

1. **Verify Prerequisites**
   - Check Go 1.21+ installation: `go version`
   - Verify Ollama is running: `curl http://127.0.0.1:11434/api/tags`
   - Confirm qwen2.5:7b model is available: `ollama list | grep qwen2.5:7b`

2. **Initialize Go Module**
   - Create `go.mod` with module path `github.com/mailsentinel/core`
   - Add essential dependencies: Gmail API, YAML parsing, crypto, logging
   - Set up proper directory structure: `cmd/`, `internal/`, `pkg/`, `testdata/`

3. **Configure Development Environment**
   - Create `.env.example` with required environment variables
   - Set up Gmail OAuth credentials (client_id, client_secret)
   - Configure logging levels and output formats
   - Create local configuration templates

4. **Setup Testing Infrastructure**
   - Install testing dependencies (testify, gomock)
   - Create test data directories with PII-redacted email samples
   - Set up golden file testing structure
   - Configure test coverage reporting

5. **Initialize Security Framework**
   - Set up encrypted storage for OAuth tokens
   - Configure audit logging with integrity checks
   - Implement input validation framework
   - Set up resource monitoring and limits

6. **Validate Setup**
   - Run health checks for all external dependencies
   - Execute basic connectivity tests (Gmail API, Ollama)
   - Verify security configurations
   - Run initial test suite to confirm everything works
