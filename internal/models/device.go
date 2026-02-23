package models

import (
	"time"
)

// Device represents a fiscal device
type Device struct {
	ID                  int64               `json:"id" db:"id"`
	DeviceID            int                 `json:"device_id" db:"device_id"`
	TaxpayerID          int64               `json:"taxpayer_id" db:"taxpayer_id"`
	DeviceSerialNo      string              `json:"device_serial_no" db:"device_serial_no"`
	DeviceModelName     string              `json:"device_model_name" db:"device_model_name"`
	DeviceModelVersion  string              `json:"device_model_version" db:"device_model_version"`
	ActivationKey       string              `json:"-" db:"activation_key"` // Don't expose in JSON
	Certificate         string              `json:"certificate,omitempty" db:"certificate"`
	CertificateThumbprint []byte            `json:"-" db:"certificate_thumbprint"`
	CertificateValidTill time.Time          `json:"certificate_valid_till" db:"certificate_valid_till"`
	OperatingMode       DeviceOperatingMode `json:"operating_mode" db:"operating_mode"`
	Status              string              `json:"status" db:"status"` // Active, Blocked, Revoked
	BranchName          string              `json:"branch_name" db:"branch_name"`
	BranchAddress       Address             `json:"branch_address" db:"branch_address"`
	BranchContacts      *Contacts           `json:"branch_contacts,omitempty" db:"branch_contacts"`
	CreatedAt           time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at" db:"updated_at"`
}

// DeviceRegistrationRequest represents device registration request
type DeviceRegistrationRequest struct {
	DeviceID           int    `json:"deviceID" binding:"required"`
	ActivationKey      string `json:"activationKey" binding:"required,len=8"`
	CertificateRequest string `json:"certificateRequest" binding:"required"`
}

// DeviceRegistrationResponse represents device registration response
type DeviceRegistrationResponse struct {
	OperationID string `json:"operationID"`
	Certificate string `json:"certificate"`
}

// VerifyTaxpayerRequest represents taxpayer verification request
type VerifyTaxpayerRequest struct {
	DeviceID       int    `json:"deviceID" binding:"required"`
	ActivationKey  string `json:"activationKey" binding:"required,len=8"`
	DeviceSerialNo string `json:"deviceSerialNo" binding:"required,max=20"`
}

// VerifyTaxpayerResponse represents taxpayer verification response
type VerifyTaxpayerResponse struct {
	OperationID          string    `json:"operationID"`
	TaxPayerName         string    `json:"taxPayerName"`
	TaxPayerTIN          string    `json:"taxPayerTIN"`
	VATNumber            string    `json:"vatNumber,omitempty"`
	DeviceBranchName     string    `json:"deviceBranchName"`
	DeviceBranchAddress  Address   `json:"deviceBranchAddress"`
	DeviceBranchContacts *Contacts `json:"deviceBranchContacts,omitempty"`
}

// IssueCertificateRequest represents certificate issuance request
type IssueCertificateRequest struct {
	DeviceID           int    `json:"deviceID" binding:"required"`
	CertificateRequest string `json:"certificateRequest" binding:"required"`
}

// IssueCertificateResponse represents certificate issuance response
type IssueCertificateResponse struct {
	OperationID string `json:"operationID"`
	Certificate string `json:"certificate"`
}

// GetConfigRequest represents configuration request
type GetConfigRequest struct {
	DeviceID int `json:"deviceID" binding:"required"`
}

// GetConfigResponse represents configuration response
type GetConfigResponse struct {
	OperationID                       string              `json:"operationID"`
	TaxPayerName                      string              `json:"taxPayerName"`
	TaxPayerTIN                       string              `json:"taxPayerTIN"`
	VATNumber                         string              `json:"vatNumber,omitempty"`
	DeviceSerialNo                    string              `json:"deviceSerialNo"`
	DeviceBranchName                  string              `json:"deviceBranchName"`
	DeviceBranchAddress               Address             `json:"deviceBranchAddress"`
	DeviceBranchContacts              *Contacts           `json:"deviceBranchContacts,omitempty"`
	DeviceOperatingMode               DeviceOperatingMode `json:"deviceOperatingMode"`
	TaxPayerDayMaxHrs                 int                 `json:"taxPayerDayMaxHrs"`
	TaxpayerDayEndNotificationHrs     int                 `json:"taxpayerDayEndNotificationHrs"`
	ApplicableTaxes                   []Tax               `json:"applicableTaxes"`
	CertificateValidTill              time.Time           `json:"certificateValidTill"`
	QrURL                             string              `json:"qrUrl"`
}

