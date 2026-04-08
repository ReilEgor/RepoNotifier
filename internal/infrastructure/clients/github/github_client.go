package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/ReilEgor/RepoNotifier/internal/domain/service"
	"github.com/sony/gobreaker"
)

const (
	// HTTP client
	HttpClientTimeout = 10 * time.Second

	// Cache TTL
	CacheTTL = 10 * time.Minute

	// Cache key prefixes
	CacheKeyRepoExists    = "repo_exists:"
	CacheKeyLatestRelease = "release:"

	// Circuit breaker
	CbName             = "GitHubAPI"
	CbMaxRequests      = 3
	CbInterval         = 5 * time.Second
	CbTimeout          = 30 * time.Second
	CbFailureThreshold = 3

	// GitHub API
	GithubAPIBase    = "https://api.github.com"
	GithubAPIVersion = "2026-03-10"
	UserAgent        = "RepoNotifier/1.0"
)

type GitHubClient struct {
	httpClient *http.Client
	cache      service.Cache
	logger     *slog.Logger
	cb         *gobreaker.CircuitBreaker
}

func NewGitHubClient(cache service.Cache) *GitHubClient {
	settings := gobreaker.Settings{
		Name:        CbName,
		MaxRequests: CbMaxRequests,
		Interval:    CbInterval,
		Timeout:     CbTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= CbFailureThreshold
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			slog.Warn("circuit breaker state changed",
				slog.String("component", "GithubClient"),
				slog.String("breaker", name),
				slog.String("from", from.String()),
				slog.String("to", to.String()),
			)
		},
	}
	return &GitHubClient{
		httpClient: &http.Client{Timeout: HttpClientTimeout},
		cache:      cache,
		logger:     slog.With(slog.String("component", "GithubClient")),
		cb:         gobreaker.NewCircuitBreaker(settings),
	}
}
func (c *GitHubClient) RepoExists(ctx context.Context, fullName string) (bool, error) {
	log := c.logger.With(slog.String("op", "RepoExists"), slog.String("repo", fullName))
	cacheKey := CacheKeyRepoExists + fullName

	cached, err := c.cache.Get(ctx, cacheKey)
	if err != nil {
		log.ErrorContext(ctx, "cache get failed", slog.String("key", cacheKey), slog.Any("error", err))
		return false, err
	}
	if cached != "" {
		log.DebugContext(ctx, "cache hit", slog.String("key", cacheKey))
		return cached == "true", nil
	}
	log.DebugContext(ctx, "cache miss", slog.String("key", cacheKey))

	result, err := c.cb.Execute(func() (interface{}, error) {
		url := fmt.Sprintf("https://api.github.com/repos/%s", fullName)
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
		if err != nil {
			return false, fmt.Errorf("create request: %w", err)
		}
		setDefaultHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.ErrorContext(ctx, "request failed", slog.String("url", url), slog.Any("error", err))
			return false, err
		}
		defer resp.Body.Close()

		log.DebugContext(ctx, "response received", slog.String("url", url), slog.Int("status", resp.StatusCode))

		switch resp.StatusCode {
		case http.StatusOK:
			return true, nil
		case http.StatusNotFound:
			return false, nil
		default:
			return false, errors.New(resp.Status)
		}
	})

	if err != nil {
		log.ErrorContext(ctx, "circuit breaker execute failed", slog.Any("error", err))
		return false, err
	}
	exists := result.(bool)
	val := "false"
	if exists {
		val = "true"
	}
	if err = c.cache.Set(ctx, cacheKey, val, CacheTTL); err != nil {
		log.ErrorContext(ctx, "cache set failed", slog.String("key", cacheKey), slog.Any("error", err))
		return false, err
	}
	log.InfoContext(ctx, "repo existence checked", slog.Bool("exists", exists))
	return exists, nil
}

func (c *GitHubClient) GetLatestRelease(ctx context.Context, fullName string) (*model.ReleaseInfo, error) {
	log := c.logger.With(slog.String("op", "GetLatestRelease"), slog.String("repo", fullName))

	cacheKey := CacheKeyLatestRelease + fullName

	cached, err := c.cache.Get(ctx, cacheKey)
	if err != nil {
		log.ErrorContext(ctx, "cache get failed", slog.String("key", cacheKey), slog.Any("error", err))
		return nil, err
	}
	if cached != "" {
		var info model.ReleaseInfo
		if err := json.Unmarshal([]byte(cached), &info); err != nil {
			log.WarnContext(ctx, "cache unmarshal failed, fetching from api",
				slog.String("key", cacheKey),
				slog.Any("error", err),
			)
		} else {
			log.DebugContext(ctx, "cache hit", slog.String("key", cacheKey))
			return &info, nil
		}
	} else {
		log.DebugContext(ctx, "cache miss", slog.String("key", cacheKey))
	}

	url := fmt.Sprintf("%s/repos/%s/releases/latest", GithubAPIBase, fullName)

	result, err := c.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		setDefaultHeaders(req)
		log.DebugContext(ctx, "sending request", slog.String("method", http.MethodGet), slog.String("url", url))
		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.ErrorContext(ctx, "request failed", slog.String("url", url), slog.Any("error", err))
			return nil, err
		}
		defer resp.Body.Close()
		log.DebugContext(ctx, "response received", slog.String("url", url), slog.Int("status", resp.StatusCode))
		if resp.StatusCode == http.StatusNotFound {
			return nil, service.ErrReleaseNotFound
		}
		if resp.StatusCode != http.StatusOK {
			var errorResponse struct {
				Message string `json:"message"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil && errorResponse.Message != "" {
				return nil, fmt.Errorf("github api error (%d): %s", resp.StatusCode, errorResponse.Message)
			}
			return nil, fmt.Errorf("github api unexpected status: %s", resp.Status)
		}
		var info model.ReleaseInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			log.ErrorContext(ctx, "response decode failed", slog.String("url", url), slog.Any("error", err))
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &info, nil
	})

	if err != nil {
		log.ErrorContext(ctx, "circuit breaker execute failed", slog.Any("error", err))
		return nil, err
	}

	info := result.(*model.ReleaseInfo)

	jsonData, err := json.Marshal(info)
	if err != nil {
		log.ErrorContext(ctx, "marshal for cache failed", slog.Any("error", err))
		return nil, fmt.Errorf("marshal release info: %w", err)
	}
	if err = c.cache.Set(ctx, cacheKey, string(jsonData), CacheTTL); err != nil {
		log.ErrorContext(ctx, "cache set failed", slog.String("key", cacheKey), slog.Any("error", err))
		return nil, err
	}

	log.InfoContext(ctx, "latest release fetched", slog.String("tag", info.TagName), slog.Time("published_at", info.PublishedAt))
	return info, nil
}

func setDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", GithubAPIVersion)
	req.Header.Set("User-Agent", UserAgent)
}
