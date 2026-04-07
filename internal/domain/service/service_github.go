package service

import (
	"context"
	"errors"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
)

var (
	ErrRateLimitExceeded  = errors.New("github api rate limit exceeded")
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrNoReleasesFound    = errors.New("no releases found for this repository")
	ErrGitHubUnavailable  = errors.New("github service is temporarily unavailable")
)

type GitHubClient interface {
	RepoExists(ctx context.Context, fullName string) (bool, error)
	GetLatestRelease(ctx context.Context, fullName string) (*model.ReleaseInfo, error)
}
