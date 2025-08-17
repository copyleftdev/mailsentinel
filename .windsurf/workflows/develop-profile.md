# Develop Email Classification Profile

**Description**: Create and validate a new email classification profile with proper testing, validation, and integration.

## Steps

1. **Profile Design**
   - Analyze email classification requirements and use cases
   - Define profile schema with required fields (id, version, model, system prompt)
   - Design few-shot examples covering edge cases and typical scenarios
   - Create policy expressions with confidence thresholds

2. **YAML Implementation**
   - Create profile YAML file following naming convention: `profiles/[category]-[version].yaml`
   - Implement inheritance if extending existing profile using `inherits_from`
   - Add dependencies with `depends_on` if profile requires other profiles
   - Include conditional execution logic with `when` expressions

3. **Validation & Testing**
   - Validate YAML syntax and schema compliance
   - Test profile against sample emails in `testdata/fixtures/`
   - Verify few-shot examples produce expected outputs
   - Check policy expressions evaluate correctly

4. **Integration Testing**
   - Test profile loading and inheritance resolution
   - Verify Ollama integration with specified model
   - Test policy execution and action determination
   - Validate audit logging captures all decisions

5. **Performance Validation**
   - Benchmark profile execution time (target: <1.5s per email)
   - Test memory usage with large email batches
   - Verify confidence calibration accuracy
   - Check error handling and fallback behavior

6. **Documentation & Deployment**
   - Document profile purpose, use cases, and configuration
   - Add profile to version control with proper commit message
   - Update profile registry and dependency mappings
   - Create deployment checklist for production use
