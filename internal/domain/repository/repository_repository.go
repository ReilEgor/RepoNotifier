package repository

import (
	"context"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
)

//go:generate mockery --name RepositoryRepository --output ../../mocks --case underscore --outpkg mocks
type RepositoryRepository interface {
	Create(ctx context.Context, repo *model.Repository) error

	GetByName(ctx context.Context, name string) (*model.Repository, error)
	GetAll(ctx context.Context) ([]model.Repository, error)

	UpdateLastSeenTag(ctx context.Context, name, tag string) error
	GetOrCreate(ctx context.Context, name string, tagName string) (*model.Repository, error)
}
