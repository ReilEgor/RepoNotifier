package usecase

import (
	"context"
	"errors"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
)

var (
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrInternalServer    = errors.New("internal server error")
)

type UserUseCase interface {
	GetByEmail(ctx context.Context, email string) (model.User, error)
	Create(ctx context.Context, email string) (int, error)
}
