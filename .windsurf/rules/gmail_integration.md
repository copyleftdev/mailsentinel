# Gmail API Integration Rules

## Gmail Client Standards
- **OAuth scopes**: Minimal required scopes (gmail.readonly, gmail.modify)
- **Rate limiting**: Respect Gmail API quotas (250 quota units/user/second)
- **Batch operations**: Use batch requests for multiple email modifications
- **History API**: Use Gmail History for incremental processing
- **Error handling**: Exponential backoff for 429/5xx responses

## Email Processing
- **Fetch strategy**: Batch processing with configurable size (default 100)
- **Query filters**: Exclude spam/trash by default, support custom queries
- **Content extraction**: Plain text preferred, minimal HTML parsing
- **Size limits**: 10MB max email size, truncate large content
- **Encoding**: Handle various email encodings properly

## Modification Safety
- **Dry run mode**: Default to dry-run, require explicit --apply flag
- **Idempotency**: Track message IDs to prevent duplicate actions
- **Change detection**: Verify current state before modifications
- **Rollback capability**: Maintain audit trail for action reversal
- **Label management**: Use MailSentinel/* namespace for all labels

## Authentication & Security
- **Token storage**: Encrypt OAuth tokens at rest (AES-256)
- **Token refresh**: Automatic refresh with fallback to manual re-auth
- **Scope validation**: Verify granted scopes match requirements
- **Network security**: TLS 1.3, certificate validation
- **Access logging**: Log all API calls with correlation IDs

## Performance Optimization
- **Partial responses**: Request only needed fields
- **Compression**: Enable gzip compression
- **Connection pooling**: Reuse HTTP connections
- **Concurrent requests**: Respect rate limits while maximizing throughput
- **Caching**: Cache user profile and label information
