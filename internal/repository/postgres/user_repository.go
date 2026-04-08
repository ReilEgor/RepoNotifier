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

const getByEmailQuery = `SELECT id, email, created_at FROM users WHERE email = $1`

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, getByEmailQuery, email).Scan(&user.ID, &user.Email, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, fmt.Errorf("%s: user not found", "UserRepository.GetByEmail")
		}
		return model.User{}, fmt.Errorf("%s: %w", "UserRepository.GetByEmail", err)
	}

	return user, nil
}

const createQuery = `INSERT INTO users (email) VALUES ($1) RETURNING id`

func (r *UserRepository) Create(ctx context.Context, email string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, createQuery, email).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", "UserRepository.Create", err)
	}

	return id, nil
}

const getOrCreateQuery = `
	INSERT INTO users (email)
	VALUES ($1)
	ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
	RETURNING id, email, created_at
`

func (r *UserRepository) GetOrCreate(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, getOrCreateQuery, email).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		return model.User{}, fmt.Errorf("%s: %w", "UserRepository.GetOrCreate", err)
	}

	return user, nil
}
