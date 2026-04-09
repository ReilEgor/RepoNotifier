package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	componentSubscriptionRepository = "SubscriptionRepository"
)

type SubscriptionRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{
		db:     db,
		logger: slog.With(slog.String("component", componentSubscriptionRepository)),
	}
}

const createSubscriptionQuery = `
	INSERT INTO subscriptions (user_id, repository_id)
	VALUES ($1, $2)
	ON CONFLICT (user_id, repository_id) DO UPDATE SET user_id = EXCLUDED.user_id
	RETURNING id
`

func (r *SubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) (int64, error) {
	const op = "SubscriptionRepository.Create"
	log := r.logger.With(slog.String("op", op))

	var id int64
	err := r.db.QueryRow(ctx, createSubscriptionQuery, sub.UserID, sub.RepositoryID).Scan(&id)
	if err != nil {
		log.ErrorContext(ctx, "failed to create subscription",
			slog.Int64("user_id", sub.UserID),
			slog.Int64("repo_id", sub.RepositoryID),
			slog.String("error", err.Error()),
		)
		return 0, fmt.Errorf("%s: query row: %w", op, err)
	}

	log.DebugContext(ctx, "subscription created", slog.Int64("id", id))
	return id, nil
}

const deleteSubscriptionQuery = `
	DELETE FROM subscriptions 
	WHERE user_id = $1 AND repository_id = (SELECT id FROM repositories WHERE full_name = $2)
`

func (r *SubscriptionRepository) Delete(ctx context.Context, userID int64, repo string) error {
	const op = "SubscriptionRepository.Delete"
	log := r.logger.With(slog.String("op", op))

	res, err := r.db.Exec(ctx, deleteSubscriptionQuery, userID, repo)
	if err != nil {
		log.ErrorContext(ctx, "failed to delete subscription",
			slog.Int64("user_id", userID),
			slog.String("repo", repo),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%s: exec: %w", op, err)
	}

	log.DebugContext(ctx, "subscription deleted",
		slog.Int64("user_id", userID),
		slog.String("repo", repo),
		slog.Int64("affected", res.RowsAffected()),
	)
	return nil
}

func (r *SubscriptionRepository) GetByRepo(ctx context.Context, repo string) ([]model.Subscription, error) {
	panic("not implemented")
}

const getAllSubscriptionQuery = `SELECT id, user_id, repository_id, created_at FROM subscriptions`

func (r *SubscriptionRepository) GetAll(ctx context.Context) ([]model.Subscription, error) {
	const op = "SubscriptionRepository.GetAll"
	log := r.logger.With(slog.String("op", op))

	rows, err := r.db.Query(ctx, getAllSubscriptionQuery)
	if err != nil {
		log.ErrorContext(ctx, "query failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.RepositoryID, &s.CreatedAt); err != nil {
			log.ErrorContext(ctx, "scan failed", slog.String("error", err.Error()))
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		subs = append(subs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}

	log.DebugContext(ctx, "subscriptions fetched", slog.Int("count", len(subs)))
	return subs, nil
}

const getByUserIDSubscriptionQuery = `SELECT id, user_id, repository_id, created_at FROM subscriptions WHERE user_id = $1`

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, id int64) ([]model.Subscription, error) {
	const op = "SubscriptionRepository.GetByUserID"
	log := r.logger.With(slog.String("op", op))

	rows, err := r.db.Query(ctx, getByUserIDSubscriptionQuery, id)
	if err != nil {
		log.ErrorContext(ctx, "query failed", slog.Int64("user_id", id), slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.RepositoryID, &s.CreatedAt); err != nil {
			log.ErrorContext(ctx, "scan failed", slog.String("error", err.Error()))
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		subs = append(subs, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}
	log.DebugContext(ctx, "user subscriptions fetched",
		slog.Int64("id", id),
		slog.Int("count", len(subs)))
	return subs, nil
}

const getEmailsByRepoIDQuery = `
    SELECT u.email 
    FROM users u
    JOIN subscriptions s ON u.id = s.user_id
    WHERE s.repository_id = $1
`

func (r *SubscriptionRepository) GetEmailsByRepoID(ctx context.Context, repoID int64) ([]string, error) {
	const op = "SubscriptionRepository.GetEmailsByRepoID"
	log := r.logger.With(slog.String("op", op))

	rows, err := r.db.Query(ctx, getEmailsByRepoIDQuery, repoID)
	if err != nil {
		log.ErrorContext(ctx, "query failed", slog.Int64("repo_id", repoID), slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			log.ErrorContext(ctx, "scan failed", slog.String("error", err.Error()))
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		emails = append(emails, email)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}
	log.DebugContext(ctx, "user emails fetched",
		slog.Int64("repo_id", repoID),
		slog.Int("count", len(emails)))
	return emails, nil
}
