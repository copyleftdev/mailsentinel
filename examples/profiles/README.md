# Example Email Classification Profiles - Spec Compliant

This directory contains example YAML profiles for the MailSentinel email classification system. Each profile defines a specific classification intent with comprehensive schemas, inheritance, dependencies, and advanced policy engines per the MailSentinel specification.

## Available Profiles

### Base Classifier (`base_classifier.yaml`)
- **Intent**: Foundation profile providing common configuration for all classifiers
- **Use Case**: Inherited by all other profiles for consistent schema and security rules
- **Features**: Standard response schema, security rules, base system prompt

### 1. Spam Detection (`spam_detection.yaml`)
- **Intent**: Identify and classify spam emails with comprehensive risk assessment
- **Inherits**: `base_classifier`
- **Dependencies**: None (base profile)
- **Features**: Advanced phishing detection, authentication analysis, risk scoring
- **Policy Engine**: High-confidence spam archival, security threat labeling

### 2. Phishing Detection (`phishing_detection.yaml`)
- **Intent**: Detect sophisticated phishing and social engineering attacks
- **Inherits**: `base_classifier`
- **Dependencies**: `spam_detection`
- **Features**: Lookalike domain detection, authentication failure analysis, urgency tactics detection
- **Policy Engine**: Security-first labeling, authentication failure tracking

### 3. Newsletter Classification (`newsletter_classifier.yaml`)
- **Intent**: Classify newsletters and subscription emails for organization
- **Inherits**: `base_classifier`
- **Dependencies**: `spam_detection`, `phishing_detection`
- **Features**: Content quality assessment, newsletter type classification, engagement value analysis
- **Policy Engine**: Quality-based labeling, promotional content archival

### 4. Promotional Filter (`promotional_filter.yaml`)
- **Intent**: Filter promotional emails by quality and brand recognition
- **Inherits**: `base_classifier`
- **Dependencies**: `spam_detection`
- **Features**: Offer quality assessment, brand recognition scoring, discount analysis
- **Policy Engine**: High-value offer starring, low-quality promotion archival

### 5. Support Ticket Classification (`support_ticket.yaml`)
- **Intent**: Classify customer support emails by urgency and routing requirements
- **Inherits**: `base_classifier`
- **Dependencies**: `spam_detection`
- **Features**: Issue severity assessment, customer tier classification, SLA-based routing
- **Policy Engine**: Critical issue escalation, enterprise customer prioritization

### 6. Work Priority Classification (`work_priority.yaml`)
- **Intent**: Prioritize work emails by urgency, importance, and sender role
- **Inherits**: `base_classifier`
- **Dependencies**: `spam_detection`
- **Features**: Executive communication detection, deadline analysis, business impact assessment
- **Policy Engine**: Critical importance starring, urgent deadline flagging

## Enhanced Profile Structure (Spec v2.1)

Each profile follows the comprehensive MailSentinel specification:

```yaml
id: "unique_profile_id"
name: "Human Readable Name"
version: "2.1"
description: "Profile description"
inherits_from: "base_classifier"
depends_on: ["dependency_profile"]
conditional_execution:
  when: 'condition_expression'
  reason: "Execution logic explanation"

model: "qwen2.5:latest"
model_params:
  temperature: 0.1
  max_tokens: 1000
  timeout_seconds: 30

response:
  schema: |
    {
      "category": "spam|promotions|updates|social|personal|work|security",
      "importance": "low|normal|high|critical",
      "urgency_hours": 0,
      "action": "none|star|archive|label",
      "confidence": 0.0,
      "reasons": ["string"],
      "risk_factors": {
        "phishing_score": 0.0,
        "malware_risk": "low|medium|high",
        "social_engineering": false
      },
      "features": {
        "auth": "string",
        "sender_domain": "string",
        "sender_reputation_score": 0.0
      }
    }
  validation:
    required_fields: ["category", "confidence", "action"]
    confidence_range: [0.0, 1.0]
    max_reasons: 5

system: |
  Enhanced system prompt with security rules and analysis framework

fewshot:
  - name: "example_name"
    input: |
      {
        "headers": {"Authentication-Results": "dkim=pass spf=pass dmarc=pass"},
        "subject": "Example Subject",
        "sender_domain": "example.com",
        "plain": "Email content"
      }
    output: |
      {
        "category": "work",
        "importance": "normal",
        "confidence": 0.85,
        "reasons": ["specific", "evidence-based", "reasoning"]
      }

policy:
  conditions:
    - name: "condition_name"
      expression: 'logical_expression'
      actions: ["action1", "action2"]
      priority: 100
  conflict_resolution:
    - "resolution rule 1"
    - "resolution rule 2"
  default_action: "none"
```

## Key Enhancements

### 1. **Inheritance System**
- All profiles inherit from `base_classifier` for consistency
- Shared security rules and response schema
- Modular configuration management

### 2. **Dependency Management**
- Profiles can depend on other profiles
- Conditional execution based on dependency results
- Sophisticated filtering pipelines

### 3. **Enhanced Response Schema**
- Comprehensive risk factor analysis
- Detailed feature extraction
- Standardized confidence and reasoning

### 4. **Advanced Policy Engine**
- Rule-based action assignment
- Priority-based conflict resolution
- Flexible labeling and routing

### 5. **Security-First Design**
- Non-negotiable security rules in all profiles
- Prompt injection resistance
- Conservative bias for important emails

## Usage

1. **Deploy all profiles**:
   ```bash
   cp examples/profiles/*.yaml ./profiles/
   ```

2. **Test dependency chain**:
   ```bash
   ./bin/mailsentinel --profile work_priority --dry-run
   ```

3. **Monitor policy execution**:
   ```bash
   ./bin/mailsentinel --profile spam_detection --verbose
   ```

## Customization Guide

### **Few-shot Examples**
- Use structured JSON input format
- Include authentication headers and metadata
- Provide specific, evidence-based reasoning

### **Policy Conditions**
- Write clear logical expressions
- Set appropriate priority levels
- Define conflict resolution rules

### **Security Configuration**
- Never modify security rules
- Test with potentially malicious content
- Validate all user inputs

### **Performance Tuning**
- Adjust confidence thresholds per use case
- Monitor classification accuracy
- Optimize dependency chains

## Best Practices

1. **Security First**: Never compromise on security rules
2. **Test Thoroughly**: Use `--dry-run` extensively
3. **Monitor Performance**: Track confidence scores and accuracy
4. **Regular Updates**: Refresh examples as threats evolve
5. **Dependency Awareness**: Understand profile execution order
6. **Policy Validation**: Test conflict resolution scenarios
