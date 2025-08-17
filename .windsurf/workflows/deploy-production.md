# Deploy to Production

**Description**: Deploy MailSentinel to production environment with proper validation, monitoring, and rollback capabilities.

## Steps

1. **Pre-Deployment Validation**
   - Run full test suite and verify all tests pass
   - Execute security review workflow: `/security-review`
   - Validate configuration files and environment variables
   - Check Ollama model availability and performance

2. **Build and Package**
   - Build production binary with optimizations: `go build -ldflags="-s -w"`
   - Create deployment package with configs and profiles
   - Generate checksums for integrity verification
   - Prepare deployment scripts and documentation

3. **Environment Setup**
   - Configure production Ollama instance with required models
   - Set up encrypted storage for OAuth tokens and secrets
   - Configure monitoring and alerting systems
   - Establish backup and disaster recovery procedures

4. **Deployment Execution**
   - Deploy binary to production environment
   - Configure systemd service for automatic startup
   - Set up log rotation and audit trail management
   - Initialize Gmail OAuth and verify connectivity

5. **Post-Deployment Validation**
   - Run health checks for all components
   - Test email processing with sample data
   - Verify monitoring and alerting functionality
   - Confirm backup systems are operational

6. **Production Monitoring**
   - Monitor performance metrics (latency, throughput, memory)
   - Track error rates and system health
   - Set up automated alerts for critical issues
   - Document operational procedures and troubleshooting
