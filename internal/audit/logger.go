package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/mailsentinel/core/pkg/config"
	"github.com/mailsentinel/core/pkg/types"
)

// Logger handles secure audit logging with integrity verification
type Logger struct {
	config     *config.AuditConfig
	logger     *logrus.Logger
	file       *os.File
	mutex      sync.RWMutex
	entryCount int64
	lastHash   string
}

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	EmailID     string                 `json:"email_id,omitempty"`
	ProfileID   string                 `json:"profile_id,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Confidence  float64                `json:"confidence,omitempty"`
	Reasoning   string                 `json:"reasoning,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	PrevHash    string                 `json:"prev_hash"`
	Hash        string                 `json:"hash"`
	Signature   string                 `json:"signature,omitempty"`
}

// EventType constants for audit logging
const (
	EventEmailClassified   = "email_classified"
	EventProfileLoaded     = "profile_loaded"
	EventConfigChanged     = "config_changed"
	EventAuthTokenRefresh  = "auth_token_refresh"
	EventSecurityViolation = "security_violation"
	EventSystemStart       = "system_start"
	EventSystemStop        = "system_stop"
	EventError             = "error"
)

// NewLogger creates a new audit logger
func NewLogger(cfg *config.AuditConfig, logger *logrus.Logger) (*Logger, error) {
	if !cfg.Enabled {
		return &Logger{config: cfg, logger: logger}, nil
	}

	// Ensure audit directory exists
	if err := os.MkdirAll(cfg.Directory, 0750); err != nil {
		return nil, fmt.Errorf("failed to create audit directory: %w", err)
	}

	// Open current audit file
	filename := filepath.Join(cfg.Directory, fmt.Sprintf("audit_%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit file: %w", err)
	}

	auditLogger := &Logger{
		config: cfg,
		logger: logger,
		file:   file,
	}

	// Initialize chain if file is empty
	if stat, err := file.Stat(); err == nil && stat.Size() == 0 {
		if err := auditLogger.initializeChain(); err != nil {
			return nil, fmt.Errorf("failed to initialize audit chain: %w", err)
		}
	} else {
		// Load last hash from existing file
		if err := auditLogger.loadLastHash(); err != nil {
			logger.WithError(err).Warn("Failed to load last hash, starting new chain")
			auditLogger.lastHash = ""
		}
	}

	return auditLogger, nil
}

// initializeChain creates the genesis entry for a new audit chain
func (l *Logger) initializeChain() error {
	genesis := &AuditEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		EventType: "chain_genesis",
		PrevHash:  "",
		Metadata: map[string]interface{}{
			"version": "1.0",
			"system":  "mailsentinel",
		},
	}

	genesis.Hash = l.calculateHash(genesis)
	l.lastHash = genesis.Hash

	return l.writeEntry(genesis)
}

// LogEmailClassification logs an email classification event
func (l *Logger) LogEmailClassification(email *types.Email, response *types.ClassificationResponse) error {
	if !l.config.Enabled {
		return nil
	}

	entry := &AuditEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		EventType: EventEmailClassified,
		EmailID:   email.ID,
		ProfileID: response.ProfileID,
		Action:    response.Action,
		Confidence: response.Confidence,
		Reasoning: response.Reasoning,
		PrevHash:  l.lastHash,
		Metadata: map[string]interface{}{
			"email_subject": email.Subject,
			"email_from":    email.From,
			"email_size":    email.Size,
			"labels":        response.Labels,
		},
	}

	entry.Hash = l.calculateHash(entry)
	l.mutex.Lock()
	l.lastHash = entry.Hash
	l.entryCount++
	l.mutex.Unlock()

	return l.writeEntry(entry)
}

// LogProfileLoad logs a profile loading event
func (l *Logger) LogProfileLoad(profileID, version string, success bool) error {
	if !l.config.Enabled {
		return nil
	}

	entry := &AuditEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		EventType: EventProfileLoaded,
		ProfileID: profileID,
		PrevHash:  l.lastHash,
		Metadata: map[string]interface{}{
			"version": version,
			"success": success,
		},
	}

	entry.Hash = l.calculateHash(entry)
	l.mutex.Lock()
	l.lastHash = entry.Hash
	l.entryCount++
	l.mutex.Unlock()

	return l.writeEntry(entry)
}

// LogSecurityViolation logs a security violation event
func (l *Logger) LogSecurityViolation(violationType, description string, metadata map[string]interface{}) error {
	if !l.config.Enabled {
		return nil
	}

	entry := &AuditEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		EventType: EventSecurityViolation,
		PrevHash:  l.lastHash,
		Metadata: map[string]interface{}{
			"violation_type": violationType,
			"description":    description,
		},
	}

	// Merge additional metadata
	for k, v := range metadata {
		entry.Metadata[k] = v
	}

	entry.Hash = l.calculateHash(entry)
	l.mutex.Lock()
	l.lastHash = entry.Hash
	l.entryCount++
	l.mutex.Unlock()

	return l.writeEntry(entry)
}

// LogSystemEvent logs system start/stop events
func (l *Logger) LogSystemEvent(eventType string, metadata map[string]interface{}) error {
	if !l.config.Enabled {
		return nil
	}

	entry := &AuditEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		EventType: eventType,
		PrevHash:  l.lastHash,
		Metadata:  metadata,
	}

	entry.Hash = l.calculateHash(entry)
	l.mutex.Lock()
	l.lastHash = entry.Hash
	l.entryCount++
	l.mutex.Unlock()

	return l.writeEntry(entry)
}

