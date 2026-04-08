package postgres

import (
	"context"
	"log/slog"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: slog.With(slog.String("component", "UserRepository")),
	}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	panic("implement me")
}
func (r *UserRepository) Create(ctx context.Context, email string) (int64, error) {
	panic("implement me")
}
func (r *UserRepository) GetOrCreate(ctx context.Context, email string) (model.User, error) {
	panic("implement me")
}
