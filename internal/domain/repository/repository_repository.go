package repository

import (
	"context"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
)

type RepositoryRepository interface {
	Create(ctx context.Context, repo *model.Repository) error

	GetByName(ctx context.Context, name string) (*model.Repository, error)
	GetAll(ctx context.Context) ([]model.Repository, error)

	UpdateLastSeenTag(ctx context.Context, name, tag string) error
}
