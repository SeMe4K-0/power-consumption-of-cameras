package email

import (
	"awesomeProject/internal/app/config"
	"fmt"
	"net/smtp"
	"strings"
)

type Service struct {
	cfg config.SMTPConfig
}

func NewEmailService(cfg config.SMTPConfig) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) SendVerificationCode(to, code string) error {
	subject := "Код подтверждения для входа"
	body := fmt.Sprintf(`Здравствуйте!
Ваш код подтверждения для входа: %s
Код действителен в течение 5 минут.
Если вы не запрашивали этот код, проигнорируйте это письмо.
`, code)

	msg := []byte(fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body)

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	toList := []string{to}
	err := smtp.SendMail(addr, auth, s.cfg.From, toList, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *Service) ValidateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
