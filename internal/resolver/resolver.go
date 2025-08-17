package resolver

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/mailsentinel/core/pkg/types"
)

// PolicyResolver handles conflict resolution between multiple profile results
type PolicyResolver struct {
	config *types.ResolverConfig
	logger *logrus.Logger
}

// NewPolicyResolver creates a new policy resolver
func NewPolicyResolver(configPath string, logger *logrus.Logger) (*PolicyResolver, error) {
	config, err := loadResolverConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load resolver config: %w", err)
	}

	return &PolicyResolver{
		config: config,
		logger: logger,
	}, nil
}

// loadResolverConfig loads resolver configuration from YAML file
func loadResolverConfig(path string) (*types.ResolverConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config types.ResolverConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// ResolveDecision resolves conflicts between multiple classification results
func (r *PolicyResolver) ResolveDecision(email *types.Email, results []*types.ClassificationResponse) (*types.ClassificationResponse, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no classification results provided")
	}

	if len(results) == 1 {
		return results[0], nil
	}

	r.logger.WithFields(logrus.Fields{
		"email_id":      email.ID,
		"result_count":  len(results),
	}).Info("Resolving classification conflicts")

	// Apply priority rules first
	if priorityResult := r.applyPriorityRules(email, results); priorityResult != nil {
		r.logger.WithFields(logrus.Fields{
			"email_id": email.ID,
			"action":   priorityResult.Action,
			"reason":   "priority_rule_override",
		}).Info("Applied priority rule override")
		return priorityResult, nil
	}

	// Apply confidence weighting
	weightedResults := r.applyConfidenceWeighting(results)

	// Resolve conflicts using conflict resolution matrix
	finalResult := r.resolveConflicts(weightedResults)

	r.logger.WithFields(logrus.Fields{
		"email_id":   email.ID,
		"action":     finalResult.Action,
		"confidence": finalResult.Confidence,
	}).Info("Resolved classification decision")

	return finalResult, nil
}

