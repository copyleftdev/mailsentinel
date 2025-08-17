package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"github.com/sirupsen/logrus"

	"github.com/mailsentinel/core/pkg/config"
	"github.com/mailsentinel/core/pkg/types"
)

// Client represents a Gmail API client with OAuth authentication
type Client struct {
	service *gmail.Service
	config  *config.GmailConfig
	logger  *logrus.Logger
}

// NewClient creates a new Gmail client with OAuth configuration
func NewClient(cfg *config.GmailConfig, logger *logrus.Logger) (*Client, error) {
	ctx := context.Background()
	
	// Create OAuth2 config
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       cfg.Scopes,
		Endpoint:     google.Endpoint,
	}
	
	// Get or refresh token
	token, err := getToken(cfg.TokenFile, oauthConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth token: %w", err)
	}
	
	// Create HTTP client with token
	httpClient := oauthConfig.Client(ctx, token)
	httpClient.Timeout = cfg.Timeout
	
	// Create Gmail service
	service, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}
	
	return &Client{
		service: service,
		config:  cfg,
		logger:  logger,
	}, nil
}

// getToken retrieves a token from file or initiates OAuth flow
func getToken(tokenFile string, config *oauth2.Config) (*oauth2.Token, error) {
	token, err := tokenFromFile(tokenFile)
	if err == nil {
		return token, nil
	}
	
	// Token doesn't exist, initiate OAuth flow
	return getTokenFromWeb(config, tokenFile)
}

// tokenFromFile retrieves a token from a local file
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// getTokenFromWeb initiates OAuth flow and saves token
func getTokenFromWeb(config *oauth2.Config, tokenFile string) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n", authURL)
	fmt.Print("Enter the authorization code: ")
	
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}
	
	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	
	// Save token to file
	if err := saveToken(tokenFile, token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}
	
	return token, nil
}

// saveToken saves a token to a file
func saveToken(path string, token *oauth2.Token) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache OAuth token: %w", err)
	}
	defer f.Close()
	
	return json.NewEncoder(f).Encode(token)
}

// ListEmails retrieves emails based on query parameters
func (c *Client) ListEmails(ctx context.Context, query string, maxResults int64) ([]*types.Email, error) {
	c.logger.WithFields(logrus.Fields{
		"query":       query,
		"max_results": maxResults,
	}).Info("Listing emails from Gmail")
	
	call := c.service.Users.Messages.List("me").Q(query)
	if maxResults > 0 {
		call = call.MaxResults(maxResults)
	}
	
	response, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	
	var emails []*types.Email
	for _, message := range response.Messages {
		email, err := c.GetEmail(ctx, message.Id)
		if err != nil {
			c.logger.WithError(err).WithField("message_id", message.Id).Warn("Failed to get email")
			continue
		}
		emails = append(emails, email)
	}
	
	return emails, nil
}

// GetEmail retrieves a single email by ID
func (c *Client) GetEmail(ctx context.Context, messageID string) (*types.Email, error) {
	message, err := c.service.Users.Messages.Get("me", messageID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	
	email := &types.Email{
		ID:       message.Id,
		ThreadID: message.ThreadId,
		Labels:   message.LabelIds,
		Headers:  make(map[string]string),
	}
	
	// Parse headers
	for _, header := range message.Payload.Headers {
		switch strings.ToLower(header.Name) {
		case "subject":
			email.Subject = header.Value
		case "from":
			email.From = header.Value
		case "to":
			email.To = strings.Split(header.Value, ",")
		case "cc":
			if header.Value != "" {
				email.CC = strings.Split(header.Value, ",")
			}
		case "date":
			if date, err := time.Parse(time.RFC1123Z, header.Value); err == nil {
				email.Date = date
			}
		}
		email.Headers[header.Name] = header.Value
	}
	
	// Extract body
	email.Body = extractBody(message.Payload)
	email.Size = message.SizeEstimate
	
	// Extract attachments
	email.Attachments = extractAttachments(message.Payload)
	
	return email, nil
}

// extractBody extracts plain text body from message payload
func extractBody(payload *gmail.MessagePart) string {
	if payload.Body != nil && payload.Body.Data != "" {
		return payload.Body.Data
	}
	
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			return part.Body.Data
		}
	}
	
	// Fallback to any text content
	for _, part := range payload.Parts {
		if strings.HasPrefix(part.MimeType, "text/") && part.Body != nil && part.Body.Data != "" {
			return part.Body.Data
		}
	}
	
	return ""
}

// extractAttachments extracts attachment information from message payload
func extractAttachments(payload *gmail.MessagePart) []types.Attachment {
	var attachments []types.Attachment
	
	for _, part := range payload.Parts {
		if part.Filename != "" && part.Body != nil && part.Body.AttachmentId != "" {
			attachments = append(attachments, types.Attachment{
				ID:       part.Body.AttachmentId,
				Filename: part.Filename,
				MimeType: part.MimeType,
				Size:     part.Body.Size,
			})
		}
	}
	
	return attachments
}

// ModifyLabels adds or removes labels from an email
func (c *Client) ModifyLabels(ctx context.Context, messageID string, addLabels, removeLabels []string) error {
	c.logger.WithFields(logrus.Fields{
		"message_id":    messageID,
		"add_labels":    addLabels,
		"remove_labels": removeLabels,
	}).Info("Modifying email labels")
	
	request := &gmail.ModifyMessageRequest{
		AddLabelIds:    addLabels,
		RemoveLabelIds: removeLabels,
	}
	
	_, err := c.service.Users.Messages.Modify("me", messageID, request).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to modify labels: %w", err)
	}
	
	return nil
}

// CreateLabel creates a new Gmail label
func (c *Client) CreateLabel(ctx context.Context, name string) (*gmail.Label, error) {
	c.logger.WithField("label_name", name).Info("Creating Gmail label")
	
	label := &gmail.Label{
		Name:                name,
		MessageListVisibility: "show",
		LabelListVisibility:   "labelShow",
	}
	
	createdLabel, err := c.service.Users.Labels.Create("me", label).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}
	
	return createdLabel, nil
}

// ListLabels retrieves all Gmail labels
func (c *Client) ListLabels(ctx context.Context) ([]*gmail.Label, error) {
	response, err := c.service.Users.Labels.List("me").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}
	
	return response.Labels, nil
}

// HealthCheck verifies Gmail API connectivity
func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.service.Users.GetProfile("me").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("Gmail health check failed: %w", err)
	}
	
	return nil
}
