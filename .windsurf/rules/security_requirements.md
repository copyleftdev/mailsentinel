# Security Requirements for MailSentinel

## Security Model
- **Threat Model**: Prevent data exfiltration, prompt injection, DoS attacks, credential theft
- **Local-only processing**: Zero cloud LLM calls, all inference via localhost Ollama
- **Principle of least privilege**: Minimal Gmail API scopes, restricted file permissions
- **Defense in depth**: Multiple security layers and validation points

## Input Validation Requirements
- **Email content**: Size limits (10MB max), encoding validation, malicious content detection
- **YAML profiles**: Schema validation, injection prevention, safe parsing with go-yaml
- **Configuration files**: Type checking, bounds validation, input sanitization
- **API requests**: Request validation, rate limiting, authentication checks

## Data Protection Standards
- **Encryption at rest**: AES-256-GCM for OAuth tokens, audit logs, sensitive config
- **Key management**: PBKDF2/Argon2 key derivation, automated rotation, secure deletion
- **File permissions**: 0600 for secrets, 0644 for configs, 0640 for logs
- **Memory protection**: Secure allocation, automatic cleanup, no memory dumps

## Network Security
- **Ollama communication**: Localhost-only (127.0.0.1:11434)
- **Gmail API**: TLS 1.3, certificate validation, proper OAuth scopes
- **No external calls**: Block all outbound connections except Gmail API
- **Firewall rules**: Restrict network access, monitor connection attempts

## Audit and Compliance
- **Integrity verification**: SHA-256 checksums for all audit logs
- **Tamper detection**: Cryptographic verification, immutable log entries
- **Access logging**: All secret access, configuration changes, security events
- **Retention policies**: Configurable retention, secure deletion, compliance-ready

## Resource Protection
- **DoS prevention**: Request throttling, resource limits (CPU/memory/disk)
- **Circuit breakers**: Automatic service isolation, failure detection
- **Rate limiting**: Gmail API compliance, Ollama request throttling
- **Monitoring**: Security event detection, anomaly alerts, threshold monitoring

## Development Security
- **Code review**: Security-focused reviews for all changes
- **Static analysis**: Security scanning, vulnerability detection
- **Dependency scanning**: Regular updates, vulnerability monitoring
- **Secrets management**: No hardcoded secrets, encrypted storage, rotation
