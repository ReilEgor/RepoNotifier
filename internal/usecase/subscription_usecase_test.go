package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/ReilEgor/RepoNotifier/internal/domain/service"
	"github.com/ReilEgor/RepoNotifier/internal/mocks"
	_ "github.com/ReilEgor/RepoNotifier/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockFields struct {
	subsRepo    *mocks.SubscriptionRepository
	userRepo    *mocks.UserRepository
	repoRepo    *mocks.RepositoryRepository
	ghClient    *mocks.GitHubClient
	emailSender *mocks.EmailSender
}

func newMockFields(t *testing.T) mockFields {
	return mockFields{
		subsRepo:    mocks.NewSubscriptionRepository(t),
		userRepo:    mocks.NewUserRepository(t),
		repoRepo:    mocks.NewRepositoryRepository(t),
		ghClient:    mocks.NewGitHubClient(t),
		emailSender: mocks.NewEmailSender(t),
	}
}

func TestSubscriptionUseCase_Subscribe(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		repoName  string
		setup     func(f mockFields)
		expected  int64
		wantErr   error
		expectErr bool
	}{
		{
			name:     "success: new subscription created",
			email:    "user@test.com",
			repoName: "golang/go",
			setup: func(f mockFields) {
				f.ghClient.On("RepoExists", mock.Anything, "golang/go").
					Return(true, nil).Once()

				f.ghClient.On("GetLatestRelease", mock.Anything, "golang/go").
					Return(&model.ReleaseInfo{TagName: "v1.22.0"}, nil).Once()

				f.userRepo.On("GetOrCreate", mock.Anything, "user@test.com").
					Return(model.User{ID: 1}, nil).Once()

				f.repoRepo.On("GetOrCreate", mock.Anything, "golang/go", "v1.22.0").
					Return(&model.Repository{ID: 10}, nil).Once()

				f.subsRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *model.Subscription) bool {
					return s.UserID == 1 && s.RepositoryID == 10
				})).Return(int64(100), nil).Once()
			},
			expected: 100,
		},
		{
			name:     "error: repository not found on github",
			email:    "user@test.com",
			repoName: "unknown/repo",
			setup: func(f mockFields) {
				f.ghClient.On("RepoExists", mock.Anything, "unknown/repo").
					Return(false, nil).Once()
			},
			wantErr: service.ErrRepositoryNotFound,
		},
		{
			name:     "error: github client fails",
			email:    "user@test.com",
			repoName: "golang/go",
			setup: func(f mockFields) {
				f.ghClient.On("RepoExists", mock.Anything, "golang/go").
					Return(false, errors.New("api connection error")).Once()
			},
			expectErr: true,
		},
		{
			name:     "error: failed to get latest release",
			email:    "user@test.com",
			repoName: "golang/go",
			setup: func(f mockFields) {
				f.ghClient.On("RepoExists", mock.Anything, "golang/go").
					Return(true, nil).Once()

				f.ghClient.On("GetLatestRelease", mock.Anything, "golang/go").
					Return(nil, errors.New("github error")).Once()
			},
			expectErr: true,
		},
		{
			name:     "error: failed to create subscription",
			email:    "user@test.com",
			repoName: "golang/go",
			setup: func(f mockFields) {
				f.ghClient.On("RepoExists", mock.Anything, "golang/go").
					Return(true, nil).Once()

				f.ghClient.On("GetLatestRelease", mock.Anything, "golang/go").
					Return(&model.ReleaseInfo{TagName: "v1.22.0"}, nil).Once()

				f.userRepo.On("GetOrCreate", mock.Anything, "user@test.com").
					Return(model.User{ID: 1}, nil).Once()

				f.repoRepo.On("GetOrCreate", mock.Anything, "golang/go", "v1.22.0").
					Return(&model.Repository{ID: 10}, nil).Once()

				f.subsRepo.On("Create", mock.Anything, mock.Anything).
					Return(int64(0), errors.New("db error")).Once()
			},
			expectErr: true,
		},
		{
			name:     "error: failed to create or get user",
			email:    "user@test.com",
			repoName: "golang/go",
			setup: func(f mockFields) {
				f.ghClient.On("RepoExists", mock.Anything, "golang/go").
					Return(true, nil).Once()

				f.ghClient.On("GetLatestRelease", mock.Anything, "golang/go").
					Return(&model.ReleaseInfo{TagName: "v1.22.0"}, nil).Once()

				f.userRepo.On("GetOrCreate", mock.Anything, "user@test.com").
					Return(model.User{}, errors.New("user repo error")).Once()
			},
			expectErr: true,
		},
		{
			name:     "error: failed to get or create repo",
			email:    "user@test.com",
			repoName: "golang/go",
			setup: func(f mockFields) {
				f.ghClient.On("RepoExists", mock.Anything, "golang/go").
					Return(true, nil).Once()

				f.ghClient.On("GetLatestRelease", mock.Anything, "golang/go").
					Return(&model.ReleaseInfo{TagName: "v1.22.0"}, nil).Once()

				f.userRepo.On("GetOrCreate", mock.Anything, "user@test.com").
					Return(model.User{ID: 1}, nil).Once()

				f.repoRepo.On("GetOrCreate", mock.Anything, "golang/go", "v1.22.0").
					Return(&model.Repository{}, errors.New("repo repo error")).Once()
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newMockFields(t)
			tt.setup(f)

			uc := NewSubscriptionUseCase(
				f.subsRepo,
				f.ghClient,
				f.userRepo,
				f.repoRepo,
				f.emailSender,
			)

			id, err := uc.Subscribe(context.Background(), tt.email, tt.repoName)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, id)

			f.emailSender.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
		})
	}
}

