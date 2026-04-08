package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/jackc/pgx/v5"
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

const getByNameRepositoryQuery = `SELECT id, full_name, last_seen_tag, updated_at FROM repositories WHERE full_name = $1`

func (r *RepositoryRepository) GetByName(ctx context.Context, name string) (*model.Repository, error) {
	var repo model.Repository
	err := r.db.QueryRow(ctx, getByNameRepositoryQuery, name).Scan(
		&repo.ID,
		&repo.FullName,
		&repo.LastSeenTag,
		&repo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: repository not found", "RepositoryRepository.GetByName")
		}
		return nil, fmt.Errorf("%s: %w", "RepositoryRepository.GetByName", err)
	}

	return &repo, nil
}

const getAllRepositoryQuery = `SELECT id, full_name, last_seen_tag, updated_at FROM repositories`

func (r *RepositoryRepository) GetAll(ctx context.Context) ([]model.Repository, error) {
	rows, err := r.db.Query(ctx, getAllRepositoryQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "RepositoryRepository.GetAll", err)
	}
	defer rows.Close()

	var repos []model.Repository
	for rows.Next() {
		var repo model.Repository
		if err := rows.Scan(&repo.ID, &repo.FullName, &repo.LastSeenTag, &repo.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", "RepositoryRepository.GetAll", err)
		}
		repos = append(repos, repo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("RepositoryRepository.GetAll: rows error: %w", err)
	}

	return repos, nil
}

const updateLastSeenTagRepositoryQuery = `
	UPDATE repositories 
	SET last_seen_tag = $1, updated_at = CURRENT_TIMESTAMP 
	WHERE full_name = $2
`

func (r *RepositoryRepository) UpdateLastSeenTag(ctx context.Context, name, tag string) error {
	_, err := r.db.Exec(ctx, updateLastSeenTagRepositoryQuery, tag, name)
	if err != nil {
		return fmt.Errorf("%s: %w", "RepositoryRepository.UpdateLastSeenTag", err)
	}

	return nil
}

const getOrCreateRepositoryQuery = `
	INSERT INTO repositories (full_name, last_seen_tag)
	VALUES ($1, $2)
	ON CONFLICT (full_name) DO UPDATE SET full_name = EXCLUDED.full_name
	RETURNING id, full_name, last_seen_tag, updated_at
`

func (r *RepositoryRepository) GetOrCreate(ctx context.Context, name string, tagName string) (*model.Repository, error) {
	var repo model.Repository
	err := r.db.QueryRow(ctx, getOrCreateRepositoryQuery, name, tagName).Scan(
		&repo.ID,
		&repo.FullName,
		&repo.LastSeenTag,
		&repo.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "RepositoryRepository.GetOrCreate", err)
	}

	return &repo, nil
}
