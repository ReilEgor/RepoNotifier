package postgres

import (
	"context"
	"fmt"
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

const createSubscriptionQuery = `
	INSERT INTO subscriptions (user_id, repository_id)
	VALUES ($1, $2)
	ON CONFLICT (user_id, repository_id) DO UPDATE SET user_id = EXCLUDED.user_id
	RETURNING id
`

func (r *SubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, createSubscriptionQuery, sub.UserID, sub.RepositoryID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "SubscriptionRepository.Create", err)
	}

	return id, nil
}

const deleteSubscriptionQuery = `
	DELETE FROM subscriptions 
	WHERE user_id = $1 AND repository_id = (SELECT id FROM repositories WHERE full_name = $2)
`

func (r *SubscriptionRepository) Delete(ctx context.Context, userID int64, repo string) error {
	_, err := r.db.Exec(ctx, deleteSubscriptionQuery, userID, repo)
	if err != nil {
		return fmt.Errorf("%s: %w", "SubscriptionRepository.Delete", err)
	}

	return nil
}

func (r *SubscriptionRepository) GetByRepo(ctx context.Context, repo string) ([]model.Subscription, error) {
	panic("not implemented")
}

const getAllSubscriptionQuery = `SELECT id, user_id, repository_id, created_at FROM subscriptions`

func (r *SubscriptionRepository) GetAll(ctx context.Context) ([]model.Subscription, error) {
	rows, err := r.db.Query(ctx, getAllSubscriptionQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "SubscriptionRepository.GetAll", err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.RepositoryID, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", "SubscriptionRepository.GetAll", err)
		}
		subs = append(subs, s)
	}

	return subs, nil
}

const getByUserIDSubscriptionQuery = `SELECT id, user_id, repository_id, created_at FROM subscriptions WHERE user_id = $1`

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, id int64) ([]model.Subscription, error) {
	rows, err := r.db.Query(ctx, getByUserIDSubscriptionQuery, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "SubscriptionRepository.GetByUserID", err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.RepositoryID, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", "SubscriptionRepository.GetByUserID", err)
		}
		subs = append(subs, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", "SubscriptionRepository.GetByUserID", err)
	}

	return subs, nil
}

const getEmailsByRepoIDQuery = `
    SELECT u.email 
    FROM users u
    JOIN subscriptions s ON u.id = s.user_id
    WHERE s.repository_id = $1
`

func (r *SubscriptionRepository) GetEmailsByRepoID(ctx context.Context, repoID int64) ([]string, error) {
	rows, err := r.db.Query(ctx, getEmailsByRepoIDQuery, repoID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "SubscriptionRepository.GetEmailsByRepoID", err)
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", "SubscriptionRepository.GetEmailsByRepoID", err)
		}
		emails = append(emails, email)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", "SubscriptionRepository.GetEmailsByRepoID", err)
	}

	return emails, nil
}