// Tax represents tax information
type Tax struct {
	TaxID        int       `json:"taxID"`
	TaxPercent   *float64  `json:"taxPercent,omitempty"` // nil for exempt
	TaxName      string    `json:"taxName"`
	TaxValidFrom time.Time `json:"taxValidFrom"`
	TaxValidTill *time.Time `json:"taxValidTill,omitempty"`
}

// PingRequest represents ping request
type PingRequest struct {
	DeviceID int `json:"deviceID" binding:"required"`
}

// PingResponse represents ping response
type PingResponse struct {
	OperationID         string `json:"operationID"`
	ReportingFrequency  int    `json:"reportingFrequency"` // in minutes
}

// GetStockListRequest represents stock list request
type GetStockListRequest struct {
	DeviceID int     `json:"deviceID" binding:"required"`
	HSCode   *string `json:"hsCode,omitempty"`
	GoodName *string `json:"goodName,omitempty"`
	Sort     *string `json:"sort,omitempty"`
	Order    *string `json:"order,omitempty"`
	Offset   int     `json:"offset" binding:"required,min=0"`
	Limit    int     `json:"limit" binding:"required,min=1,max=100"`
	Operator *string `json:"operator,omitempty"`
}

// GetStockListResponse represents stock list response
type GetStockListResponse struct {
	Total int     `json:"total"`
	Rows  []Good  `json:"rows"`
}

// Good represents a stock item
type Good struct {
	HSCode        string  `json:"hsCode"`
	GoodName      string  `json:"goodName"`
	Quantity      float64 `json:"quantity"`
	TaxPayerID    int64   `json:"taxPayerId"`
	TaxPayerName  string  `json:"taxPayerName"`
	BranchID      *int64  `json:"branchId,omitempty"`
	BranchName    *string `json:"branchName,omitempty"`
}

// GetServerCertificateRequest represents server certificate request
type GetServerCertificateRequest struct {
	Thumbprint []byte `json:"thumbprint,omitempty"`
}

// GetServerCertificateResponse represents server certificate response
type GetServerCertificateResponse struct {
	Certificate         []string  `json:"certificate"`
	CertificateValidTill time.Time `json:"certificateValidTill"`
}

// GetStatusResponse represents device status response
type GetStatusResponse struct {
	OperationID                 string                      `json:"operationID"`
	FiscalDayStatus             string                      `json:"fiscalDayStatus"` // Opened, Closed, CloseInitiated, CloseFailed
	FiscalDayNo                 *int                        `json:"fiscalDayNo,omitempty"`
	FiscalDayReconciliationMode *string                     `json:"fiscalDayReconciliationMode,omitempty"`
	FiscalDayServerSignature    *SignatureDataEx            `json:"fiscalDayServerSignature,omitempty"`
	FiscalDayClosed             *time.Time                  `json:"fiscalDayClosed,omitempty"`
	LastReceiptGlobalNo         *int                        `json:"lastReceiptGlobalNo,omitempty"`
	FiscalDayCounters           []FiscalDayCounter          `json:"fiscalDayCounters,omitempty"`
	FiscalDayDocumentQuantities []FiscalDayDocumentQuantity `json:"fiscalDayDocumentQuantities,omitempty"`
}

// FiscalDayCounter is imported from fiscal_day package
type FiscalDayCounter struct {
	FiscalCounterType       int      `json:"fiscalCounterType"`
	FiscalCounterCurrency   string   `json:"fiscalCounterCurrency"`
	FiscalCounterTaxID      *int     `json:"fiscalCounterTaxID,omitempty"`
	FiscalCounterTaxPercent *float64 `json:"fiscalCounterTaxPercent,omitempty"`
	FiscalCounterMoneyType  *int     `json:"fiscalCounterMoneyType,omitempty"`
	FiscalCounterValue      float64  `json:"fiscalCounterValue"`
}

// FiscalDayDocumentQuantity is imported from fiscal_day package
type FiscalDayDocumentQuantity struct {
	ReceiptType        int     `json:"receiptType"`
	ReceiptCurrency    string  `json:"receiptCurrency"`
	ReceiptQuantity    int     `json:"receiptQuantity"`
	ReceiptTotalAmount float64 `json:"receiptTotalAmount"`
}
