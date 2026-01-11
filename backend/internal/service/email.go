package service

import (
	"context"
	"fmt"
)

// EmailMessage represents an email to be sent.
type EmailMessage struct {
	To      string
	Subject string
	HTML    string
	Text    string // Plain text fallback
}

// EmailSender defines the interface for sending emails.
// Implementations can use SendGrid, SES, SMTP, etc.
type EmailSender interface {
	Send(ctx context.Context, msg *EmailMessage) error
}

// EmailTemplates contains email template helpers.
type EmailTemplates struct {
	BaseURL string
}

// NewEmailTemplates creates a new email templates instance.
func NewEmailTemplates(baseURL string) *EmailTemplates {
	return &EmailTemplates{BaseURL: baseURL}
}

// EmailVerificationData holds data for email verification template.
type EmailVerificationData struct {
	Email           string
	VerificationURL string
	ExpiresIn       string
}

// PasswordResetData holds data for password reset template.
type PasswordResetData struct {
	Email     string
	ResetURL  string
	ExpiresIn string
}

// BuildVerificationEmail builds an email verification message.
func (t *EmailTemplates) BuildVerificationEmail(data EmailVerificationData) *EmailMessage {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Verify your email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2563eb;">Verify your email</h1>
        <p>Please click the button below to verify your email address:</p>
        <p style="margin: 30px 0;">
            <a href="%s" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">
                Verify Email
            </a>
        </p>
        <p style="color: #666; font-size: 14px;">
            This link expires in %s.
        </p>
        <p style="color: #666; font-size: 14px;">
            If you didn't create an account, you can safely ignore this email.
        </p>
    </div>
</body>
</html>`, data.VerificationURL, data.ExpiresIn)

	text := fmt.Sprintf(`Verify your email

Please click the link below to verify your email address:

%s

This link expires in %s.

If you didn't create an account, you can safely ignore this email.`,
		data.VerificationURL, data.ExpiresIn)

	return &EmailMessage{
		To:      data.Email,
		Subject: "Verify your email address",
		HTML:    html,
		Text:    text,
	}
}

// BuildPasswordResetEmail builds a password reset message.
func (t *EmailTemplates) BuildPasswordResetEmail(data PasswordResetData) *EmailMessage {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Reset your password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2563eb;">Reset your password</h1>
        <p>We received a request to reset your password. Click the button below to create a new password:</p>
        <p style="margin: 30px 0;">
            <a href="%s" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">
                Reset Password
            </a>
        </p>
        <p style="color: #666; font-size: 14px;">
            This link expires in %s.
        </p>
        <p style="color: #666; font-size: 14px;">
            If you didn't request a password reset, please ignore this email. Your password will remain unchanged.
        </p>
    </div>
</body>
</html>`, data.ResetURL, data.ExpiresIn)

	text := fmt.Sprintf(`Reset your password

We received a request to reset your password. Click the link below to create a new password:

%s

This link expires in %s.

If you didn't request a password reset, please ignore this email. Your password will remain unchanged.`,
		data.ResetURL, data.ExpiresIn)

	return &EmailMessage{
		To:      data.Email,
		Subject: "Reset your password",
		HTML:    html,
		Text:    text,
	}
}
