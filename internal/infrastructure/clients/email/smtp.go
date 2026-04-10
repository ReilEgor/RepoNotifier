package email

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/ReilEgor/RepoNotifier/internal/config"
	"github.com/ReilEgor/RepoNotifier/internal/domain/service"
)

const (
	componentEmailClient = "EmailClient"

	emailSubjectTemplate = "New release in %s!"
	emailBodyTemplate    = "Hello!\n\nA new version %s has been released for the repository %s.\n" +
		"Check it out here: https://github.com/%s/releases/tag/%s\n\nBest regards,\nRepoNotifier"
	emailMsgTemplate = "From: %s\r\n" +
		"To: %s\r\n" +
		"Subject: %s\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		"%s\r\n"
)

const (
	errMsgSendMail = "failed to send email"
	errMsgBuildMsg = "failed to build message"
)

func (c *SmtpClient) buildMessage(to, repoName, tagName string) []byte {
	subject := fmt.Sprintf(emailSubjectTemplate, repoName)
	body := fmt.Sprintf(emailBodyTemplate, tagName, repoName, repoName, tagName)
	return []byte(fmt.Sprintf(emailMsgTemplate, c.from, to, subject, body))
}

type SmtpClient struct {
	host     config.EmailHostType
	port     config.EmailPortType
	from     config.EmailFromType
	auth     smtp.Auth
	sendMail func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
	logger   *slog.Logger
}

func NewSmtpClient(
	host config.EmailHostType,
	port config.EmailPortType,
	from config.EmailFromType,
	password config.EmailPasswordType,
	user config.EmailUserType,
) *SmtpClient {
	return &SmtpClient{
		host:     host,
		port:     port,
		from:     from,
		auth:     smtp.PlainAuth("", string(user), string(password), string(host)),
		logger:   slog.With(slog.String("component", componentEmailClient)),
		sendMail: smtp.SendMail,
	}
}

func (c *SmtpClient) SendNotification(ctx context.Context, to string, repoName string, tagName string) error {
	const op = "SmtpClient.SendNotification"
	log := c.logger.With(slog.String("op", op))

	msg := c.buildMessage(to, repoName, tagName)
	addr := fmt.Sprintf("%s:%s", c.host, c.port)
	if err := c.sendMail(addr, c.auth, string(c.from), []string{to}, msg); err != nil {
		if strings.Contains(err.Error(), "535") || strings.Contains(err.Error(), "Authentication failed") {
			log.ErrorContext(ctx, "smtp auth failed", slog.String("error", err.Error()))
			return fmt.Errorf("%s: %w", op, service.ErrAuthFailed)
		}
		log.ErrorContext(ctx, errMsgSendMail,
			slog.String("to", to),
			slog.String("repo", repoName),
			slog.String("tag", tagName),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%s: %w", op, service.ErrSMTPUnavailable)
	}

	log.InfoContext(ctx, "email sent successfully",
		slog.String("to", to),
		slog.String("repo", repoName),
		slog.String("tag", tagName),
	)
	return nil
}
