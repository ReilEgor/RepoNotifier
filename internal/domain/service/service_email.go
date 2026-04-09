package service

import (
	"context"
	"errors"
)

var (
	ErrEmailRateLimit   = errors.New("email provider rate limit reached")
	ErrInvalidRecipient = errors.New("invalid recipient email address")
	ErrSMTPUnavailable  = errors.New("smtp server is unreachable")
	ErrAuthFailed       = errors.New("email service authentication failed")
)

type EmailSender interface {
	SendNotification(ctx context.Context, to, subject, body string) error
}
