package repository

import (
	"context"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, email, repo string) error

	GetByRepo(ctx context.Context, repo string) ([]model.Subscription, error)
	GetAll(ctx context.Context) ([]model.Subscription, error)
}
