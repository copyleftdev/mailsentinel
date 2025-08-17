# Run Comprehensive Test Suite

**Description**: Execute full test suite including unit tests, integration tests, security tests, and performance benchmarks.

## Steps

1. **Pre-Test Setup**
   - Verify test environment is clean and isolated
   - Start local Ollama instance with qwen2.5:7b model
   - Set up test Gmail account with proper OAuth scopes
   - Initialize test data with PII-redacted email samples

2. **Unit Tests**
   - Run unit tests with race detection: `go test -race ./...`
   - Generate coverage report: `go test -coverprofile=coverage.out ./...`
   - Verify minimum 80% test coverage threshold
   - Check for any flaky or failing tests

3. **Integration Tests**
   - Test Gmail API integration with real test account
   - Verify Ollama client connectivity and model availability
   - Test profile loading, inheritance, and dependency resolution
   - Validate end-to-end email processing pipeline

4. **Security Tests**
   - Run input validation tests with malicious payloads
   - Test resource exhaustion and DoS protection
   - Verify encrypted storage and token security
   - Check audit log integrity and tamper detection

5. **Performance Benchmarks**
   - Benchmark single email processing (target: p95 ≤ 1.5s)
   - Test batch processing performance (100 emails in ≤ 30s)
   - Monitor memory usage (≤ 512MB for 1000 email batch)
   - Verify concurrent processing efficiency

6. **Test Reporting**
   - Generate comprehensive test report with coverage metrics
   - Document any failing tests with root cause analysis
   - Create performance baseline for regression testing
   - Update test documentation with new findings
