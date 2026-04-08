package handlers

import (
	"errors"
	"net/http"

	"github.com/ReilEgor/RepoNotifier/internal/domain/service"
	"github.com/ReilEgor/RepoNotifier/internal/transport/http/dto"
	"github.com/gin-gonic/gin"
)

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
// @Router       /subscriptions [post]
func (h *Handler) Subscribe(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}
	id, err := h.subscriptionUC.Subscribe(ctx, req.Email, req.Repository)
	if err != nil {
		if errors.Is(err, service.ErrRepositoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to subscribe: " + err.Error()})
		return
	}
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
// @Router       /subscriptions [delete]
func (h *Handler) Unsubscribe(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.DeleteSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}
	err := h.subscriptionUC.Unsubscribe(ctx, req.Email, req.Repository)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unsubscribe: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.DeleteSubscriptionResponse{
		Message: "Subscription deleted successfully",
	})
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
// @Router       /subscriptions [get]
func (h *Handler) ListSubscriptions(c *gin.Context) {
	ctx := c.Request.Context()
	email := c.Query("email")

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	subs, err := h.subscriptionUC.ListByEmail(ctx, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions: " + err.Error()})
		return
	}

	responseSubs := make([]dto.SubscriptionResponse, len(subs))
	for i, s := range subs {
		responseSubs[i] = dto.SubscriptionResponse{
			ID:           s.ID,
			Email:        email,
			RepositoryID: s.RepositoryID,
			CreatedAt:    s.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.ListSubscriptionsResponse{
		Subscriptions: responseSubs,
		Total:         len(responseSubs),
	})
}
