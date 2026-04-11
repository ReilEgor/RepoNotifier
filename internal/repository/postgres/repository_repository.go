package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/jackc/pgx/v5"
)

const (
	componentRepositoryRepository = "RepositoryRepository"
)

// Sentinel errors.
var (
	ErrRepositoryNotFound = errors.New("repository not found")
)

type RepositoryRepository struct {
	db     PgxInterface
	logger *slog.Logger
}

func NewRepositoryRepository(db PgxInterface) *RepositoryRepository {
	return &RepositoryRepository{
		db:     db,
		logger: slog.With(slog.String("component", componentRepositoryRepository)),
	}
}
func (r *RepositoryRepository) Create(ctx context.Context, repo *model.Repository) error {
	panic("not implemented")
}

const getByNameRepositoryQuery = `SELECT id, full_name, last_seen_tag, updated_at FROM repositories WHERE full_name = $1`

func (r *RepositoryRepository) GetByName(ctx context.Context, name string) (*model.Repository, error) {
	const op = "RepositoryRepository.GetByName"
	log := r.logger.With(slog.String("op", op))

	var repo model.Repository
	err := r.db.QueryRow(ctx, getByNameRepositoryQuery, name).Scan(
		&repo.ID,
		&repo.FullName,
		&repo.LastSeenTag,
		&repo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.DebugContext(ctx, "repository not found", slog.String("op", op), slog.String("name", name))
			return nil, fmt.Errorf("%s: %w", op, ErrRepositoryNotFound)
		}
		log.ErrorContext(ctx, "query failed",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("%s: query row: %w", op, err)
	}

	return &repo, nil
}

const getAllRepositoryQuery = `SELECT id, full_name, last_seen_tag, updated_at FROM repositories`

func (r *RepositoryRepository) GetAll(ctx context.Context) ([]model.Repository, error) {
	const op = "RepositoryRepository.GetAll"
	log := r.logger.With(slog.String("op", op))

	rows, err := r.db.Query(ctx, getAllRepositoryQuery)
	if err != nil {
		log.ErrorContext(ctx, "query failed",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	repos := make([]model.Repository, 0)
	for rows.Next() {
		var repo model.Repository
		if err := rows.Scan(&repo.ID, &repo.FullName, &repo.LastSeenTag, &repo.UpdatedAt); err != nil {
			log.ErrorContext(ctx, "scan failed",
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		repos = append(repos, repo)
	}

	if err := rows.Err(); err != nil {
		log.ErrorContext(ctx, "rows iteration failed",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}
	log.DebugContext(ctx, "repositories fetched", slog.String("op", op), slog.Int("count", len(repos)))
	return repos, nil
}

const updateLastSeenTagRepositoryQuery = `
	UPDATE repositories 
	SET last_seen_tag = $1, updated_at = CURRENT_TIMESTAMP 
	WHERE full_name = $2
`

func (r *RepositoryRepository) UpdateLastSeenTag(ctx context.Context, name, tag string) error {
	const op = "RepositoryRepository.UpdateLastSeenTag"
	log := r.logger.With(slog.String("op", op))

	if _, err := r.db.Exec(ctx, updateLastSeenTagRepositoryQuery, tag, name); err != nil {
		log.ErrorContext(ctx, "exec failed",
			slog.String("name", name),
			slog.String("tag", tag),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%s: exec: %w", op, err)
	}
	log.DebugContext(ctx, "last seen tag updated",
		slog.String("name", name),
		slog.String("tag", tag),
	)
	return nil
}

const getOrCreateRepositoryQuery = `
	INSERT INTO repositories (full_name, last_seen_tag)
	VALUES ($1, $2)
	ON CONFLICT (full_name) DO UPDATE SET full_name = EXCLUDED.full_name
	RETURNING id, full_name, last_seen_tag, updated_at
`

func (r *RepositoryRepository) GetOrCreate(ctx context.Context, name string, tagName string) (*model.Repository, error) {
	const op = "RepositoryRepository.GetOrCreate"
	log := r.logger.With(slog.String("op", op))

	var repo model.Repository
	err := r.db.QueryRow(ctx, getOrCreateRepositoryQuery, name, tagName).Scan(
		&repo.ID,
		&repo.FullName,
		&repo.LastSeenTag,
		&repo.UpdatedAt,
	)
	if err != nil {
		log.ErrorContext(ctx, "query failed",
			slog.String("name", name),
			slog.String("tag", tagName),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("%s: query row: %w", op, err)
	}

	log.DebugContext(ctx, "repository get or created",
		slog.String("name", name),
		slog.Int64("id", repo.ID),
	)
	return &repo, nil
}
