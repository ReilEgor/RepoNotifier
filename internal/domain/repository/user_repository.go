package repository

import (
	"context"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (model.User, error)
	Create(ctx context.Context, email string) (int64, error)
	GetOrCreate(ctx context.Context, email string) (model.User, error)
}
