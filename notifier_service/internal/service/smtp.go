package service

import (
	"fmt"
	"strconv"
	"errors"

	"notifier_service/internal/config"
	"notifier_service/pkg/log"

	"gopkg.in/gomail.v2"
)

var ErrSendEmail = errors.New("Failed to send email")

type SMTPService struct {
	dialer *gomail.Dialer
	from string
}

func NewSMTPService() *SMTPService {
	port, err := strconv.Atoi(config.SmtpPort)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("Invalid SMTP port:")
	}

	dialer := gomail.NewDialer(config.SmtpHost, port, "", "")
	
	return &SMTPService{
		dialer: dialer,
		from: config.SmtpFrom,
	}
}

func (s *SMTPService) Send(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetHeader("text/plain", body)

	err := s.dialer.DialAndSend(m) 
	if err != nil {
		return fmt.Errorf("failed to send email %w", err)
	}

	log.Logger.Info().Str("to", to).Str("subject", subject).Msg("Email sent")
	return nil
}