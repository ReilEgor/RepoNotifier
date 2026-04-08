package postgres

import (
	"context"
	"log/slog"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{
		db:     db,
		logger: slog.With(slog.String("component", "SubscriptionRepository")),
	}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) (int64, error) {
	panic("not implemented")
}
func (r *SubscriptionRepository) Delete(ctx context.Context, userID int64, repo string) error {
	panic("not implemented")
}

func (r *SubscriptionRepository) GetByRepo(ctx context.Context, repo string) ([]model.Subscription, error) {
	panic("not implemented")
}
func (r *SubscriptionRepository) GetAll(ctx context.Context) ([]model.Subscription, error) {
	panic("not implemented")
}
func (r *SubscriptionRepository) GetByUserID(ctx context.Context, id int64) ([]model.Subscription, error) {
	panic("not implemented")
}
