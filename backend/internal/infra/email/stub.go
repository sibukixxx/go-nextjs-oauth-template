// Package email provides email sending implementations.
package email

import (
	"context"
	"log/slog"

	"github.com/your-org/go-nextjs-oauth-template/backend/internal/service"
)

// StubSender logs emails instead of sending them.
// Use this for development and testing.
type StubSender struct {
	logger *slog.Logger
}

// NewStubSender creates a new stub email sender.
func NewStubSender(logger *slog.Logger) *StubSender {
	if logger == nil {
		logger = slog.Default()
	}
	return &StubSender{logger: logger}
}

// Send logs the email message instead of actually sending it.
func (s *StubSender) Send(ctx context.Context, msg *service.EmailMessage) error {
	// Truncate text preview for logging
	textPreview := msg.Text
	if len(textPreview) > 200 {
		textPreview = textPreview[:200] + "..."
	}

	s.logger.Info("email would be sent (stub)",
		"to", msg.To,
		"subject", msg.Subject,
		"text_preview", textPreview,
	)
	return nil
}

// Ensure StubSender implements EmailSender interface
var _ service.EmailSender = (*StubSender)(nil)
