package handlers

import (
	"log/slog"
	"net/http"

	"github.com/ReilEgor/RepoNotifier/internal/domain/usecase"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	subscriptionUC usecase.SubscriptionUseCase
	userUC         usecase.UserUseCase
	logger         *slog.Logger
}

func NewHandler(subscriptionUC usecase.SubscriptionUseCase, userUC usecase.UserUseCase) *Handler {
	return &Handler{
		subscriptionUC: subscriptionUC,
		userUC:         userUC,
		logger:         slog.With(slog.String("component", "handler")),
	}
}

func (h *Handler) InitRoutes(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := router.Group("/api/v1")
	{
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.POST("/", h.Subscribe)
			subscriptions.DELETE("/", h.Unsubscribe)
			subscriptions.GET("/", h.ListSubscriptions)
		}
	}
}
