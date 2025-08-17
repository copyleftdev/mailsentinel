# Debug Production Issue

**Description**: Systematic approach to diagnosing and resolving production issues in MailSentinel.

## Steps

1. **Issue Triage**
   - Collect error symptoms and user reports
   - Check system health dashboards and alerts
   - Review recent deployments and configuration changes
   - Determine severity level and impact scope

2. **Log Analysis**
   - Examine audit logs for error patterns and correlation IDs
   - Check Ollama connectivity and model performance logs
   - Review Gmail API rate limiting and authentication logs
   - Analyze resource usage (CPU, memory, disk) trends

3. **Component Isolation**
   - Test Gmail API connectivity independently
   - Verify Ollama health and model availability
   - Check profile loading and validation systems
   - Test email processing pipeline components

4. **Root Cause Analysis**
   - Reproduce issue in development environment
   - Add detailed logging to problematic code paths
   - Use profiling tools to identify performance bottlenecks
   - Check for resource leaks or concurrency issues

5. **Resolution Implementation**
   - Implement fix with comprehensive testing
   - Deploy hotfix using blue-green deployment strategy
   - Monitor system behavior post-deployment
   - Verify issue resolution with affected users

6. **Post-Incident Review**
   - Document root cause and resolution steps
   - Update monitoring and alerting rules
   - Improve error handling and logging
   - Create preventive measures for similar issues
