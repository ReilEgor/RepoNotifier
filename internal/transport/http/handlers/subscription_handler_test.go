package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ReilEgor/RepoNotifier/internal/domain/model"
	"github.com/ReilEgor/RepoNotifier/internal/domain/service"
	"github.com/ReilEgor/RepoNotifier/internal/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionHandler_Subscribe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(uc *mocks.SubscriptionUseCase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success: subscription created",
			requestBody: map[string]string{
				"email":      "test@example.com",
				"repository": "golang/go",
			},
			mockSetup: func(uc *mocks.SubscriptionUseCase) {
				uc.On("Subscribe", mock.Anything, "test@example.com", "golang/go").
					Return(int64(123), nil).Once()
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "error: invalid email format",
			requestBody: map[string]string{
				"email":      "bad-email",
				"repository": "golang/go",
			},
			mockSetup:      func(uc *mocks.SubscriptionUseCase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "error: repository not found on github",
			requestBody: map[string]string{
				"email":      "test@example.com",
				"repository": "owner/unknown",
			},
			mockSetup: func(uc *mocks.SubscriptionUseCase) {
				uc.On("Subscribe", mock.Anything, "test@example.com", "owner/unknown").
					Return(int64(0), service.ErrRepositoryNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "error: already subscribed",
			requestBody: map[string]string{
				"email":      "test@example.com",
				"repository": "golang/go",
			},
			mockSetup: func(uc *mocks.SubscriptionUseCase) {
				uc.On("Subscribe", mock.Anything, "test@example.com", "golang/go").
					Return(int64(0), ErrAlreadySubscribed).Once()
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "error: github api unavailable",
			requestBody: map[string]string{
				"email":      "test@example.com",
				"repository": "golang/go",
			},
			mockSetup: func(uc *mocks.SubscriptionUseCase) {
				uc.On("Subscribe", mock.Anything, "test@example.com", "golang/go").
					Return(int64(0), service.ErrGitHubUnavailable).Once()
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "error: internal server error",
			requestBody: map[string]string{
				"email":      "test@example.com",
				"repository": "golang/go",
			},
			mockSetup: func(uc *mocks.SubscriptionUseCase) {
				uc.On("Subscribe", mock.Anything, "test@example.com", "golang/go").
					Return(int64(0), errors.New("db connection lost")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewSubscriptionUseCase(t)
			tt.mockSetup(mockUC)
			h := &Handler{subscriptionUC: mockUC}

			r := gin.New()
			r.POST("/subscriptions", h.Subscribe)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestSubscriptionHandler_Unsubscribe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(uc *mocks.SubscriptionUseCase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success: unsubscribed",
			requestBody: map[string]string{
				"email":      "user@test.com",
				"repository": "golang/go",
			},
			mockSetup: func(uc *mocks.SubscriptionUseCase) {
				uc.On("Unsubscribe", mock.Anything, "user@test.com", "golang/go").
					Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   msgSubscriptionDeleted,
		},
		{
			name: "error: validation failed",
			requestBody: map[string]string{
				"email":      "not-an-email",
				"repository": "golang/go",
			},
			mockSetup:      func(uc *mocks.SubscriptionUseCase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name: "error: subscription not found",
			requestBody: map[string]string{
				"email":      "user@test.com",
				"repository": "owner/repo",
			},
			mockSetup: func(uc *mocks.SubscriptionUseCase) {
				uc.On("Unsubscribe", mock.Anything, "user@test.com", "owner/repo").
					Return(service.ErrSubscriptionNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "error: internal server error",
			requestBody: map[string]string{
				"email":      "user@test.com",
				"repository": "golang/go",
			},
			mockSetup: func(uc *mocks.SubscriptionUseCase) {
				uc.On("Unsubscribe", mock.Anything, "user@test.com", "golang/go").
					Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   errFailedToUnsubscribe,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewSubscriptionUseCase(t)
			tt.mockSetup(mockUC)
			h := &Handler{subscriptionUC: mockUC}

			r := gin.New()
			r.DELETE("/subscriptions", h.Unsubscribe)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodDelete, "/subscriptions", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestSubscriptionHandler_ListSubscriptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryEmail     string
		mockSetup      func(uc *mocks.SubscriptionUseCase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "success: list returned",
			queryEmail: "test@example.com",
			mockSetup: func(uc *mocks.SubscriptionUseCase) {
				expectedSubs := []model.Subscription{
					{ID: 1, RepositoryID: 101},
					{ID: 2, RepositoryID: 102},
				}
				uc.On("ListByEmail", mock.Anything, "test@example.com").
					Return(expectedSubs, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"total":2`,
		},
		{
			name:           "error: email missing",
			queryEmail:     "",
			mockSetup:      func(uc *mocks.SubscriptionUseCase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrEmailRequired.Error(),
		},
		{
			name:           "error: invalid email format",
			queryEmail:     "not-an-email",
			mockSetup:      func(uc *mocks.SubscriptionUseCase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrInvalidEmailFormat.Error(),
		},
		{
			name:       "error: internal server error",
			queryEmail: "test@example.com",
			mockSetup: func(uc *mocks.SubscriptionUseCase) {
				uc.On("ListByEmail", mock.Anything, "test@example.com").
					Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   errFailedToList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewSubscriptionUseCase(t)
			tt.mockSetup(mockUC)
			h := &Handler{subscriptionUC: mockUC}

			r := gin.New()
			r.GET("/subscriptions", h.ListSubscriptions)

			url := "/subscriptions"
			if tt.queryEmail != "" {
				url = fmt.Sprintf("%s?email=%s", url, tt.queryEmail)
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}
