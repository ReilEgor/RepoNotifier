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

const (
	componentSubscriptionUseCase = "SubscriptionUseCase"

	repoCheckTimeout = 10 * time.Second
)

const (
	errMsgGetUser         = "get user"
	errMsgDeleteSub       = "delete subscription"
	errMsgGetRepos        = "get repos"
	errMsgFetchRelease    = "fetch latest release"
	errMsgUpdateTag       = "update last seen tag"
	errMsgGetSubscribers  = "get subscribers"
	errMsgGetOrCreateUser = "get or create user"
	errMsgGetOrCreateRepo = "get or create repo"
	errMsgCreateSub       = "create subscription"
	errMsgCheckRepoExists = "check repo exists"
)

type SubscriptionUseCase struct {
	logger      *slog.Logger
	subsRepo    repository.SubscriptionRepository
	userRepo    repository.UserRepository
	repoRepo    repository.RepositoryRepository
	emailSender service.EmailSender
	ghClient    service.GitHubClient
}

func NewSubscriptionUseCase(
	sr repository.SubscriptionRepository,
	gh service.GitHubClient,
	ur repository.UserRepository,
	rr repository.RepositoryRepository,
	es service.EmailSender,
) *SubscriptionUseCase {
	return &SubscriptionUseCase{
		logger:      slog.With(slog.String("useCase", componentSubscriptionUseCase)),
		subsRepo:    sr,
		ghClient:    gh,
		userRepo:    ur,
		repoRepo:    rr,
		emailSender: es,
	}
}

func (uc *SubscriptionUseCase) Subscribe(ctx context.Context, email string, repoName string) (int64, error) {
	const op = "SubscriptionUseCase.Subscribe"
	log := uc.logger.With(slog.String("op", op), slog.String("email", email), slog.String("repo", repoName))

	exists, err := uc.ghClient.RepoExists(ctx, repoName)
	if err != nil {
		log.ErrorContext(ctx, errMsgCheckRepoExists, slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %s: %w", op, errMsgCheckRepoExists, err)
	}
	if !exists {
		log.WarnContext(ctx, "repository not found on github")
		return 0, service.ErrRepositoryNotFound
	}
	release, err := uc.ghClient.GetLatestRelease(ctx, repoName)
	if err != nil {
		log.ErrorContext(ctx, errMsgFetchRelease, slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %s: %w", op, errMsgFetchRelease, err)
	}
	user, err := uc.userRepo.GetOrCreate(ctx, email)
	if err != nil {
		log.ErrorContext(ctx, errMsgGetOrCreateUser, slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %s: %w", op, errMsgGetOrCreateUser, err)
	}
	repo, err := uc.repoRepo.GetOrCreate(ctx, repoName, release.TagName)
	if err != nil {
		log.ErrorContext(ctx, errMsgGetOrCreateRepo, slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %s: %w", op, errMsgGetOrCreateRepo, err)
	}
	sub := &model.Subscription{
		UserID:       user.ID,
		RepositoryID: repo.ID,
		LastSeenTag:  release.TagName,
		CreatedAt:    time.Now(),
	}

	id, err := uc.subsRepo.Create(ctx, sub)
	if err != nil {
		log.ErrorContext(ctx, errMsgCreateSub, slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %s: %w", op, errMsgCreateSub, err)
	}

	log.InfoContext(ctx, "subscribed successfully", slog.Int64("subscription_id", id), slog.String("tag", release.TagName))
	return id, nil
}

func (uc *SubscriptionUseCase) Unsubscribe(ctx context.Context, email string, repoName string) error {
	const op = "SubscriptionUseCase.Unsubscribe"
	log := uc.logger.With(slog.String("op", op), slog.String("email", email), slog.String("repo", repoName))

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			log.DebugContext(ctx, "user not found, nothing to unsubscribe")
			return nil
		}
		log.ErrorContext(ctx, errMsgGetUser, slog.String("error", err.Error()))
		return fmt.Errorf("%s: %s: %w", op, errMsgGetUser, err)
	}

	if err = uc.subsRepo.Delete(ctx, user.ID, repoName); err != nil {
		log.ErrorContext(ctx, errMsgDeleteSub, slog.String("error", err.Error()))
		return fmt.Errorf("%s: %s: %w", op, errMsgDeleteSub, err)
	}

	log.InfoContext(ctx, "unsubscribed successfully")
	return nil
}

func (uc *SubscriptionUseCase) ListByEmail(ctx context.Context, email string) ([]model.Subscription, error) {
	const op = "SubscriptionUseCase.ListByEmail"
	log := uc.logger.With(slog.String("op", op), slog.String("email", email))

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			log.DebugContext(ctx, "user not found, returning empty list")
			return []model.Subscription{}, nil
		}
		log.ErrorContext(ctx, errMsgGetUser, slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %s: %w", op, errMsgGetUser, err)
	}

	subs, err := uc.subsRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		log.ErrorContext(ctx, "get subscriptions by user id", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: get subscriptions: %w", op, err)
	}

	log.DebugContext(ctx, "subscriptions listed", slog.Int("count", len(subs)))
	return subs, nil
}

func (uc *SubscriptionUseCase) ProcessNotifications(ctx context.Context) error {
	const op = "SubscriptionUseCase.ProcessNotifications"
	log := uc.logger.With(slog.String("op", op))

	repos, err := uc.repoRepo.GetAll(ctx)
	if err != nil {
		log.ErrorContext(ctx, errMsgGetRepos, slog.String("error", err.Error()))
		return fmt.Errorf("%s: %s: %w", op, errMsgGetRepos, err)
	}

	log.InfoContext(ctx, "starting check for updates", slog.Int("count", len(repos)))

	for _, repo := range repos {
		log := log.With(slog.String("repo", repo.FullName))
		log.InfoContext(ctx, "checking repo")

		repoCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		latestRelease, err := uc.ghClient.GetLatestRelease(repoCtx, repo.FullName)
		cancel()
		if err != nil {
			log.ErrorContext(ctx, errMsgFetchRelease, slog.String("error", err.Error()))
			continue
		}

		if latestRelease.TagName == repo.LastSeenTag {
			log.InfoContext(ctx, "no new updates", slog.String("tag", repo.LastSeenTag))
			continue
		}

		log.InfoContext(ctx, "new release detected",
			slog.String("old_tag", repo.LastSeenTag),
			slog.String("new_tag", latestRelease.TagName),
		)

		if err = uc.repoRepo.UpdateLastSeenTag(ctx, repo.FullName, latestRelease.TagName); err != nil {
			log.ErrorContext(ctx, errMsgUpdateTag, slog.String("error", err.Error()))
			continue
		}

		emails, err := uc.subsRepo.GetEmailsByRepoID(ctx, repo.ID)
		if err != nil {
			log.ErrorContext(ctx, errMsgGetSubscribers, slog.String("error", err.Error()))
			continue
		}

		sendCtx := context.WithoutCancel(ctx)
		for _, email := range emails {
			go func() {
				if err := uc.emailSender.SendNotification(sendCtx, email, repo.FullName, latestRelease.TagName); err != nil {
					log.ErrorContext(sendCtx, "failed to send email",
						slog.String("to", email),
						slog.String("error", err.Error()),
					)
				}
			}()
		}
	}
	return nil
}
