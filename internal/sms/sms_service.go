package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type SMSService struct {
	apiURL    string
	apiKey    string
	senderID  string
	client    *http.Client
	logger    *zap.Logger
}

type SMSConfig struct {
	APIURL   string
	APIKey   string
	SenderID string
}

type SMSRequest struct {
	To       string `json:"to"`
	From     string `json:"from"`
	Message  string `json:"message"`
}

type SMSResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id"`
}

func NewSMSService(cfg SMSConfig, logger *zap.Logger) *SMSService {
	return &SMSService{
		apiURL:   cfg.APIURL,
		apiKey:   cfg.APIKey,
		senderID: cfg.SenderID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// SendSecurityCode sends a security code via SMS
func (s *SMSService) SendSecurityCode(to, code string) error {
	message := fmt.Sprintf("Your ZIMRA Fiscalization security code is: %s. Valid for 15 minutes.", code)
	return s.sendSMS(to, message)
}

// SendPasswordReset sends password reset code via SMS
func (s *SMSService) SendPasswordReset(to, code string) error {
	message := fmt.Sprintf("Your ZIMRA password reset code is: %s. Valid for 15 minutes.", code)
	return s.sendSMS(to, message)
}

// SendWelcomeMessage sends welcome SMS to new user
func (s *SMSService) SendWelcomeMessage(to, username string) error {
	message := fmt.Sprintf("Welcome to ZIMRA Fiscalization, %s! Your account is now active.", username)
	return s.sendSMS(to, message)
}

// SendFiscalDayAlert sends fiscal day closing alert
func (s *SMSService) SendFiscalDayAlert(to, deviceName string, hoursLeft int) error {
	message := fmt.Sprintf("ZIMRA Alert: Fiscal day for %s closes in %d hour(s).", deviceName, hoursLeft)
	return s.sendSMS(to, message)
}

// sendSMS sends an SMS using the configured SMS gateway
func (s *SMSService) sendSMS(to, message string) error {
	// Prepare request
	smsReq := SMSRequest{
		To:      to,
		From:    s.senderID,
		Message: message,
	}

	jsonData, err := json.Marshal(smsReq)
	if err != nil {
		s.logger.Error("Failed to marshal SMS request", zap.Error(err))
		return err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.Error("Failed to create HTTP request", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("Failed to send SMS request",
			zap.String("to", to),
			zap.Error(err),
		)
		return err
	}
	defer resp.Body.Close()

	// Parse response
	var smsResp SMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
		s.logger.Error("Failed to decode SMS response", zap.Error(err))
		return err
	}

	if !smsResp.Success {
		s.logger.Error("SMS sending failed",
			zap.String("to", to),
			zap.String("error", smsResp.Message),
		)
		return fmt.Errorf("SMS sending failed: %s", smsResp.Message)
	}

	s.logger.Info("SMS sent successfully",
		zap.String("to", to),
		zap.String("id", smsResp.ID),
	)

	return nil
}

// MockSMSService for testing/development
type MockSMSService struct {
	logger *zap.Logger
}

func NewMockSMSService(logger *zap.Logger) *MockSMSService {
	return &MockSMSService{logger: logger}
}

func (s *MockSMSService) SendSecurityCode(to, code string) error {
	s.logger.Info("MOCK SMS: Security Code",
		zap.String("to", to),
		zap.String("code", code),
	)
	return nil
}

func (s *MockSMSService) SendPasswordReset(to, code string) error {
	s.logger.Info("MOCK SMS: Password Reset",
		zap.String("to", to),
		zap.String("code", code),
	)
	return nil
}

func (s *MockSMSService) SendWelcomeMessage(to, username string) error {
	s.logger.Info("MOCK SMS: Welcome",
		zap.String("to", to),
		zap.String("username", username),
	)
	return nil
}

func (s *MockSMSService) SendFiscalDayAlert(to, deviceName string, hoursLeft int) error {
	s.logger.Info("MOCK SMS: Fiscal Day Alert",
		zap.String("to", to),
		zap.String("device", deviceName),
		zap.Int("hoursLeft", hoursLeft),
	)
	return nil
}
