package email

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/ReilEgor/RepoNotifier/internal/config"
)

type SmtpClient struct {
	host   config.EmailHostType
	port   config.EmailPortType
	from   config.EmailFromType
	auth   smtp.Auth
	logger *slog.Logger
}

func NewSmtpClient(
	host config.EmailHostType,
	port config.EmailPortType,
	from config.EmailFromType,
	password config.EmailPasswordType,
	user config.EmailUserType,
) *SmtpClient {
	return &SmtpClient{
		host:   host,
		port:   port,
		from:   from,
		auth:   smtp.PlainAuth("", string(user), string(password), string(host)),
		logger: slog.With(slog.String("component", "EmailClient")),
	}
}

func (c *SmtpClient) SendNotification(ctx context.Context, to string, repoName string, tagName string) error {
	const op = "EmailClient.SendNotification"

	subject := fmt.Sprintf("New release in %s!", repoName)
	body := fmt.Sprintf(
		"Hello!\n\nA new version %s has been released for the repository %s.\nCheck it out here: https://github.com/%s/releases/tag/%s\n\nBest regards,\nRepoNotifier",
		tagName, repoName, repoName, tagName,
	)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n", c.from, to, subject, body))

	addr := fmt.Sprintf("%s:%s", c.host, c.port)

	err := smtp.SendMail(addr, c.auth, string(c.from), []string{to}, msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	c.logger.Info("email sent successfully", slog.String("to", to), slog.String("repo", repoName))
	return nil
}
