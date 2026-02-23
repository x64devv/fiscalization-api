package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Taxpayer represents a company/organization
type Taxpayer struct {
	ID                           int64     `json:"id" db:"id"`
	TIN                          string    `json:"tin" db:"tin"`
	Name                         string    `json:"name" db:"name"`
	VATNumber                    *string   `json:"vat_number,omitempty" db:"vat_number"`
	Status                       string    `json:"status" db:"status"` // Active, Inactive
	TaxPayerDayMaxHrs            int       `json:"taxpayer_day_max_hrs" db:"taxpayer_day_max_hrs"`
	TaxpayerDayEndNotificationHrs int      `json:"taxpayer_day_end_notification_hrs" db:"taxpayer_day_end_notification_hrs"`
	QrURL                        string    `json:"qr_url" db:"qr_url"`
	CreatedAt                    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at" db:"updated_at"`
}

// Address represents a physical address
type Address struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Street   string `json:"street"`
	HouseNo  string `json:"houseNo"`
}

// Value implements driver.Valuer for Address
func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements sql.Scanner for Address
func (a *Address) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Address: not a byte slice")
	}

	return json.Unmarshal(bytes, a)
}

// Contacts represents contact information
type Contacts struct {
	PhoneNo *string `json:"phoneNo,omitempty"`
	Email   *string `json:"email,omitempty"`
}

// Value implements driver.Valuer for Contacts
func (c Contacts) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner for Contacts
func (c *Contacts) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Contacts: not a byte slice")
	}

	return json.Unmarshal(bytes, c)
}

// SignatureData represents signature information
type SignatureData struct {
	Hash      []byte `json:"hash"`      // SHA-256 hash (32 bytes)
	Signature []byte `json:"signature"` // Cryptographic signature (variable length)
}

// Value implements driver.Valuer for SignatureData
func (s SignatureData) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner for SignatureData
func (s *SignatureData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan SignatureData: not a byte slice")
	}

	return json.Unmarshal(bytes, s)
}

// SignatureDataEx extends SignatureData with certificate thumbprint
type SignatureDataEx struct {
	SignatureData
	CertificateThumbprint []byte `json:"certificateThumbprint"` // SHA-1 thumbprint (20 bytes)
}

// Value implements driver.Valuer for SignatureDataEx
func (s SignatureDataEx) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner for SignatureDataEx
func (s *SignatureDataEx) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan SignatureDataEx: not a byte slice")
	}

	return json.Unmarshal(bytes, s)
}
/*
// APIError represents an API error response
type APIError struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	ErrorCode string `json:"errorCode,omitempty"`
}

// Error implements the error interface
func (e APIError) Error() string {
	return e.Title
}

// Common error codes

// NewAPIError creates a new API error
func NewAPIError(status int, title string, errorCode string) *APIError {
	return &APIError{
		Type:      "about:blank",
		Title:     title,
		Status:    status,
		ErrorCode: errorCode,
	}
}
*/
// OperationResponse represents a standard operation response
type OperationResponse struct {
	OperationID string `json:"operationID"`
}
