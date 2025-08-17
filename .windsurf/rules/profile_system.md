# Profile System Development Rules

## Profile Architecture
- YAML-based configuration with strict schema validation
- Profile inheritance using `inherits_from` field
- Dependency management with `depends_on` arrays
- Conditional execution with expression evaluation
- Hot-reload capability with validation

## Profile Structure Standards
```yaml
id: profile_name
version: semantic_version
inherits_from: base_profile  # optional
depends_on: [dependency_list]  # optional
conditional_execution:
  when: "expression"
  reason: "human_readable_explanation"
model: qwen2.5:7b
model_params:
  temperature: 0.1
  max_tokens: 1000
  timeout_seconds: 30
response:
  schema: |
    strict_json_schema
  validation:
    required_fields: []
    confidence_range: [0.0, 1.0]
system: |
  enhanced_system_prompt
fewshot:
  - name: example_name
    input: example_input
    output: expected_output
policy:
  conditions:
    - name: condition_name
      expression: policy_expression
      actions: [action_list]
      priority: numeric_priority
```

## Profile Development Guidelines
- Always include comprehensive few-shot examples
- Use semantic versioning for profile updates
- Test profile dependencies and conditional logic
- Validate JSON schema compliance
- Document policy expressions clearly
- Include confidence calibration rules

## Policy Expression Language
- Support boolean operators: &&, ||, !
- Comparison operators: ==, !=, >=, <=, >, <
- Field access: profile.field, sender_reputation.trust_score
- Array operations: in, contains
- Built-in functions: any(), all(), count()
