package usecase

import (
	"context"
	"errors"
)

var (
	ErrRepoNotFound      = errors.New("repository not found on GitHub")
	ErrAlreadySubscribed = errors.New("user already subscribed to this repository")
	ErrInvalidFormat     = errors.New("invalid repository format")
)

type SubscriptionUsecase interface {
	Subscribe(ctx context.Context, email string, repoName string) error
	Unsubscribe(ctx context.Context, email string, repoName string) error
	ProcessNotifications(ctx context.Context) error
}
