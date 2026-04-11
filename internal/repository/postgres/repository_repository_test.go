package postgres

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryRepository_GetByName(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &RepositoryRepository{
		db:     mock,
		logger: discardLogger,
	}

	now := time.Now()

	tests := []struct {
		name        string
		repoName    string
		mockSetup   func()
		expectError bool
		errorIs     error
		checkErrMsg string
		expected    *model.Repository
	}{
		{
			name:     "success",
			repoName: "golang/go",
			mockSetup: func() {
				mock.ExpectQuery("SELECT (.+) FROM repositories WHERE full_name = \\$1").
					WithArgs("golang/go").
					WillReturnRows(pgxmock.NewRows([]string{"id", "full_name", "last_seen_tag", "updated_at"}).
						AddRow(int64(1), "golang/go", "v1.25.0", now))
			},
			expectError: false,
			expected: &model.Repository{
				ID:          1,
				FullName:    "golang/go",
				LastSeenTag: "v1.25.0",
				UpdatedAt:   now,
			},
		},
		{
			name:     "not found",
			repoName: "unknown/repo",
			mockSetup: func() {
				mock.ExpectQuery("SELECT (.+) FROM repositories WHERE full_name = \\$1").
					WithArgs("unknown/repo").
					WillReturnError(pgx.ErrNoRows)
			},
			expectError: true,
			errorIs:     ErrRepositoryNotFound,
		},
		{
			name:     "database error",
			repoName: "any/repo",
			mockSetup: func() {
				mock.ExpectQuery("SELECT (.+) FROM repositories").
					WithArgs("any/repo").
					WillReturnError(pgx.ErrTxClosed)
			},
			expectError: true,
			checkErrMsg: "query row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.GetByName(context.Background(), tt.repoName)

			if tt.expectError {
				assert.Error(t, err)

				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}

				if tt.checkErrMsg != "" {
					assert.Contains(t, err.Error(), tt.checkErrMsg)
				}

				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.FullName, result.FullName)
				assert.Equal(t, tt.expected.LastSeenTag, result.LastSeenTag)

				assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepositoryRepository_GetAll(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &RepositoryRepository{
		db:     mock,
		logger: discardLogger,
	}

	now := time.Now()

	tests := []struct {
		name          string
		mockSetup     func()
		expectError   bool
		checkErrMsg   string
		expectedLen   int
		expectedRepos []model.Repository
	}{
		{
			name: "success with multiple repos",
			mockSetup: func() {
				mock.ExpectQuery("SELECT (.+) FROM repositories").
					WillReturnRows(pgxmock.NewRows([]string{"id", "full_name", "last_seen_tag", "updated_at"}).
						AddRow(int64(1), "golang/go", "v1.25.0", now).
						AddRow(int64(2), "google/wire", "v0.6.0", now))
			},
			expectError: false,
			expectedLen: 2,
			expectedRepos: []model.Repository{
				{ID: 1, FullName: "golang/go", LastSeenTag: "v1.25.0", UpdatedAt: now},
				{ID: 2, FullName: "google/wire", LastSeenTag: "v0.6.0", UpdatedAt: now},
			},
		},
		{
			name: "success empty list",
			mockSetup: func() {
				mock.ExpectQuery("SELECT (.+) FROM repositories").
					WillReturnRows(pgxmock.NewRows([]string{"id", "full_name", "last_seen_tag", "updated_at"}))
			},
			expectError:   false,
			expectedLen:   0,
			expectedRepos: []model.Repository{},
		},
		{
			name: "database query error",
			mockSetup: func() {
				mock.ExpectQuery("SELECT (.+) FROM repositories").
					WillReturnError(fmt.Errorf("db connection lost"))
			},
			expectError: true,
			checkErrMsg: "query:",
		},
		{
			name: "row scan error",
			mockSetup: func() {
				mock.ExpectQuery("SELECT (.+) FROM repositories").
					WillReturnRows(pgxmock.NewRows([]string{"id", "full_name", "last_seen_tag", "updated_at"}).
						AddRow("not-an-id", "owner/repo", "v1", now))
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
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectedLen)

				for i, expected := range tt.expectedRepos {
					assert.Equal(t, expected.ID, result[i].ID)
					assert.Equal(t, expected.FullName, result[i].FullName)
					assert.Equal(t, expected.LastSeenTag, result[i].LastSeenTag)
					assert.Equal(t, expected.UpdatedAt, result[i].UpdatedAt)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepositoryRepository_UpdateLastSeenTag(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &RepositoryRepository{
		db:     mock,
		logger: discardLogger,
	}

	tests := []struct {
		name        string
		repoName    string
		tag         string
		mockSetup   func(repoName, tag string)
		expectError bool
		checkErrMsg string
	}{
		{
			name:     "success update",
			repoName: "golang/go",
			tag:      "v1.26.0",
			mockSetup: func(repoName, tag string) {
				mock.ExpectExec("UPDATE repositories").
					WithArgs(tag, repoName).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectError: false,
		},
		{
			name:     "database error on exec",
			repoName: "golang/go",
			tag:      "v1.26.0",
			mockSetup: func(repoName, tag string) {
				mock.ExpectExec("UPDATE repositories").
					WithArgs(tag, repoName).
					WillReturnError(fmt.Errorf("connection timeout"))
			},
			expectError: true,
			checkErrMsg: "exec:",
		},
		{
			name:     "repository not found (no rows affected)",
			repoName: "non-existent",
			tag:      "v1",
			mockSetup: func(repoName, tag string) {
				mock.ExpectExec("UPDATE repositories").
					WithArgs(tag, repoName).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.repoName, tt.tag)

			err := repo.UpdateLastSeenTag(context.Background(), tt.repoName, tt.tag)

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

func TestRepositoryRepository_GetOrCreate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &RepositoryRepository{
		db:     mock,
		logger: discardLogger,
	}

	repoName := "golang/go"
	tagName := "v1.25.0"
	now := time.Now()

	tests := []struct {
		name        string
		mockSetup   func()
		expectError bool
		checkErrMsg string
	}{
		{
			name: "success create new",
			mockSetup: func() {
				mock.ExpectQuery("INSERT INTO repositories").
					WithArgs(repoName, tagName).
					WillReturnRows(pgxmock.NewRows([]string{"id", "full_name", "last_seen_tag", "updated_at"}).
						AddRow(int64(1), repoName, tagName, now))
			},
			expectError: false,
		},
		{
			name: "success get existing (on conflict)",
			mockSetup: func() {
				mock.ExpectQuery("INSERT INTO repositories").
					WithArgs(repoName, tagName).
					WillReturnRows(pgxmock.NewRows([]string{"id", "full_name", "last_seen_tag", "updated_at"}).
						AddRow(int64(1), repoName, "old-tag", now))
			},
			expectError: false,
		},
		{
			name: "database error",
			mockSetup: func() {
				mock.ExpectQuery("INSERT INTO repositories").
					WithArgs(repoName, tagName).
					WillReturnError(fmt.Errorf("unique constraint violation somehow failed"))
			},
			expectError: true,
			checkErrMsg: "query row:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.GetOrCreate(context.Background(), repoName, tagName)

			if tt.expectError {
				assert.Error(t, err)
				if tt.checkErrMsg != "" {
					assert.Contains(t, err.Error(), tt.checkErrMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, repoName, result.FullName)
				assert.Equal(t, int64(1), result.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
