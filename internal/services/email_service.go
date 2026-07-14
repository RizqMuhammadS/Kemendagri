package services

import (
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"

	"github.com/yourusername/meeting-minutes-ai/internal/config"
)

// EmailService handles sending meeting minutes via email
type EmailService struct {
	cfg *config.Config
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

// SendMinutes sends meeting minutes to specified recipients
func (s *EmailService) SendMinutes(title string, recipients []string, attachmentPath string) error {
	if s.cfg.SMTPHost == "" || s.cfg.SMTPUser == "" {
		return fmt.Errorf("SMTP not configured")
	}

	subject := fmt.Sprintf("Notulensi Rapat: %s", title)
	body := fmt.Sprintf("Berikut terlampir notulensi rapat: %s\n\nDokumen ini dibuat secara otomatis oleh Sistem Notulensi AI.", title)

	// Construct email message with attachment
	msg, err := s.buildMessage(recipients, subject, body, attachmentPath)
	if err != nil {
		return fmt.Errorf("failed to build email message: %w", err)
	}

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	if err := smtp.SendMail(addr, auth, s.cfg.SMTPUser, recipients, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *EmailService) buildMessage(recipients []string, subject, body, attachmentPath string) ([]byte, error) {
	// Read attachment if provided
	var attachmentData []byte
	var err error

	to := ""
	for i, r := range recipients {
		if i > 0 {
			to += ", "
		}
		to += r
	}

	// Build multipart MIME message
	boundary := "boundary123"
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\n", s.cfg.SMTPUser, to, subject)

	if attachmentPath != "" {
		message += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary)
		message += fmt.Sprintf("--%s\r\n", boundary)
		message += "Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n"
		message += body + "\r\n\r\n"

		// Read attachment
		attachmentData, err = os.ReadFile(attachmentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read attachment: %w", err)
		}

		fileName := filepath.Base(attachmentPath)
		message += fmt.Sprintf("--%s\r\n", boundary)
		message += fmt.Sprintf("Content-Type: application/octet-stream\r\n")
		message += fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", fileName)
		message += "Content-Transfer-Encoding: base64\r\n\r\n"

		// Convert attachment to base64 (simplified - in production use encoding/base64)
		message += string(attachmentData) + "\r\n"
		message += fmt.Sprintf("--%s--\r\n", boundary)
	} else {
		message += "Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n"
		message += body + "\r\n"
	}

	return []byte(message), nil
}