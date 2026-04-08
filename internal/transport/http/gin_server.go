package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/ReilEgor/RepoNotifier/internal/domain/usecase"
	handler "github.com/ReilEgor/RepoNotifier/internal/transport/http/handlers"
	"github.com/ReilEgor/RepoNotifier/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type GinServer struct {
	router         *gin.Engine
	subscriptionUC usecase.SubscriptionUseCase
	userUC         usecase.UserUseCase
	logger         *slog.Logger
}

func NewGinServer(
	subscriptionUC usecase.SubscriptionUseCase,
	userUC usecase.UserUseCase,
	redisClient *redis.Client,
) *GinServer {
	router := gin.New()
	logger := slog.With(slog.String("component", "gin_server"))
	middleware.SetupMiddleware(router, logger, redisClient)

	s := &GinServer{
		router:         router,
		subscriptionUC: subscriptionUC,
		userUC:         userUC,
		logger:         logger,
	}

	h := handler.NewHandler(subscriptionUC, userUC)
	h.InitRoutes(s.router)

	return s
}

func (s *GinServer) Run(ctx context.Context, port string) error {
	srv := &http.Server{Addr: port, Handler: s.router}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("forced shutdown", "error", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
