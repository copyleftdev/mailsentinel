# Testing Standards for MailSentinel

## Test Strategy
- **Unit tests**: 80%+ coverage for all core packages
- **Integration tests**: Gmail API, Ollama client, end-to-end pipeline
- **Performance tests**: Latency benchmarks, memory usage, load testing
- **Security tests**: Input validation, injection resistance, resource limits

## Test Structure
```
testdata/
├── fixtures/
│   ├── spam/           # Spam email samples
│   ├── legitimate/     # Legitimate email samples
│   ├── phishing/       # Phishing attempt samples
│   └── edge_cases/     # Edge case scenarios
├── golden/
│   ├── decisions/      # Expected policy decisions
│   ├── audit/          # Expected audit log entries
│   └── profiles/       # Expected profile outputs
└── mocks/              # Generated mocks
```

## Testing Requirements
- **Mocking**: Use gomock for external dependencies (Gmail API, Ollama)
- **Fixtures**: Real email samples with PII redacted
- **Golden files**: Expected outputs for regression testing
- **Benchmarks**: Performance tests for all critical paths
- **Race detection**: All tests run with -race flag

## Test Categories
- **Unit tests**: Individual component testing with mocks
- **Integration tests**: Real Ollama instance, test Gmail account
- **Performance tests**: Latency, throughput, memory benchmarks
- **Security tests**: Malicious input handling, resource exhaustion
- **End-to-end tests**: Full pipeline with real data

## Quality Gates
- **Coverage**: Minimum 80% test coverage
- **Performance**: All benchmarks within SLA targets
- **Security**: No high-severity vulnerabilities
- **Reliability**: All tests pass consistently

## Test Data Management
- **PII redaction**: Remove all personal information from test emails
- **Synthetic data**: Generate realistic but fake email content
- **Edge cases**: Malformed emails, large attachments, encoding issues
- **Profile testing**: Validate inheritance, dependencies, conditional execution
