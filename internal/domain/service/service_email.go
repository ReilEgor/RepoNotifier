package service

import (
	"context"
	"errors"
)

var (
	ErrEmailRateLimit  = errors.New("email provider rate limit reached")
	ErrSMTPUnavailable = errors.New("email server is unreachable")
	ErrAuthFailed      = errors.New("email service authentication failed")
)

type EmailSender interface {
	SendNotification(ctx context.Context, to, subject, body string) error
}