func TestSubscriptionUseCase_Unsubscribe(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		repoName  string
		setup     func(f mockFields)
		wantErr   error
		expectErr bool
	}{
		{
			name:     "success: unsubscribe",
			email:    "user@test.com",
			repoName: "golang/go",
			setup: func(f mockFields) {
				f.userRepo.On("GetByEmail", mock.Anything, "user@test.com").
					Return(model.User{ID: 1}, nil).Once()

				f.subsRepo.On("Delete", mock.Anything, int64(1), "golang/go").
					Return(nil).Once()
			},
		},
		{
			name:     "user not found: no error",
			email:    "user@test.com",
			repoName: "golang/go",
			setup: func(f mockFields) {
				f.userRepo.On("GetByEmail", mock.Anything, "user@test.com").
					Return(model.User{}, model.ErrUserNotFound).Once()
			},
		},
		{
			name:     "error: get user fails",
			email:    "user@test.com",
			repoName: "golang/go",
			setup: func(f mockFields) {
				f.userRepo.On("GetByEmail", mock.Anything, "user@test.com").
					Return(model.User{}, errors.New("db error")).Once()
			},
			expectErr: true,
		},
		{
			name:     "error: delete subscription fails",
			email:    "user@test.com",
			repoName: "golang/go",
			setup: func(f mockFields) {
				f.userRepo.On("GetByEmail", mock.Anything, "user@test.com").
					Return(model.User{ID: 1}, nil).Once()

				f.subsRepo.On("Delete", mock.Anything, int64(1), "golang/go").
					Return(errors.New("delete error")).Once()
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newMockFields(t)
			tt.setup(f)

			uc := NewSubscriptionUseCase(
				f.subsRepo,
				f.ghClient,
				f.userRepo,
				f.repoRepo,
				f.emailSender,
			)

			err := uc.Unsubscribe(context.Background(), tt.email, tt.repoName)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.name == "user not found: no error" {
				f.subsRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
			}

			mock.AssertExpectationsForObjects(
				t,
				f.userRepo,
				f.subsRepo,
			)
		})
	}
}

func TestSubscriptionUseCase_ListByEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		setup     func(f mockFields)
		expected  []model.Subscription
		wantErr   error
		expectErr bool
	}{
		{
			name:  "success: subscriptions returned",
			email: "user@test.com",
			setup: func(f mockFields) {
				f.userRepo.On("GetByEmail", mock.Anything, "user@test.com").
					Return(model.User{ID: 1}, nil).Once()

				expectedSubs := []model.Subscription{
					{ID: 1, UserID: 1, RepositoryID: 10},
					{ID: 2, UserID: 1, RepositoryID: 20},
				}

				f.subsRepo.On("GetByUserID", mock.Anything, int64(1)).
					Return(expectedSubs, nil).Once()
			},
			expected: []model.Subscription{
				{ID: 1, UserID: 1, RepositoryID: 10},
				{ID: 2, UserID: 1, RepositoryID: 20},
			},
		},
		{
			name:  "user not found: return empty list",
			email: "user@test.com",
			setup: func(f mockFields) {
				f.userRepo.On("GetByEmail", mock.Anything, "user@test.com").
					Return(model.User{}, model.ErrUserNotFound).Once()
			},
			expected: []model.Subscription{},
		},
		{
			name:  "error: get user fails",
			email: "user@test.com",
			setup: func(f mockFields) {
				f.userRepo.On("GetByEmail", mock.Anything, "user@test.com").
					Return(model.User{}, errors.New("db error")).Once()
			},
			expectErr: true,
		},
		{
			name:  "error: get subscriptions fails",
			email: "user@test.com",
			setup: func(f mockFields) {
				f.userRepo.On("GetByEmail", mock.Anything, "user@test.com").
					Return(model.User{ID: 1}, nil).Once()

				f.subsRepo.On("GetByUserID", mock.Anything, int64(1)).
					Return(nil, errors.New("db error")).Once()
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newMockFields(t)
			tt.setup(f)

			uc := NewSubscriptionUseCase(
				f.subsRepo,
				f.ghClient,
				f.userRepo,
				f.repoRepo,
				f.emailSender,
			)

			subs, err := uc.ListByEmail(context.Background(), tt.email)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, subs)

			if tt.name == "user not found: return empty list" {
				f.subsRepo.AssertNotCalled(t, "GetByUserID", mock.Anything, mock.Anything)
			}

			mock.AssertExpectationsForObjects(
				t,
				f.userRepo,
				f.subsRepo,
			)
		})
	}
}

