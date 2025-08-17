package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	if err := validateTestData(); err != nil {
		fmt.Fprintf(os.Stderr, "Test data validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ All test data is valid")
}

func validateTestData() error {
	testDataDir := "testdata"
	
	// Validate JSON files
	jsonFiles := []string{
		"fixtures/emails.json",
		"fixtures/gmail_responses.json", 
		"fixtures/ollama_responses.json",
		"fixtures/audit_logs.json",
		"golden/classification_outputs.json",
		"golden/policy_resolutions.json",
		"mocks/oauth_tokens.json",
	}
	
	for _, file := range jsonFiles {
		path := filepath.Join(testDataDir, file)
		if err := validateJSONFile(path); err != nil {
			return fmt.Errorf("invalid JSON in %s: %w", file, err)
		}
	}
	
	// Validate YAML files
	yamlFiles := []string{
		"fixtures/profiles.yaml",
		"mocks/config_templates.yaml",
	}
	
	for _, file := range yamlFiles {
		path := filepath.Join(testDataDir, file)
		if err := validateYAMLFile(path); err != nil {
			return fmt.Errorf("invalid YAML in %s: %w", file, err)
		}
	}
	
	// Validate email fixtures structure
	if err := validateEmailFixtures(); err != nil {
		return fmt.Errorf("email fixtures validation failed: %w", err)
	}
	
	// Validate golden files consistency
	if err := validateGoldenFiles(); err != nil {
		return fmt.Errorf("golden files validation failed: %w", err)
	}
	
	return nil
}

func validateJSONFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	var obj interface{}
	return json.Unmarshal(data, &obj)
}

func validateYAMLFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	var obj interface{}
	return yaml.Unmarshal(data, &obj)
}

func validateEmailFixtures() error {
	data, err := os.ReadFile("testdata/fixtures/emails.json")
	if err != nil {
		return err
	}
	
	var emails []map[string]interface{}
	if err := json.Unmarshal(data, &emails); err != nil {
		return err
	}
	
	requiredFields := []string{"id", "subject", "from", "to", "body", "classification", "expected_action"}
	
	for i, email := range emails {
		for _, field := range requiredFields {
			if _, exists := email[field]; !exists {
				return fmt.Errorf("email %d missing required field: %s", i, field)
			}
		}
		
		// Validate confidence is between 0 and 1
		if conf, exists := email["expected_confidence"]; exists {
			if confFloat, ok := conf.(float64); ok {
				if confFloat < 0 || confFloat > 1 {
					return fmt.Errorf("email %d has invalid confidence: %f", i, confFloat)
				}
			}
		}
	}
	
	return nil
}

func validateGoldenFiles() error {
	// Load emails to get valid IDs
	emailData, err := os.ReadFile("testdata/fixtures/emails.json")
	if err != nil {
		return err
	}
	
	var emails []map[string]interface{}
	if err := json.Unmarshal(emailData, &emails); err != nil {
		return err
	}
	
	emailIDs := make(map[string]bool)
	for _, email := range emails {
		if id, ok := email["id"].(string); ok {
			emailIDs[id] = true
		}
	}
	
	// Validate classification outputs reference valid emails
	goldData, err := os.ReadFile("testdata/golden/classification_outputs.json")
	if err != nil {
		return err
	}
	
	var goldOutputs map[string]interface{}
	if err := json.Unmarshal(goldData, &goldOutputs); err != nil {
		return err
	}
	
	for key, value := range goldOutputs {
		if data, ok := value.(map[string]interface{}); ok {
			if input, exists := data["input"].(map[string]interface{}); exists {
				if emailID, ok := input["email_id"].(string); ok {
					if !emailIDs[emailID] {
						return fmt.Errorf("golden file %s references unknown email ID: %s", key, emailID)
					}
				}
			}
		}
	}
	
	return nil
}