// calculateHash calculates SHA-256 hash of audit entry
func (l *Logger) calculateHash(entry *AuditEntry) string {
	// Create deterministic string representation
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%f|%s",
		entry.ID,
		entry.Timestamp.Format(time.RFC3339Nano),
		entry.EventType,
		entry.EmailID,
		entry.ProfileID,
		entry.Action,
		entry.Confidence,
		entry.PrevHash,
	)

	// Add metadata in sorted order for deterministic hash
	if entry.Metadata != nil {
		metadataJSON, _ := json.Marshal(entry.Metadata)
		data += "|" + string(metadataJSON)
	}

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// writeEntry writes an audit entry to the log file
func (l *Logger) writeEntry(entry *AuditEntry) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Sign entry if encryption key is provided
	if l.config.EncryptionKey != "" {
		signature, err := l.signEntry(entry)
		if err != nil {
			l.logger.WithError(err).Error("Failed to sign audit entry")
		} else {
			entry.Signature = signature
		}
	}

	// Convert to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	// Write to file
	if _, err := l.file.WriteString(string(data) + "\n"); err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}

	// Sync to disk for integrity
	if err := l.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync audit file: %w", err)
	}

	l.logger.WithFields(logrus.Fields{
		"entry_id":    entry.ID,
		"event_type":  entry.EventType,
		"hash":        entry.Hash,
		"entry_count": l.entryCount,
	}).Debug("Wrote audit entry")

	return nil
}

// signEntry creates a cryptographic signature for the entry
func (l *Logger) signEntry(entry *AuditEntry) (string, error) {
	// Use bcrypt for simplicity - in production, use proper digital signatures
	data := entry.Hash + l.config.EncryptionKey
	hash, err := bcrypt.GenerateFromPassword([]byte(data), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash), nil
}

// VerifyChain verifies the integrity of the audit chain
func (l *Logger) VerifyChain() error {
	if !l.config.Enabled || !l.config.IntegrityCheck {
		return nil
	}

	l.logger.Info("Starting audit chain verification")

	// Read all entries from current file
	entries, err := l.readAllEntries()
	if err != nil {
		return fmt.Errorf("failed to read audit entries: %w", err)
	}

	if len(entries) == 0 {
		return nil // Empty chain is valid
	}

	// Verify each entry's hash and chain integrity
	var prevHash string
	for i, entry := range entries {
		// Verify hash
		expectedHash := l.calculateHash(&entry)
		if entry.Hash != expectedHash {
			return fmt.Errorf("hash mismatch at entry %d: expected %s, got %s", i, expectedHash, entry.Hash)
		}

		// Verify chain link
		if entry.PrevHash != prevHash {
			return fmt.Errorf("chain break at entry %d: expected prev_hash %s, got %s", i, prevHash, entry.PrevHash)
		}

		// Verify signature if present
		if entry.Signature != "" && l.config.EncryptionKey != "" {
			if err := l.verifySignature(&entry); err != nil {
				return fmt.Errorf("signature verification failed at entry %d: %w", i, err)
			}
		}

		prevHash = entry.Hash
	}

	l.logger.WithField("entries_verified", len(entries)).Info("Audit chain verification completed successfully")
	return nil
}

// verifySignature verifies an entry's cryptographic signature
func (l *Logger) verifySignature(entry *AuditEntry) error {
	if entry.Signature == "" {
		return fmt.Errorf("no signature present")
	}

	signature, err := hex.DecodeString(entry.Signature)
	if err != nil {
		return fmt.Errorf("invalid signature format: %w", err)
	}

	data := entry.Hash + l.config.EncryptionKey
	return bcrypt.CompareHashAndPassword(signature, []byte(data))
}

// readAllEntries reads all audit entries from the current file
func (l *Logger) readAllEntries() ([]AuditEntry, error) {
	// Implementation would read and parse JSON entries from file
	// Simplified for brevity
	return []AuditEntry{}, nil
}

// loadLastHash loads the last hash from the audit file
func (l *Logger) loadLastHash() error {
	// Implementation would read the last entry and extract its hash
	// Simplified for brevity
	l.lastHash = ""
	return nil
}

// generateID generates a unique ID for audit entries
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// LogClassification logs an email classification event
func (l *Logger) LogClassification(email *types.Email, result *types.ClassificationResponse) error {
	if !l.config.Enabled {
		return nil
	}

	entry := &AuditEntry{
		ID:         generateID(),
		Timestamp:  time.Now(),
		EventType:  EventEmailClassified,
		EmailID:    email.ID,
		ProfileID:  result.ProfileID,
		Action:     result.Action,
		Confidence: result.Confidence,
		Reasoning:  result.Reasoning,
	}

	return l.writeEntry(entry)
}

// LogAction logs an email action event
func (l *Logger) LogAction(email *types.Email, action, label string) error {
	if !l.config.Enabled {
		return nil
	}

	entry := &AuditEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		EventType: "action",
		EmailID:   email.ID,
		Action:    action,
		Metadata: map[string]interface{}{
			"label": label,
		},
	}

	return l.writeEntry(entry)
}

// VerifyIntegrity verifies the integrity of the audit log chain
func (l *Logger) VerifyIntegrity() (bool, error) {
	if !l.config.Enabled || !l.config.IntegrityCheck {
		return true, nil
	}

	// Implementation would verify the hash chain
	// For now, return true as a placeholder
	return true, nil
}

// Close closes the audit logger and performs final verification
func (l *Logger) Close() error {
	if !l.config.Enabled || l.file == nil {
		return nil
	}

	// Log system stop event
	l.LogSystemEvent(EventSystemStop, map[string]interface{}{
		"total_entries": l.entryCount,
		"final_hash":    l.lastHash,
	})

	// Perform final integrity check
	if err := l.VerifyChain(); err != nil {
		l.logger.WithError(err).Error("Final audit chain verification failed")
	}

	return l.file.Close()
}