func TestSubscriptionUseCase_ProcessNotifications(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(f mockFields, wg *sync.WaitGroup)
		expectErr bool
	}{
		{
			name: "success: new release triggers emails",
			setup: func(f mockFields, wg *sync.WaitGroup) {
				repos := []model.Repository{
					{
						ID:          1,
						FullName:    "golang/go",
						LastSeenTag: "v1.21.0",
					},
				}

				f.repoRepo.On("GetAll", mock.Anything).
					Return(repos, nil).Once()

				f.ghClient.On("GetLatestRelease", mock.Anything, "golang/go").
					Return(&model.ReleaseInfo{TagName: "v1.22.0"}, nil).Once()

				f.repoRepo.On("UpdateLastSeenTag", mock.Anything, "golang/go", "v1.22.0").
					Return(nil).Once()

				emails := []string{"a@test.com", "b@test.com"}

				f.subsRepo.On("GetEmailsByRepoID", mock.Anything, int64(1)).
					Return(emails, nil).Once()

				wg.Add(len(emails))

				for _, email := range emails {
					f.emailSender.
						On("SendNotification", mock.Anything, email, "golang/go", "v1.22.0").
						Run(func(args mock.Arguments) {
							wg.Done()
						}).
						Return(nil).
						Once()
				}
			},
		},
		{
			name: "no updates: no emails sent",
			setup: func(f mockFields, wg *sync.WaitGroup) {
				repos := []model.Repository{
					{
						ID:          1,
						FullName:    "golang/go",
						LastSeenTag: "v1.22.0",
					},
				}

				f.repoRepo.On("GetAll", mock.Anything).
					Return(repos, nil).Once()

				f.ghClient.On("GetLatestRelease", mock.Anything, "golang/go").
					Return(&model.ReleaseInfo{TagName: "v1.22.0"}, nil).Once()
			},
		},
		{
			name: "error: GetAll fails",
			setup: func(f mockFields, wg *sync.WaitGroup) {
				f.repoRepo.On("GetAll", mock.Anything).
					Return(nil, errors.New("db error")).Once()
			},
			expectErr: true,
		},
		{
			name: "error: github fails then continue",
			setup: func(f mockFields, wg *sync.WaitGroup) {
				repos := []model.Repository{
					{ID: 1, FullName: "golang/go", LastSeenTag: "v1.21.0"},
				}

				f.repoRepo.On("GetAll", mock.Anything).
					Return(repos, nil).Once()

				f.ghClient.On("GetLatestRelease", mock.Anything, "golang/go").
					Return(nil, errors.New("api error")).Once()
			},
		},
		{
			name: "error: update tag fails then continue",
			setup: func(f mockFields, wg *sync.WaitGroup) {
				repos := []model.Repository{
					{ID: 1, FullName: "golang/go", LastSeenTag: "v1.21.0"},
				}

				f.repoRepo.On("GetAll", mock.Anything).
					Return(repos, nil).Once()

				f.ghClient.On("GetLatestRelease", mock.Anything, "golang/go").
					Return(&model.ReleaseInfo{TagName: "v1.22.0"}, nil).Once()

				f.repoRepo.On("UpdateLastSeenTag", mock.Anything, "golang/go", "v1.22.0").
					Return(errors.New("db error")).Once()
			},
		},
		{
			name: "error: get subscribers fails then continue",
			setup: func(f mockFields, wg *sync.WaitGroup) {
				repos := []model.Repository{
					{ID: 1, FullName: "golang/go", LastSeenTag: "v1.21.0"},
				}

				f.repoRepo.On("GetAll", mock.Anything).
					Return(repos, nil).Once()

				f.ghClient.On("GetLatestRelease", mock.Anything, "golang/go").
					Return(&model.ReleaseInfo{TagName: "v1.22.0"}, nil).Once()

				f.repoRepo.On("UpdateLastSeenTag", mock.Anything, "golang/go", "v1.22.0").
					Return(nil).Once()

				f.subsRepo.On("GetEmailsByRepoID", mock.Anything, int64(1)).
					Return(nil, errors.New("db error")).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newMockFields(t)
			var wg sync.WaitGroup

			tt.setup(f, &wg)

			uc := NewSubscriptionUseCase(
				f.subsRepo,
				f.ghClient,
				f.userRepo,
				f.repoRepo,
				f.emailSender,
			)

			err := uc.ProcessNotifications(context.Background())

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			wg.Wait()

			mock.AssertExpectationsForObjects(
				t,
				f.repoRepo,
				f.subsRepo,
				f.ghClient,
				f.emailSender,
			)
		})
	}
}
