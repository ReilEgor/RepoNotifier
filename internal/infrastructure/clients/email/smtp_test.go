package email

import (
	"context"
	"errors"
	"net/smtp"
	"testing"

	"github.com/ReilEgor/RepoNotifier/internal/domain/service"
	"github.com/stretchr/testify/assert"
)

func newTestClient(sendMail func(string, smtp.Auth, string, []string, []byte) error) *SmtpClient {
	c := NewSmtpClient("localhost", "25", "from@example.com", "pass", "user")
	c.sendMail = sendMail
	return c
}

func TestSmtpClient_SendNotification(t *testing.T) {
	tests := []struct {
		name        string
		to          string
		repoName    string
		tagName     string
		sendMailErr error
		wantErr     error
	}{
		{
			name:        "success",
			to:          "user@example.com",
			repoName:    "owner/repo",
			tagName:     "v1.0.0",
			sendMailErr: nil,
			wantErr:     nil,
		},
		{
			name:        "smtp auth failed (535)",
			to:          "user@example.com",
			repoName:    "owner/repo",
			tagName:     "v1.0.0",
			sendMailErr: errors.New("535 5.7.8 Authentication failed"),
			wantErr:     service.ErrAuthFailed,
		},
		{
			name:        "smtp auth failed (text match)",
			to:          "user@example.com",
			repoName:    "owner/repo",
			tagName:     "v1.0.0",
			sendMailErr: errors.New("Authentication failed: bad credentials"),
			wantErr:     service.ErrAuthFailed,
		},
		{
			name:        "smtp server unavailable",
			to:          "user@example.com",
			repoName:    "owner/repo",
			tagName:     "v1.0.0",
			sendMailErr: errors.New("connection refused"),
			wantErr:     service.ErrSMTPUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(func(_ string, _ smtp.Auth, _ string, _ []string, _ []byte) error {
				return tt.sendMailErr
			})

			err := client.SendNotification(context.Background(), tt.to, tt.repoName, tt.tagName)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestSmtpClient_buildMessage(t *testing.T) {
	tests := []struct {
		name             string
		to               string
		repoName         string
		tagName          string
		expectedSubject  string
		expectedBodyPart string
		expectedFrom     string
	}{
		{
			name:             "contains correct subject",
			to:               "user@example.com",
			repoName:         "owner/repo",
			tagName:          "v2.0.0",
			expectedSubject:  "Subject: New release in owner/repo!",
			expectedBodyPart: "A new version v2.0.0 has been released",
			expectedFrom:     "From: from@example.com",
		},
		{
			name:             "contains github release url",
			to:               "user@example.com",
			repoName:         "owner/repo",
			tagName:          "v3.0.0",
			expectedBodyPart: "https://github.com/owner/repo/releases/tag/v3.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSmtpClient("localhost", "25", "from@example.com", "pass", "user")
			msg := string(c.buildMessage(tt.to, tt.repoName, tt.tagName))

			if tt.expectedSubject != "" {
				assert.Contains(t, msg, tt.expectedSubject)
			}
			if tt.expectedBodyPart != "" {
				assert.Contains(t, msg, tt.expectedBodyPart)
			}
			if tt.expectedFrom != "" {
				assert.Contains(t, msg, tt.expectedFrom)
			}
		})
	}
}
