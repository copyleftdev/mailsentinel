package types

import (
	"time"
)

// Email represents a Gmail message with extracted content
type Email struct {
	ID          string            `json:"id"`
	ThreadID    string            `json:"thread_id"`
	Subject     string            `json:"subject"`
	From        string            `json:"from"`
	To          []string          `json:"to"`
	CC          []string          `json:"cc,omitempty"`
	BCC         []string          `json:"bcc,omitempty"`
	Date        time.Time         `json:"date"`
	Body        string            `json:"body"`
	BodyHTML    string            `json:"body_html,omitempty"`
	Labels      []string          `json:"labels"`
	Headers     map[string]string `json:"headers"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Size        int64             `json:"size"`
}

// Attachment represents an email attachment
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// ClassificationRequest represents a request to classify an email
type ClassificationRequest struct {
	Email     Email             `json:"email"`
	ProfileID string            `json:"profile_id"`
	Context   map[string]string `json:"context,omitempty"`
}

// ClassificationResponse represents the result of email classification
type ClassificationResponse struct {
	ProfileID   string                 `json:"profile_id"`
	Action      string                 `json:"action"`
	Confidence  float64                `json:"confidence"`
	Reasoning   string                 `json:"reasoning"`
	Labels      []string               `json:"labels,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// BatchRequest represents a batch of emails to process
type BatchRequest struct {
	Emails    []Email           `json:"emails"`
	ProfileID string            `json:"profile_id"`
	Context   map[string]string `json:"context,omitempty"`
	DryRun    bool              `json:"dry_run"`
}

// BatchResponse represents the results of batch processing
type BatchResponse struct {
	Results     []ClassificationResponse `json:"results"`
	Summary     BatchSummary             `json:"summary"`
	ProcessedAt time.Time                `json:"processed_at"`
	DryRun      bool                     `json:"dry_run"`
}

// BatchSummary provides aggregate statistics for batch processing
type BatchSummary struct {
	TotalEmails     int                    `json:"total_emails"`
	ProcessedEmails int                    `json:"processed_emails"`
	FailedEmails    int                    `json:"failed_emails"`
	ActionCounts    map[string]int         `json:"action_counts"`
	AvgConfidence   float64                `json:"avg_confidence"`
	ProcessingTime  time.Duration          `json:"processing_time"`
	Errors          []string               `json:"errors,omitempty"`
}
