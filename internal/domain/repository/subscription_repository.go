package repository

import (
	"context"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) (int64, error)
	Delete(ctx context.Context, userID int64, repo string) error

	GetByRepo(ctx context.Context, repo string) ([]model.Subscription, error)
	GetAll(ctx context.Context) ([]model.Subscription, error)
	GetByUserID(ctx context.Context, id int64) ([]model.Subscription, error)
	GetEmailsByRepoID(ctx context.Context, repoID int64) ([]string, error)
}
