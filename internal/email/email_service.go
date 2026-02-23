package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"go.uber.org/zap"
)

type EmailService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromAddress  string
	logger       *zap.Logger
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromAddress  string
}

func NewEmailService(cfg EmailConfig, logger *zap.Logger) *EmailService {
	return &EmailService{
		smtpHost:     cfg.SMTPHost,
		smtpPort:     cfg.SMTPPort,
		smtpUsername: cfg.SMTPUsername,
		smtpPassword: cfg.SMTPPassword,
		fromAddress:  cfg.FromAddress,
		logger:       logger,
	}
}

// SendSecurityCode sends a security code via email
func (s *EmailService) SendSecurityCode(to, code, username string) error {
	subject := "ZIMRA Fiscalization - Security Code"
	body := fmt.Sprintf(`
Dear %s,

Your security code for ZIMRA Fiscalization system is:

%s

This code will expire in 15 minutes.

If you did not request this code, please ignore this email.

Best regards,
ZIMRA Fiscalization System
`, username, code)

	return s.sendEmail(to, subject, body)
}

// SendPasswordReset sends password reset code via email
func (s *EmailService) SendPasswordReset(to, code, username string) error {
	subject := "ZIMRA Fiscalization - Password Reset"
	body := fmt.Sprintf(`
Dear %s,

You have requested to reset your password.

Your password reset code is:

%s

This code will expire in 15 minutes.

If you did not request this, please ignore this email and your password will remain unchanged.

Best regards,
ZIMRA Fiscalization System
`, username, code)

	return s.sendEmail(to, subject, body)
}

// SendWelcomeEmail sends welcome email to new user
func (s *EmailService) SendWelcomeEmail(to, username string) error {
	subject := "Welcome to ZIMRA Fiscalization System"
	body := fmt.Sprintf(`
Dear %s,

Welcome to the ZIMRA Fiscalization System!

Your account has been successfully created and activated.

You can now log in using your username and password.

If you have any questions, please contact support.

Best regards,
ZIMRA Fiscalization System
`, username)

	return s.sendEmail(to, subject, body)
}

// SendFiscalDayCloseNotification sends notification when fiscal day is about to close
func (s *EmailService) SendFiscalDayCloseNotification(to, deviceName string, hoursLeft int) error {
	subject := "ZIMRA Fiscalization - Fiscal Day Closing Soon"
	body := fmt.Sprintf(`
Dear User,

This is a reminder that the fiscal day for device "%s" will close in %d hour(s).

Please ensure all receipts are submitted before the fiscal day closes.

Best regards,
ZIMRA Fiscalization System
`, deviceName, hoursLeft)

	return s.sendEmail(to, subject, body)
}

// sendEmail sends an email using SMTP
func (s *EmailService) sendEmail(to, subject, body string) error {
	// Build message
	message := fmt.Sprintf("From: %s\r\n", s.fromAddress)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "\r\n"
	message += body

	// Setup authentication
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Connect to SMTP server
	serverAddr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)

	// Setup TLS config
	tlsConfig := &tls.Config{
		ServerName:         s.smtpHost,
		InsecureSkipVerify: false,
	}

	// Connect and send
	conn, err := tls.Dial("tcp", serverAddr, tlsConfig)
	if err != nil {
		s.logger.Error("Failed to connect to SMTP server",
			zap.String("server", serverAddr),
			zap.Error(err),
		)
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		s.logger.Error("Failed to create SMTP client", zap.Error(err))
		return err
	}
	defer client.Quit()

	// Authenticate
	if err := client.Auth(auth); err != nil {
		s.logger.Error("SMTP authentication failed", zap.Error(err))
		return err
	}

	// Set sender and recipient
	if err := client.Mail(s.fromAddress); err != nil {
		s.logger.Error("Failed to set sender", zap.Error(err))
		return err
	}

	if err := client.Rcpt(to); err != nil {
		s.logger.Error("Failed to set recipient", zap.Error(err))
		return err
	}

	// Send message
	w, err := client.Data()
	if err != nil {
		s.logger.Error("Failed to get data writer", zap.Error(err))
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		s.logger.Error("Failed to write message", zap.Error(err))
		return err
	}

	err = w.Close()
	if err != nil {
		s.logger.Error("Failed to close writer", zap.Error(err))
		return err
	}

	s.logger.Info("Email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject),
	)

	return nil
}

// MockEmailService for testing/development
type MockEmailService struct {
	logger *zap.Logger
}

func NewMockEmailService(logger *zap.Logger) *MockEmailService {
	return &MockEmailService{logger: logger}
}

func (s *MockEmailService) SendSecurityCode(to, code, username string) error {
	s.logger.Info("MOCK EMAIL: Security Code",
		zap.String("to", to),
		zap.String("code", code),
		zap.String("username", username),
	)
	return nil
}

func (s *MockEmailService) SendPasswordReset(to, code, username string) error {
	s.logger.Info("MOCK EMAIL: Password Reset",
		zap.String("to", to),
		zap.String("code", code),
		zap.String("username", username),
	)
	return nil
}

func (s *MockEmailService) SendWelcomeEmail(to, username string) error {
	s.logger.Info("MOCK EMAIL: Welcome",
		zap.String("to", to),
		zap.String("username", username),
	)
	return nil
}

func (s *MockEmailService) SendFiscalDayCloseNotification(to, deviceName string, hoursLeft int) error {
	s.logger.Info("MOCK EMAIL: Fiscal Day Closing",
		zap.String("to", to),
		zap.String("device", deviceName),
		zap.Int("hoursLeft", hoursLeft),
	)
	return nil
}
