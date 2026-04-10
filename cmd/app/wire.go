//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/ReilEgor/RepoNotifier/internal/config"
	repositoryInterface "github.com/ReilEgor/RepoNotifier/internal/domain/repository"
	servicesInterface "github.com/ReilEgor/RepoNotifier/internal/domain/service"
	usecaseInterface "github.com/ReilEgor/RepoNotifier/internal/domain/usecase"
	cacheRealization "github.com/ReilEgor/RepoNotifier/internal/infrastructure/cache/redis"
	servicesRealizationEmail "github.com/ReilEgor/RepoNotifier/internal/infrastructure/clients/email"
	servicesRealizationGitHub "github.com/ReilEgor/RepoNotifier/internal/infrastructure/clients/github"
	repository "github.com/ReilEgor/RepoNotifier/internal/infrastructure/storage/postgres"
	repositoryRealization "github.com/ReilEgor/RepoNotifier/internal/repository/postgres"
	"github.com/ReilEgor/RepoNotifier/internal/transport/http"
	"github.com/ReilEgor/RepoNotifier/internal/transport/http/handlers"
	usecaseRealization "github.com/ReilEgor/RepoNotifier/internal/usecase"
	"github.com/google/wire"
)

var UseCaseSet = wire.NewSet(
	usecaseRealization.NewSubscriptionUseCase,
	usecaseRealization.NewUserUseCase,
	wire.Bind(new(usecaseInterface.UserUseCase), new(*usecaseRealization.UserUseCase)),
	wire.Bind(new(usecaseInterface.SubscriptionUseCase), new(*usecaseRealization.SubscriptionUseCase)),
)

var RepositorySet = wire.NewSet(
	repository.New,
	repositoryRealization.NewRepositoryRepository,
	repositoryRealization.NewSubscriptionRepository,
	repositoryRealization.NewUserRepository,
	wire.Bind(new(repositoryInterface.RepositoryRepository), new(*repositoryRealization.RepositoryRepository)),
	wire.Bind(new(repositoryInterface.SubscriptionRepository), new(*repositoryRealization.SubscriptionRepository)),
	wire.Bind(new(repositoryInterface.UserRepository), new(*repositoryRealization.UserRepository)),
)

var RestSet = wire.NewSet(
	http.NewGinServer,
	handlers.NewHandler,
)

var CacheSet = wire.NewSet(
	cacheRealization.NewRedisClient,
	cacheRealization.NewCache,
	wire.Bind(new(servicesInterface.Cache), new(*cacheRealization.Cache)),
)
var ServicesSet = wire.NewSet(
	servicesRealizationGitHub.NewGitHubClient,
	servicesRealizationEmail.NewSmtpClient,
	wire.Bind(new(servicesInterface.EmailSender), new(*servicesRealizationEmail.SmtpClient)),
	wire.Bind(new(servicesInterface.GitHubClient), new(*servicesRealizationGitHub.GitHubClient)),
)

type App struct {
	Server              *http.GinServer
	SubscriptionUseCase usecaseInterface.SubscriptionUseCase
}

func InitializeApp(
	ctx context.Context,
	redisHost config.RedisHostType,
	redisPort config.RedisPortType,
	redisPassword config.RedisPasswordType,
	redisDB int,
	dsn config.DSNType,
	emailHost config.EmailHostType,
	emailPort config.EmailPortType,
	emailPassword config.EmailPasswordType,
	emailFrom config.EmailFromType,
	emailUser config.EmailUserType,
	apiKey config.ApiKeyType,
) (*App, func(), error) {
	wire.Build(
		ServicesSet,
		RepositorySet,
		UseCaseSet,
		CacheSet,
		RestSet,
		wire.Struct(new(App), "*"),
	)
	return nil, nil, nil
}
