package models

import "time"

// Email represents the structure for an email message.
type Email struct {
	ID            uint           `json:"id"`
	MessageID     string         `json:"messageId"`
	Subject       string         `json:"subject"`
	From          []EmailAddress `json:"from"`
	To            []EmailAddress `json:"to"`
	Cc            []EmailAddress `json:"cc"`
	Date          time.Time      `json:"date"`
	Snippet       string         `json:"snippet"`
	Body          string         `json:"body"`
	HTMLBody      string         `json:"htmlBody"`
	IsRead        bool           `json:"isRead"`
	HasAttachment bool           `json:"hasAttachment"`
	Attachments   []Attachment   `json:"attachments"`
}

// EmailAddress represents a single email address (name and address).
type EmailAddress struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename  string `json:"filename"`
	MimeType  string `json:"mimeType"`
	Size      int64  `json:"size"`
	ContentID string `json:"contentId"` // Used for inline images
}