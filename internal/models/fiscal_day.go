package models

import (
	"time"
)

// FiscalDay represents a fiscal day
type FiscalDay struct {
	ID                       int64                       `json:"id" db:"id"`
	DeviceID                 int                         `json:"deviceID" db:"device_id"`
	FiscalDayNo              int                         `json:"fiscalDayNo" db:"fiscal_day_no"`
	FiscalDayOpened          time.Time                   `json:"fiscalDayOpened" db:"fiscal_day_opened"`
	FiscalDayClosed          *time.Time                  `json:"fiscalDayClosed,omitempty" db:"fiscal_day_closed"`
	Status                   FiscalDayStatus             `json:"status" db:"status"`
	ReconciliationMode       *FiscalDayReconciliationMode `json:"reconciliationMode,omitempty" db:"reconciliation_mode"`
	FiscalDayDeviceSignature *SignatureData              `json:"fiscalDayDeviceSignature,omitempty" db:"fiscal_day_device_signature"`
	FiscalDayServerSignature *SignatureDataEx            `json:"fiscalDayServerSignature,omitempty" db:"fiscal_day_server_signature"`
	ClosingErrorCode         *FiscalDayProcessingError   `json:"closingErrorCode,omitempty" db:"closing_error_code"`
	LastReceiptGlobalNo      *int                        `json:"lastReceiptGlobalNo,omitempty" db:"last_receipt_global_no"`
	CreatedAt                time.Time                   `json:"-" db:"created_at"`
	UpdatedAt                time.Time                   `json:"-" db:"updated_at"`
}
/*
// FiscalDayCounter represents a fiscal counter
type FiscalDayCounter struct {
	ID                      int64             `json:"-" db:"id"`
	FiscalDayID             int64             `json:"-" db:"fiscal_day_id"`
	FiscalCounterType       FiscalCounterType `json:"fiscalCounterType" db:"fiscal_counter_type"`
	FiscalCounterCurrency   string            `json:"fiscalCounterCurrency" db:"fiscal_counter_currency"`
	FiscalCounterTaxID      *int              `json:"fiscalCounterTaxID,omitempty" db:"fiscal_counter_tax_id"`
	FiscalCounterTaxPercent *float64          `json:"fiscalCounterTaxPercent,omitempty" db:"fiscal_counter_tax_percent"`
	FiscalCounterMoneyType  *MoneyType        `json:"fiscalCounterMoneyType,omitempty" db:"fiscal_counter_money_type"`
	FiscalCounterValue      float64           `json:"fiscalCounterValue" db:"fiscal_counter_value"`
}
// FiscalDayDocumentQuantity represents document quantities for a fiscal day
type FiscalDayDocumentQuantity struct {
	ReceiptType        ReceiptType `json:"receiptType"`
	ReceiptCurrency    string      `json:"receiptCurrency"`
	ReceiptQuantity    int         `json:"receiptQuantity"`
	ReceiptTotalAmount float64     `json:"receiptTotalAmount"`
}


// GetStatusResponse represents status response
type GetStatusResponse struct {
	OperationID                 string                       `json:"operationID"`
	FiscalDayStatus             FiscalDayStatus              `json:"fiscalDayStatus"`
	FiscalDayReconciliationMode *FiscalDayReconciliationMode `json:"fiscalDayReconciliationMode,omitempty"`
	FiscalDayServerSignature    *SignatureDataEx             `json:"fiscalDayServerSignature,omitempty"`
	FiscalDayClosed             *time.Time                   `json:"fiscalDayClosed,omitempty"`
	FiscalDayClosingErrorCode   *FiscalDayProcessingError    `json:"fiscalDayClosingErrorCode,omitempty"`
	FiscalDayCounters           []FiscalDayCounter           `json:"fiscalDayCounters,omitempty"`
	FiscalDayDocumentQuantities []FiscalDayDocumentQuantity  `json:"fiscalDayDocumentQuantities,omitempty"`
	LastReceiptGlobalNo         *int                         `json:"lastReceiptGlobalNo,omitempty"`
	LastFiscalDayNo             *int                         `json:"lastFiscalDayNo,omitempty"`
}


*/
// OpenDayRequest represents fiscal day opening request
type OpenDayRequest struct {
	DeviceID        int       `json:"deviceID" binding:"required"`
	FiscalDayOpened time.Time `json:"fiscalDayOpened" binding:"required"`
	FiscalDayNo     *int      `json:"fiscalDayNo,omitempty"`
}

// OpenDayResponse represents fiscal day opening response
type OpenDayResponse struct {
	OperationID string `json:"operationID"`
	FiscalDayNo int    `json:"fiscalDayNo"`
}

// CloseDayRequest represents fiscal day closing request
type CloseDayRequest struct {
	DeviceID                 int                `json:"deviceID" binding:"required"`
	FiscalDayNo              int                `json:"fiscalDayNo" binding:"required"`
	FiscalCounters           []FiscalDayCounter `json:"fiscalDayCounters" binding:"required"`
	FiscalDayDeviceSignature SignatureData      `json:"fiscalDayDeviceSignature" binding:"required"`
	ReceiptCounter           int                `json:"receiptCounter" binding:"required"`
}

// CloseDayResponse represents fiscal day closing response
type CloseDayResponse struct {
	OperationID string `json:"operationID"`
}

// GetStatusRequest represents status request
type GetStatusRequest struct {
	DeviceID int `json:"deviceID" binding:"required"`
}

// OpenFiscalDayRequest represents opening fiscal day request (simplified)
type OpenFiscalDayRequest struct {
	DeviceID int `json:"deviceID" binding:"required"`
}

// OpenFiscalDayResponse represents opening fiscal day response (simplified)
type OpenFiscalDayResponse struct {
	OperationID string `json:"operationID"`
	FiscalDayNo int    `json:"fiscalDayNo"`
}

// CloseFiscalDayRequest represents closing fiscal day request (simplified)
type CloseFiscalDayRequest struct {
	DeviceID                 int                 `json:"deviceID" binding:"required"`
	FiscalDayDeviceSignature *SignatureData      `json:"fiscalDayDeviceSignature,omitempty"`
	FiscalDayCounters        []FiscalDayCounter  `json:"fiscalDayCounters,omitempty"`
}

// CloseFiscalDayResponse represents closing fiscal day response (simplified)
type CloseFiscalDayResponse struct {
	OperationID                 string                      `json:"operationID"`
	FiscalDayServerSignature    SignatureDataEx             `json:"fiscalDayServerSignature"`
	FiscalDayCounters           []FiscalDayCounter          `json:"fiscalDayCounters"`
	FiscalDayDocumentQuantities []FiscalDayDocumentQuantity `json:"fiscalDayDocumentQuantities"`
}

// GetFiscalDayStatusResponse is an alias for GetStatusResponse
type GetFiscalDayStatusResponse = GetStatusResponse
