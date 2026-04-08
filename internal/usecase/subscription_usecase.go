package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/ReilEgor/RepoNotifier/internal/domain/repository"
	"github.com/ReilEgor/RepoNotifier/internal/domain/service"
)

type SubscriptionUseCase struct {
	logger   *slog.Logger
	subsRepo repository.SubscriptionRepository
	userRepo repository.UserRepository
	repoRepo repository.RepositoryRepository
	ghClient service.GitHubClient
}

func NewSubscriptionUseCase(sr repository.SubscriptionRepository, gh service.GitHubClient, ur repository.UserRepository, rr repository.RepositoryRepository) *SubscriptionUseCase {
	return &SubscriptionUseCase{
		logger:   slog.With(slog.String("useCase", "SubscriptionUseCase")),
		subsRepo: sr,
		ghClient: gh,
		userRepo: ur,
		repoRepo: rr,
	}
}

func (uc *SubscriptionUseCase) Subscribe(ctx context.Context, email string, repoName string) (int64, error) {
	exists, err := uc.ghClient.RepoExists(ctx, repoName)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, service.ErrRepositoryNotFound
	}
	release, err := uc.ghClient.GetLatestRelease(ctx, repoName)
	if err != nil {
		return 0, err
	}
	user, err := uc.userRepo.GetOrCreate(ctx, email)
	if err != nil {
		return 0, err
	}
	repo, err := uc.repoRepo.GetOrCreate(ctx, repoName, release.TagName)
	if err != nil {
		return 0, err
	}
	sub := &model.Subscription{
		UserID:       user.ID,
		RepositoryID: repo.ID,
		LastSeenTag:  release.TagName,
		CreatedAt:    time.Now(),
	}
	return uc.subsRepo.Create(ctx, sub)
}

func (uc *SubscriptionUseCase) Unsubscribe(ctx context.Context, email string, repoName string) error {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get user: %w", err)
	}
	err = uc.subsRepo.Delete(ctx, user.ID, repoName)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	uc.logger.Info("user unsubscribed", "email", email, "repo", repoName)
	return nil
}

func (uc *SubscriptionUseCase) ListByEmail(ctx context.Context, email string) ([]model.Subscription, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return []model.Subscription{}, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	subs, err := uc.subsRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscriptions: %w", err)
	}

	return subs, nil
}

func (uc *SubscriptionUseCase) ProcessNotifications(ctx context.Context) error {
	return nil
}
