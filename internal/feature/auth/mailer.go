package auth

import (
	"context"
	"fmt"
	"github.com/Fitnow08/fitnow-backend-sso/internal/config"
	"gopkg.in/gomail.v2"
	"log/slog"
)

type Mailer struct {
	log    *slog.Logger
	dialer *gomail.Dialer
	from   string
}

func NewMailer(log *slog.Logger, cfg config.Mail) *Mailer {
	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Password)
	d.SSL = cfg.SSL
	return &Mailer{log: log, dialer: d, from: cfg.From}
}

func (m *Mailer) SendVerifyCode(ctx context.Context, to string, code int) error {
	const op = "Mailer.SendVerifyCode"
	log := m.log.With("op", op, "to", to)

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "🔐 Код подтверждения доступа")
	msg.SetBody("text/html", renderVerifyCodeHTML(code))

	done := make(chan error, 1)
	go func() { done <- m.dialer.DialAndSend(msg) }()

	select {
	case err := <-done:
		if err != nil {
			log.Error("failed to send mail", "error", err)
			return fmt.Errorf("%s: %w", op, err)
		}
		log.Info("verification code sent")
		return nil
	case <-ctx.Done():
		log.Warn("context cancelled while sending mail", "error", ctx.Err())
		return ctx.Err()
	}
}

func renderVerifyCodeHTML(code int) string {
	return fmt.Sprintf(`
	<html>
		<body style="font-family: sans-serif; color: #333;">
			<h2 style="color: #2c3e50;">Здравствуйте!</h2>
			<p>Ваш код подтверждения для входа:</p>
			<h1 style="color: #2980b9;">%d</h1>
			<p>Пожалуйста, введите этот код в приложении. Он действителен в течение нескольких минут.</p>
			<br/>
			<p style="font-size: 0.9em; color: #888;">Если вы не запрашивали код, просто проигнорируйте это письмо.</p>
			<p>С уважением,<br/>Команда FitNow</p>
		</body>
	</html>`, code)
}
