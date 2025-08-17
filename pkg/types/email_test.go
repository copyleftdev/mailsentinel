package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEmail_Validation(t *testing.T) {
	tests := []struct {
		name  string
		email Email
		valid bool
	}{
		{
			name: "valid_email",
			email: Email{
				ID:       "test123",
				ThreadID: "thread123",
				Subject:  "Test Subject",
				From:     "sender@example.com",
				To:       []string{"recipient@example.com"},
				Date:     time.Now(),
				Body:     "Test email body",
				Size:     1024,
			},
			valid: true,
		},
		{
			name: "empty_id",
			email: Email{
				Subject: "Test Subject",
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Date:    time.Now(),
				Body:    "Test email body",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.email.ID, "Valid email should have ID")
			} else {
				assert.Empty(t, tt.email.ID, "Invalid email should have empty ID")
			}
		})
	}
}

func TestBatchSummary_Calculations(t *testing.T) {
	summary := BatchSummary{
		TotalEmails:     100,
		ProcessedEmails: 95,
		FailedEmails:    5,
		ActionCounts: map[string]int{
			"archive": 50,
			"star":    25,
			"delete":  15,
			"none":    5,
		},
		AvgConfidence:  0.85,
		ProcessingTime: 30 * time.Second,
	}

	assert.Equal(t, 100, summary.TotalEmails)
	assert.Equal(t, 95, summary.ProcessedEmails)
	assert.Equal(t, 5, summary.FailedEmails)
	assert.Equal(t, 0.85, summary.AvgConfidence)
	assert.Equal(t, 30*time.Second, summary.ProcessingTime)

	// Verify action counts sum correctly
	totalActions := 0
	for _, count := range summary.ActionCounts {
		totalActions += count
	}
	assert.Equal(t, 95, totalActions, "Action counts should sum to processed emails")
}

func TestClassificationResponse_Validation(t *testing.T) {
	response := ClassificationResponse{
		ProfileID:   "spam",
		Action:      "archive",
		Confidence:  0.85,
		Reasoning:   "High spam probability detected",
		Labels:      []string{"spam", "automated"},
		ProcessedAt: time.Now(),
	}

	assert.Equal(t, "spam", response.ProfileID)
	assert.Equal(t, "archive", response.Action)
	assert.Equal(t, 0.85, response.Confidence)
	assert.NotEmpty(t, response.Reasoning)
	assert.Len(t, response.Labels, 2)
	assert.False(t, response.ProcessedAt.IsZero())
}

func TestAttachment_Structure(t *testing.T) {
	attachment := Attachment{
		ID:       "att123",
		Filename: "document.pdf",
		MimeType: "application/pdf",
		Size:     1024000,
	}

	assert.Equal(t, "att123", attachment.ID)
	assert.Equal(t, "document.pdf", attachment.Filename)
	assert.Equal(t, "application/pdf", attachment.MimeType)
	assert.Equal(t, int64(1024000), attachment.Size)
}
