package postgres

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionRepository_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &SubscriptionRepository{
		db:     mock,
		logger: discardLogger,
	}

	tests := []struct {
		name        string
		userID      int64
		repoID      int64
		mockSetup   func(uid, rid int64)
		expectError bool
		expectedID  int64
	}{
		{
			name:   "success create",
			userID: 1,
			repoID: 10,
			mockSetup: func(uid, rid int64) {
				mock.ExpectQuery("INSERT INTO subscriptions").
					WithArgs(uid, rid).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(100)))
			},
			expectError: false,
			expectedID:  100,
		},
		{
			name:   "on conflict success",
			userID: 1,
			repoID: 10,
			mockSetup: func(uid, rid int64) {
				mock.ExpectQuery("INSERT INTO subscriptions").
					WithArgs(uid, rid).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(100)))
			},
			expectError: false,
			expectedID:  100,
		},
		{
			name:   "database error",
			userID: 1,
			repoID: 10,
			mockSetup: func(uid, rid int64) {
				mock.ExpectQuery("INSERT INTO subscriptions").
					WithArgs(uid, rid).
					WillReturnError(fmt.Errorf("foreign key violation"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.userID, tt.repoID)

			sub := &model.Subscription{
				UserID:       tt.userID,
				RepositoryID: tt.repoID,
			}

			id, err := repo.Create(context.Background(), sub)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, int64(0), id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestSubscriptionRepository_Delete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &SubscriptionRepository{
		db:     mock,
		logger: discardLogger,
	}

	userID := int64(1)

	tests := []struct {
		name        string
		repoName    string
		mockSetup   func(uid int64, repoName string)
		expectError bool
		checkErrMsg string
	}{
		{
			name:     "success delete",
			repoName: "google/wire",
			mockSetup: func(uid int64, repoName string) {
				mock.ExpectExec("DELETE FROM subscriptions").
					WithArgs(uid, repoName).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectError: false,
		},
		{
			name:     "subscription not found",
			repoName: "non-existent",
			mockSetup: func(uid int64, repoName string) {
				mock.ExpectExec("DELETE FROM subscriptions").
					WithArgs(uid, repoName).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectError: false,
		},
		{
			name:     "database error",
			repoName: "google/wire",
			mockSetup: func(uid int64, repoName string) {
				mock.ExpectExec("DELETE FROM subscriptions").
					WithArgs(uid, repoName).
					WillReturnError(fmt.Errorf("connection closed"))
			},
			expectError: true,
			checkErrMsg: "exec:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(userID, tt.repoName)

			err := repo.Delete(context.Background(), userID, tt.repoName)

			if tt.expectError {
				assert.Error(t, err)
				if tt.checkErrMsg != "" {
					assert.Contains(t, err.Error(), tt.checkErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestSubscriptionRepository_GetAll(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &SubscriptionRepository{
		db:     mock,
		logger: discardLogger,
	}

	now := time.Now()

	tests := []struct {
		name         string
		mockSetup    func()
		expectError  bool
		checkErrMsg  string
		expectedSubs []model.Subscription
	}{
		{
			name: "success with data",
			mockSetup: func() {
				mock.ExpectQuery("SELECT (.+) FROM subscriptions").
					WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "repository_id", "created_at"}).
						AddRow(int64(10), int64(1), int64(100), now).
						AddRow(int64(11), int64(1), int64(101), now))
			},
			expectError: false,
			expectedSubs: []model.Subscription{
				{ID: 10, UserID: 1, RepositoryID: 100, CreatedAt: now},
				{ID: 11, UserID: 1, RepositoryID: 101, CreatedAt: now},
			},
		},
		{
			name: "success empty results",
			mockSetup: func() {
				mock.ExpectQuery("SELECT (.+) FROM subscriptions").
					WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "repository_id", "created_at"}))
			},
			expectError:  false,
			expectedSubs: []model.Subscription{},
		},
		{
			name: "query error",
			mockSetup: func() {
				mock.ExpectQuery("SELECT (.+) FROM subscriptions").
					WillReturnError(fmt.Errorf("db error"))
			},
			expectError: true,
			checkErrMsg: "query:",
		},
		{
			name: "scan error",
			mockSetup: func() {
				mock.ExpectQuery("SELECT (.+) FROM subscriptions").
					WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "repository_id", "created_at"}).
						AddRow("wrong-id-type", 1, 100, now))
			},
			expectError: true,
			checkErrMsg: "scan:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.GetAll(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				if tt.checkErrMsg != "" {
					assert.Contains(t, err.Error(), tt.checkErrMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expectedSubs))

				for i, expected := range tt.expectedSubs {
					assert.Equal(t, expected.ID, result[i].ID)
					assert.Equal(t, expected.UserID, result[i].UserID)
					assert.Equal(t, expected.RepositoryID, result[i].RepositoryID)
					assert.Equal(t, expected.CreatedAt, result[i].CreatedAt)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSubscriptionRepository_GetByUserID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &SubscriptionRepository{
		db:     mock,
		logger: discardLogger,
	}

	userID := int64(42)
	now := time.Now()

	tests := []struct {
		name         string
		targetID     int64
		mockSetup    func(id int64)
		expectError  bool
		checkErrMsg  string
		expectedSubs []model.Subscription
	}{
		{
			name:     "success with subscriptions",
			targetID: userID,
			mockSetup: func(id int64) {
				columns := []string{"id", "user_id", "repository_id", "full_name", "created_at"}

				mock.ExpectQuery("SELECT (.+) FROM subscriptions s JOIN repositories r").
					WithArgs(id).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(100), id, int64(1), "google/wire", now).
						AddRow(int64(101), id, int64(2), "jackc/pgx", now))
			},
			expectError: false,
			expectedSubs: []model.Subscription{
				{ID: 100, UserID: userID, RepositoryID: 1, RepositoryName: "google/wire", CreatedAt: now},
				{ID: 101, UserID: userID, RepositoryID: 2, RepositoryName: "jackc/pgx", CreatedAt: now},
			},
		},
		{
			name:     "success no subscriptions",
			targetID: userID,
			mockSetup: func(id int64) {
				mock.ExpectQuery("SELECT (.+) FROM subscriptions").
					WithArgs(id).
					WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "repository_id", "full_name", "created_at"}))
			},
			expectError:  false,
			expectedSubs: []model.Subscription{},
		},
		{
			name:     "database error",
			targetID: userID,
			mockSetup: func(id int64) {
				mock.ExpectQuery("SELECT (.+) FROM subscriptions").
					WithArgs(id).
					WillReturnError(fmt.Errorf("connection refused"))
			},
			expectError: true,
			checkErrMsg: "query:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.targetID)

			result, err := repo.GetByUserID(context.Background(), tt.targetID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.checkErrMsg != "" {
					assert.Contains(t, err.Error(), tt.checkErrMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expectedSubs))

				for i, expected := range tt.expectedSubs {
					assert.Equal(t, expected.ID, result[i].ID)
					assert.Equal(t, expected.UserID, result[i].UserID)
					assert.Equal(t, expected.RepositoryID, result[i].RepositoryID)
					assert.Equal(t, expected.RepositoryName, result[i].RepositoryName)
					assert.Equal(t, expected.CreatedAt, result[i].CreatedAt)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSubscriptionRepository_GetEmailsByRepoID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &SubscriptionRepository{
		db:     mock,
		logger: discardLogger,
	}

	repoID := int64(777)

	tests := []struct {
		name        string
		repoID      int64
		mockSetup   func(id int64)
		expectError bool
		checkErrMsg string
		expected    []string
	}{
		{
			name:   "success with emails",
			repoID: repoID,
			mockSetup: func(id int64) {
				mock.ExpectQuery("SELECT u.email FROM users u").
					WithArgs(id).
					WillReturnRows(pgxmock.NewRows([]string{"email"}).
						AddRow("user1@example.com").
						AddRow("user2@test.com"))
			},
			expectError: false,
			expected:    []string{"user1@example.com", "user2@test.com"},
		},
		{
			name:   "no subscriptions for repo",
			repoID: repoID,
			mockSetup: func(id int64) {
				mock.ExpectQuery("SELECT u.email FROM users u").
					WithArgs(id).
					WillReturnRows(pgxmock.NewRows([]string{"email"}))
			},
			expectError: false,
			expected:    []string{},
		},
		{
			name:   "database error",
			repoID: repoID,
			mockSetup: func(id int64) {
				mock.ExpectQuery("SELECT u.email FROM users u").
					WithArgs(id).
					WillReturnError(fmt.Errorf("db failure"))
			},
			expectError: true,
			checkErrMsg: "query:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.repoID)

			result, err := repo.GetEmailsByRepoID(context.Background(), tt.repoID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.checkErrMsg != "" {
					assert.Contains(t, err.Error(), tt.checkErrMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
				assert.Len(t, result, len(tt.expected))
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