// applyPriorityRules checks if any priority rules should override normal resolution
func (r *PolicyResolver) applyPriorityRules(email *types.Email, results []*types.ClassificationResponse) *types.ClassificationResponse {
	// Sort priority rules by priority (highest first)
	rules := make([]*types.PriorityRule, len(r.config.PriorityRules))
	for i := range r.config.PriorityRules {
		rules[i] = &r.config.PriorityRules[i]
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	for _, rule := range rules {
		if r.evaluateCondition(rule.Condition, email, results) {
			r.logger.WithFields(logrus.Fields{
				"rule_name": rule.Name,
				"priority":  rule.Priority,
				"reason":    rule.Reason,
			}).Info("Priority rule triggered")

			// Create override result
			result := &types.ClassificationResponse{
				Action:      rule.Action,
				Confidence:  1.0, // Priority rules have maximum confidence
				Reasoning:   rule.Reason,
				ProcessedAt: time.Now(),
			}

			// Apply confidence boost if specified
			if rule.ConfidenceBoost > 0 {
				// Find highest confidence result and boost it
				var highestResult *types.ClassificationResponse
				for _, r := range results {
					if highestResult == nil || r.Confidence > highestResult.Confidence {
						highestResult = r
					}
				}
				if highestResult != nil {
					result.Action = highestResult.Action
					result.Confidence = min(1.0, highestResult.Confidence+rule.ConfidenceBoost)
					result.Reasoning = fmt.Sprintf("%s (boosted by %s)", highestResult.Reasoning, rule.Name)
				}
			}

			return result
		}
	}

	return nil
}

// evaluateCondition evaluates a condition expression
func (r *PolicyResolver) evaluateCondition(condition string, email *types.Email, results []*types.ClassificationResponse) bool {
	// Simple condition evaluation - in production, use a proper expression evaluator
	
	// Security override: check for high phishing scores
	if strings.Contains(condition, "phishing_score >= 0.8") {
		for _, result := range results {
			if phishingScore, exists := result.Metadata["phishing_score"]; exists {
				if score, ok := phishingScore.(float64); ok && score >= 0.8 {
					return true
				}
			}
		}
	}

	// Importance override: check for critical importance
	if strings.Contains(condition, "importance == 'critical'") {
		for _, result := range results {
			if importance, exists := result.Metadata["importance"]; exists {
				if imp, ok := importance.(string); ok && imp == "critical" && result.Confidence >= 0.7 {
					return true
				}
			}
		}
	}

	// Trusted sender boost: check sender reputation
	if strings.Contains(condition, "sender_reputation.trust_score >= 0.9") {
		if trustScore, exists := email.Headers["X-Sender-Trust-Score"]; exists {
			// Parse trust score from header (simplified)
			if strings.Contains(trustScore, "0.9") || strings.Contains(trustScore, "1.0") {
				return true
			}
		}
	}

	return false
}

// applyConfidenceWeighting applies confidence weighting to results
func (r *PolicyResolver) applyConfidenceWeighting(results []*types.ClassificationResponse) []*types.ClassificationResponse {
	weightedResults := make([]*types.ClassificationResponse, len(results))

	for i, result := range results {
		// Copy result
		weighted := *result
		
		// Apply profile weight
		if weight, exists := r.config.ConfidenceWeighting.ProfileWeights[result.ProfileID]; exists {
			weighted.Confidence = min(1.0, result.Confidence*weight)
			r.logger.WithFields(logrus.Fields{
				"profile_id":        result.ProfileID,
				"original_conf":     result.Confidence,
				"weight":           weight,
				"weighted_conf":    weighted.Confidence,
			}).Debug("Applied confidence weighting")
		}
		
		weightedResults[i] = &weighted
	}

	return weightedResults
}

// resolveConflicts resolves conflicts using the configured method
func (r *PolicyResolver) resolveConflicts(results []*types.ClassificationResponse) *types.ClassificationResponse {
	switch r.config.ConfidenceWeighting.Method {
	case "highest_confidence":
		return r.resolveByHighestConfidence(results)
	case "consensus":
		return r.resolveByConsensus(results)
	case "weighted_average":
		fallthrough
	default:
		return r.resolveByWeightedAverage(results)
	}
}

// resolveByHighestConfidence returns the result with highest confidence
func (r *PolicyResolver) resolveByHighestConfidence(results []*types.ClassificationResponse) *types.ClassificationResponse {
	var best *types.ClassificationResponse
	for _, result := range results {
		if best == nil || result.Confidence > best.Confidence {
			best = result
		}
	}
	return best
}

// resolveByConsensus finds consensus among results
func (r *PolicyResolver) resolveByConsensus(results []*types.ClassificationResponse) *types.ClassificationResponse {
	// Count actions
	actionCounts := make(map[string]int)
	actionResults := make(map[string][]*types.ClassificationResponse)
	
	for _, result := range results {
		actionCounts[result.Action]++
		actionResults[result.Action] = append(actionResults[result.Action], result)
	}
	
	// Find most common action
	var bestAction string
	var maxCount int
	for action, count := range actionCounts {
		if count > maxCount {
			maxCount = count
			bestAction = action
		}
	}
	
	// Return highest confidence result for the consensus action
	consensusResults := actionResults[bestAction]
	return r.resolveByHighestConfidence(consensusResults)
}

// resolveByWeightedAverage creates a weighted average result
func (r *PolicyResolver) resolveByWeightedAverage(results []*types.ClassificationResponse) *types.ClassificationResponse {
	// Group by action and calculate weighted averages
	actionGroups := make(map[string][]*types.ClassificationResponse)
	for _, result := range results {
		actionGroups[result.Action] = append(actionGroups[result.Action], result)
	}
	
	// Calculate weighted confidence for each action
	actionConfidences := make(map[string]float64)
	for action, group := range actionGroups {
		var totalWeight, weightedSum float64
		for _, result := range group {
			weight := 1.0 // Default weight
			if w, exists := r.config.ConfidenceWeighting.ProfileWeights[result.ProfileID]; exists {
				weight = w
			}
			totalWeight += weight
			weightedSum += result.Confidence * weight
		}
		actionConfidences[action] = weightedSum / totalWeight
	}
	
	// Find action with highest weighted confidence
	var bestAction string
	var bestConfidence float64
	for action, confidence := range actionConfidences {
		if confidence > bestConfidence {
			bestConfidence = confidence
			bestAction = action
		}
	}
	
	// Create combined result
	combinedResult := &types.ClassificationResponse{
		Action:      bestAction,
		Confidence:  bestConfidence,
		Reasoning:   r.combineReasonings(actionGroups[bestAction]),
		ProcessedAt: time.Now(),
	}
	
	// Combine labels from all results for this action
	labelSet := make(map[string]bool)
	for _, result := range actionGroups[bestAction] {
		for _, label := range result.Labels {
			labelSet[label] = true
		}
	}
	
	for label := range labelSet {
		combinedResult.Labels = append(combinedResult.Labels, label)
	}
	
	return combinedResult
}

// combineReasonings combines reasoning from multiple results
func (r *PolicyResolver) combineReasonings(results []*types.ClassificationResponse) string {
	var reasonings []string
	for _, result := range results {
		if result.Reasoning != "" {
			reasonings = append(reasonings, fmt.Sprintf("%s (conf: %.2f)", result.Reasoning, result.Confidence))
		}
	}
	return strings.Join(reasonings, "; ")
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
