package usecase

import (
	"context"
	"log/slog"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/ReilEgor/RepoNotifier/internal/domain/repository"
)

type UserUseCase struct {
	logger   *slog.Logger
	userRepo repository.UserRepository
}

func NewUserUseCase(ur repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		logger:   slog.With(slog.String("useCase", "UserUseCase")),
		userRepo: ur,
	}
}

func (uc *UserUseCase) GetByEmail(ctx context.Context, email string) (model.User, error) {
	panic("implement me")
}
func (uc *UserUseCase) Create(ctx context.Context, email string) (int, error) {
	panic("implement me")
}
