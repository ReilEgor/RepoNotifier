package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/ReilEgor/RepoNotifier/internal/domain/service"
	"github.com/ReilEgor/RepoNotifier/internal/transport/http/dto"
	"github.com/gin-gonic/gin"
)

const (
	timeoutSubscribe   = 10 * time.Second
	timeoutUnsubscribe = 3 * time.Second
	timeoutList        = 3 * time.Second
)

const (
	errInvalidRequestBody  = "invalid request body"
	errFailedToSubscribe   = "failed to subscribe"
	errFailedToUnsubscribe = "failed to unsubscribe"
	errFailedToList        = "failed to list subscriptions"
)

const (
	msgSubscriptionDeleted = "Subscription deleted successfully"
)

var (
	ErrInvalidEmailFormat   = errors.New("invalid email format")
	ErrInvalidRepoFormat    = errors.New("invalid repository format (expected 'owner/repo')")
	ErrEmailRequired        = errors.New("email is required")
	ErrAlreadySubscribed    = errors.New("user is already subscribed to this repository")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

var repoRegex = regexp.MustCompile(`^[a-zA-Z0-9-._]+/[a-zA-Z0-9-._]+$`)

func validateEmail(email string) error {
	if _, err := mail.ParseAddress(strings.TrimSpace(email)); err != nil {
		return ErrInvalidEmailFormat
	}
	return nil
}
func validateSubscription(email, repo string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	if !repoRegex.MatchString(strings.TrimSpace(repo)) {
		return ErrInvalidRepoFormat
	}
	return nil
}

// @Summary      Subscribe to a repository
// @Description  Create a subscription for a user to track the latest releases of a GitHub repository.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateSubscriptionRequest  true  "Subscription details"
// @Success      201      {object}  dto.CreateSubscriptionResponse
// @Failure      400      {object}  map[string]string "Invalid request body"
// @Failure      404      {object}  map[string]string "Repository not found on GitHub"
// @Failure      500      {object}  map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router       /subscriptions [post]
func (h *Handler) Subscribe(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeoutSubscribe)
	defer cancel()
	log := slog.With(slog.String("handler", "Subscribe"))
	var req dto.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnContext(ctx, errInvalidRequestBody, slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errInvalidRequestBody, err.Error())})
		return
	}
	if err := validateSubscription(req.Email, req.Repository); err != nil {
		log.WarnContext(ctx, "validation failed",
			slog.String("email", req.Email),
			slog.String("repo", req.Repository),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.subscriptionUC.Subscribe(ctx, req.Email, req.Repository)
	if err != nil {
		if errors.Is(err, service.ErrRepositoryNotFound) {
			log.WarnContext(ctx, "repository not found",
				slog.String("email", req.Email),
				slog.String("repo", req.Repository),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrAlreadySubscribed) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrGitHubUnavailable) || errors.Is(err, service.ErrRateLimitExceeded) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "GitHub API is currently unavailable, please try again later"})
			return
		}
		log.ErrorContext(ctx, errFailedToSubscribe,
			slog.String("email", req.Email),
			slog.String("repo", req.Repository),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errFailedToSubscribe, err.Error())})
		return
	}
	log.InfoContext(ctx, "subscribed successfully",
		slog.String("email", req.Email),
		slog.String("repo", req.Repository),
		slog.Any("id", id),
	)
	c.JSON(http.StatusCreated, dto.CreateSubscriptionResponse{ID: id})
}

// @Summary      Unsubscribe from a repository
// @Description  Remove a subscription for a specific user and repository.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        request  body      dto.DeleteSubscriptionRequest  true  "Unsubscribe details"
// @Success      200      {object}  dto.DeleteSubscriptionResponse
// @Failure      400      {object}  map[string]string "Invalid request body"
// @Failure      500      {object}  map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router       /subscriptions [delete]
func (h *Handler) Unsubscribe(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeoutUnsubscribe)
	defer cancel()
	log := slog.With(slog.String("handler", "Unsubscribe"))
	var req dto.DeleteSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnContext(ctx, errInvalidRequestBody, slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errInvalidRequestBody, err.Error())})
		return
	}
	if err := validateSubscription(req.Email, req.Repository); err != nil {
		log.WarnContext(ctx, "validation failed", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.subscriptionUC.Unsubscribe(ctx, req.Email, req.Repository); err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		log.ErrorContext(ctx, errFailedToUnsubscribe,
			slog.String("email", req.Email),
			slog.String("repo", req.Repository),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errFailedToUnsubscribe, err.Error())})
		return
	}

	log.InfoContext(ctx, "unsubscribed successfully",
		slog.String("email", req.Email),
		slog.String("repo", req.Repository),
	)
	c.JSON(http.StatusOK, dto.DeleteSubscriptionResponse{Message: msgSubscriptionDeleted})
}

// ListSubscriptions godoc
// @Summary      List user subscriptions
// @Description  Get a list of all repositories the user is currently subscribed to.
// @Tags         subscriptions
// @Produce      json
// @Param        email    query     string  true  "User email address"
// @Success      200      {object}  dto.ListSubscriptionsResponse
// @Failure      400      {object}  map[string]string "Email is required"
// @Failure      500      {object}  map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router       /subscriptions [get]
func (h *Handler) ListSubscriptions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeoutList)
	defer cancel()
	log := slog.With(slog.String("handler", "ListSubscriptions"))
	email := c.Query("email")

	if email == "" {
		log.WarnContext(ctx, "email query param missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrEmailRequired.Error()})
		return
	}
	if err := validateEmail(email); err != nil {
		log.WarnContext(ctx, "invalid email query param",
			slog.String("email", email),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subs, err := h.subscriptionUC.ListByEmail(ctx, email)
	if err != nil {
		log.ErrorContext(ctx, errFailedToList,
			slog.String("email", email),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errFailedToList, err.Error())})
		return
	}

	responseSubs := make([]dto.SubscriptionResponse, len(subs))
	for i, s := range subs {
		responseSubs[i] = dto.SubscriptionResponse{
			ID:             s.ID,
			Email:          email,
			RepositoryID:   s.RepositoryID,
			RepositoryName: s.RepositoryName,
			CreatedAt:      s.CreatedAt,
		}
	}

	log.InfoContext(ctx, "listed subscriptions",
		slog.String("email", email),
		slog.Int("total", len(responseSubs)),
	)
	c.JSON(http.StatusOK, dto.ListSubscriptionsResponse{
		Subscriptions: responseSubs,
		Total:         len(responseSubs),
	})
}
