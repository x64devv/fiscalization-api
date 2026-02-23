package models

import (
	"fmt"
)

// Error codes constants
const (
	// Device errors
	ErrCodeDEV01 = "DEV01" // Device not found or invalid
	ErrCodeDEV02 = "DEV02" // Activation key incorrect
	ErrCodeDEV03 = "DEV03" // Certificate request invalid
	ErrCodeDEV04 = "DEV04" // Device model blacklisted
	ErrCodeDEV05 = "DEV05" // Taxpayer not active
	ErrCodeDEV06 = "DEV06" // Device already registered
	ErrCodeDEV07 = "DEV07" // Certificate expired
	ErrCodeDEV08 = "DEV08" // Invalid certificate
	ErrCodeDEV09 = "DEV09" // Device blocked
	ErrCodeDEV10 = "DEV10" // Operating mode invalid

	// Fiscal day errors
	ErrCodeFISC01 = "FISC01" // Fiscal day already opened
	ErrCodeFISC02 = "FISC02" // Previous fiscal day not closed
	ErrCodeFISC03 = "FISC03" // No fiscal day to close
	ErrCodeFISC04 = "FISC04" // Fiscal day has validation errors

	// Receipt errors
	ErrCodeRCPT01 = "RCPT01" // No fiscal day opened
	ErrCodeRCPT02 = "RCPT02" // Receipt validation failed
	ErrCodeRCPT03 = "RCPT03" // Invalid signature
	ErrCodeRCPT04 = "RCPT04" // Duplicate receipt
	ErrCodeRCPT05 = "RCPT05" // Invalid receipt type
	ErrCodeRCPT06 = "RCPT06" // Invalid currency
	ErrCodeRCPT07 = "RCPT07" // Invalid total
	ErrCodeRCPT08 = "RCPT08" // Invalid tax
	ErrCodeRCPT09 = "RCPT09" // Invalid line item
	ErrCodeRCPT10 = "RCPT10" // Invalid payment

	// Receipt validation codes (RCPT010-RCPT048)
	ErrCodeRCPT010 = "RCPT010" // Currency not valid
	ErrCodeRCPT011 = "RCPT011" // Counter incorrect
	ErrCodeRCPT012 = "RCPT012" // Date in future
	ErrCodeRCPT013 = "RCPT013" // Date too old
	ErrCodeRCPT014 = "RCPT014" // Total calculation incorrect
	ErrCodeRCPT015 = "RCPT015" // Tax calculation incorrect
	ErrCodeRCPT016 = "RCPT016" // Payment total mismatch
	ErrCodeRCPT017 = "RCPT017" // Invalid buyer data
	ErrCodeRCPT018 = "RCPT018" // Missing receipt lines
	ErrCodeRCPT019 = "RCPT019" // Invalid HS code
	ErrCodeRCPT020 = "RCPT020" // Invalid tax ID
	ErrCodeRCPT021 = "RCPT021" // Tax not applicable
	ErrCodeRCPT022 = "RCPT022" // Line price invalid
	ErrCodeRCPT023 = "RCPT023" // Line quantity invalid
	ErrCodeRCPT024 = "RCPT024" // Line total incorrect
	ErrCodeRCPT025 = "RCPT025" // Discount exceeds total
	ErrCodeRCPT026 = "RCPT026" // VAT taxpayer required
	ErrCodeRCPT027 = "RCPT027" // Invoice number duplicate
	ErrCodeRCPT028 = "RCPT028" // Credit note amount exceeds original
	ErrCodeRCPT029 = "RCPT029" // Debit note amount exceeds limit
	ErrCodeRCPT030 = "RCPT030" // Original invoice not found
	ErrCodeRCPT031 = "RCPT031" // Original invoice invalid type
	ErrCodeRCPT032 = "RCPT032" // Credit/debit note currency mismatch
	ErrCodeRCPT033 = "RCPT033" // Previous hash missing
	ErrCodeRCPT034 = "RCPT034" // Previous hash invalid
	ErrCodeRCPT035 = "RCPT035" // Signature verification failed
	ErrCodeRCPT036 = "RCPT036" // Missing required field
	ErrCodeRCPT037 = "RCPT037" // Field length exceeded
	ErrCodeRCPT038 = "RCPT038" // Invalid character in field
	ErrCodeRCPT039 = "RCPT039" // Tax percent mismatch
	ErrCodeRCPT040 = "RCPT040" // Tax amount calculation error
	ErrCodeRCPT041 = "RCPT041" // Multiple tax rates on line
	ErrCodeRCPT042 = "RCPT042" // Tax valid from date invalid
	ErrCodeRCPT043 = "RCPT043" // Tax valid till date invalid
	ErrCodeRCPT044 = "RCPT044" // Receipt type invalid for operation
	ErrCodeRCPT045 = "RCPT045" // Line sequence invalid
	ErrCodeRCPT046 = "RCPT046" // Payment method invalid
	ErrCodeRCPT047 = "RCPT047" // Receipt counter sequence broken
	ErrCodeRCPT048 = "RCPT048" // Global counter sequence broken

	// File errors
	ErrCodeFILE01 = "FILE01" // File format invalid
	ErrCodeFILE02 = "FILE02" // File too large
	ErrCodeFILE03 = "FILE03" // File processing failed
	ErrCodeFILE04 = "FILE04" // File sent for closed day
	ErrCodeFILE05 = "FILE05" // File exceeded waiting time

	// User errors
	ErrCodeUSER01 = "USER01" // User not found
	ErrCodeUSER02 = "USER02" // Invalid credentials
	ErrCodeUSER03 = "USER03" // User not active
	ErrCodeUSER04 = "USER04" // Security code invalid
	ErrCodeUSER05 = "USER05" // Security code expired
	ErrCodeUSER06 = "USER06" // Username already exists
	ErrCodeUSER07 = "USER07" // Password too weak
	ErrCodeUSER08 = "USER08" // Token invalid
	ErrCodeUSER09 = "USER09" // Token expired
	ErrCodeUSER10 = "USER10" // Insufficient permissions


	ErrCodeDEV11 = "DEV11" // User credentials are incorrect
	ErrCodeDEV12 = "DEV12" // Token is not valid
	ErrCodeDEV13 = "DEV13" // User is not confirmed
	ErrCodeDEV14 = "DEV14" // Email or phone number is not valid or doesn't exist
	ErrCodeDEV15 = "DEV15" // Email or phone number already confirmed

)

// APIError represents an API error response
type APIError struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	ErrorCode string `json:"errorCode,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

// NewAPIError creates a new API error
func NewAPIError(status int, title string, errorCode string) *APIError {
	return &APIError{
		Type:      "about:blank",
		Title:     title,
		Status:    status,
		ErrorCode: errorCode,
	}
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.ErrorCode != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Title, e.Detail, e.ErrorCode)
	}
	return fmt.Sprintf("%s: %s", e.Title, e.Detail)
}

// WithDetail adds detail to the error
func (e *APIError) WithDetail(detail string) *APIError {
	e.Detail = detail
	return e
}
