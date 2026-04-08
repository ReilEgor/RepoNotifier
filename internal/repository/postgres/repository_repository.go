package postgres

import (
	"context"
	"log/slog"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewRepositoryRepository(db *pgxpool.Pool) *RepositoryRepository {
	return &RepositoryRepository{
		db:     db,
		logger: slog.With(slog.String("component", "RepositoryRepository")),
	}
}
func (r *RepositoryRepository) Create(ctx context.Context, repo *model.Repository) error {
	panic("not implemented")
}

func (r *RepositoryRepository) GetByName(ctx context.Context, name string) (*model.Repository, error) {
	panic("not implemented")
}
func (r *RepositoryRepository) GetAll(ctx context.Context) ([]model.Repository, error) {
	panic("not implemented")
}

func (r *RepositoryRepository) UpdateLastSeenTag(ctx context.Context, name, tag string) error {
	panic("not implemented")
}
func (r *RepositoryRepository) GetOrCreate(ctx context.Context, name string, tagName string) (*model.Repository, error) {
	panic("not implemented")
}
