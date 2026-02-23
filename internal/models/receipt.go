package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Receipt represents a fiscal receipt (invoice, credit note, or debit note)
type Receipt struct {
	ID                      int64            `json:"id" db:"id"`
	ReceiptID               int64            `json:"receiptID" db:"receipt_id"` // FDMS assigned ID
	DeviceID                int              `json:"deviceID" db:"device_id"`
	FiscalDayID             int64            `json:"-" db:"fiscal_day_id"`
	ReceiptType             ReceiptType      `json:"receiptType" db:"receipt_type"`
	ReceiptCurrency         string           `json:"receiptCurrency" db:"receipt_currency"`
	ReceiptCounter          int              `json:"receiptCounter" db:"receipt_counter"`
	ReceiptGlobalNo         int              `json:"receiptGlobalNo" db:"receipt_global_no"`
	InvoiceNo               string           `json:"invoiceNo" db:"invoice_no"`
	BuyerData               *Buyer           `json:"buyerData,omitempty" db:"buyer_data"`
	ReceiptNotes            *string          `json:"receiptNotes,omitempty" db:"receipt_notes"`
	ReceiptDate             time.Time        `json:"receiptDate" db:"receipt_date"`
	CreditDebitNote         *CreditDebitNote `json:"creditDebitNote,omitempty" db:"credit_debit_note"`
	ReceiptLinesTaxInclusive bool            `json:"receiptLinesTaxInclusive" db:"receipt_lines_tax_inclusive"`
	ReceiptLines            []ReceiptLine    `json:"receiptLines" db:"-"`
	ReceiptTaxes            []ReceiptTax     `json:"receiptTaxes" db:"-"`
	ReceiptPayments         []Payment        `json:"receiptPayments" db:"-"`
	ReceiptTotal            float64          `json:"receiptTotal" db:"receipt_total"`
	ReceiptPrintForm        ReceiptPrintForm `json:"receiptPrintForm,omitempty" db:"receipt_print_form"`
	ReceiptDeviceSignature  SignatureData    `json:"receiptDeviceSignature" db:"receipt_device_signature"`
	ReceiptServerSignature  *SignatureDataEx `json:"receiptServerSignature,omitempty" db:"receipt_server_signature"`
	ReceiptHash             []byte           `json:"-" db:"receipt_hash"`
	Username                *string          `json:"username,omitempty" db:"username"`
	UserNameSurname         *string          `json:"userNameSurname,omitempty" db:"user_name_surname"`
	ValidationColor         *ValidationColor `json:"-" db:"validation_color"`
	ValidationErrors        []string         `json:"-" db:"validation_errors"`
	ServerDate              *time.Time       `json:"serverDate,omitempty" db:"server_date"`
	CreatedAt               time.Time        `json:"-" db:"created_at"`
	UpdatedAt               time.Time        `json:"-" db:"updated_at"`
}

// Buyer represents buyer information
type Buyer struct {
	BuyerRegisterName string    `json:"buyerRegisterName"`
	BuyerTradeName    *string   `json:"buyerTradeName,omitempty"`
	BuyerTIN          string    `json:"buyerTIN"`
	VATNumber         *string   `json:"VATNumber,omitempty"`
	BuyerContacts     *Contacts `json:"buyerContacts,omitempty"`
	BuyerAddress      *Address  `json:"buyerAddress,omitempty"`
}

// Value implements driver.Valuer for Buyer
func (b Buyer) Value() (driver.Value, error) {
	return json.Marshal(b)
}

// Scan implements sql.Scanner for Buyer
func (b *Buyer) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Buyer")
	}
	return json.Unmarshal(bytes, b)
}

// CreditDebitNote represents credited/debited receipt information
type CreditDebitNote struct {
	ReceiptID       *int64 `json:"receiptID,omitempty"`
	DeviceID        *int   `json:"deviceID,omitempty"`
	ReceiptGlobalNo *int   `json:"receiptGlobalNo,omitempty"`
	FiscalDayNo     *int   `json:"fiscalDayNo,omitempty"`
}

// Value implements driver.Valuer for CreditDebitNote
func (c CreditDebitNote) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner for CreditDebitNote
func (c *CreditDebitNote) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan CreditDebitNote")
	}
	return json.Unmarshal(bytes, c)
}

// ReceiptLine represents a receipt line item
type ReceiptLine struct {
	ID                 int64           `json:"-" db:"id"`
	ReceiptID          int64           `json:"-" db:"receipt_id"`
	ReceiptLineType    ReceiptLineType `json:"receiptLineType" db:"receipt_line_type"`
	ReceiptLineNo      int             `json:"receiptLineNo" db:"receipt_line_no"`
	ReceiptLineHSCode  *string         `json:"receiptLineHSCode,omitempty" db:"receipt_line_hs_code"`
	ReceiptLineName    string          `json:"receiptLineName" db:"receipt_line_name"`
	ReceiptLinePrice   *float64        `json:"receiptLinePrice,omitempty" db:"receipt_line_price"`
	ReceiptLineQuantity float64        `json:"receiptLineQuantity" db:"receipt_line_quantity"`
	ReceiptLineTotal   float64         `json:"receiptLineTotal" db:"receipt_line_total"`
	TaxCode            *string         `json:"taxCode,omitempty" db:"tax_code"`
	TaxPercent         *float64        `json:"taxPercent,omitempty" db:"tax_percent"`
	TaxID              int             `json:"taxID" db:"tax_id"`
}

// ReceiptTax represents tax breakdown for a receipt
type ReceiptTax struct {
	ID                  int64    `json:"-" db:"id"`
	ReceiptID           int64    `json:"-" db:"receipt_id"`
	TaxCode             *string  `json:"taxCode,omitempty" db:"tax_code"`
	TaxPercent          *float64 `json:"taxPercent,omitempty" db:"tax_percent"`
	TaxID               int      `json:"taxID" db:"tax_id"`
	TaxAmount           float64  `json:"taxAmount" db:"tax_amount"`
	SalesAmountWithTax  float64  `json:"salesAmountWithTax" db:"sales_amount_with_tax"`
}

// Payment represents a payment method
type Payment struct {
	ID            int64     `json:"-" db:"id"`
	ReceiptID     int64     `json:"-" db:"receipt_id"`
	MoneyTypeCode MoneyType `json:"moneyTypeCode" db:"money_type_code"`
	PaymentAmount float64   `json:"paymentAmount" db:"payment_amount"`
}

// SubmitReceiptRequest represents receipt submission request
type SubmitReceiptRequest struct {
	DeviceID int     `json:"deviceID" binding:"required"`
	Receipt  Receipt `json:"receipt" binding:"required"`
}

// SubmitReceiptResponse represents receipt submission response
type SubmitReceiptResponse struct {
	OperationID            string           `json:"operationID"`
	ReceiptID              int64            `json:"receiptID"`
	ServerDate             time.Time        `json:"serverDate"`
	ReceiptServerSignature SignatureDataEx  `json:"receiptServerSignature"`
}

// SubmitFileRequest represents file submission request
type SubmitFileRequest struct {
	DeviceID int    `json:"deviceID" binding:"required"`
	File     []byte `json:"file" binding:"required"` // Base64 encoded
}

// SubmitFileResponse represents file submission response
type SubmitFileResponse struct {
	OperationID string `json:"operationID"`
}

// FileUpload represents uploaded file structure
type FileUpload struct {
	Header  FileHeader   `json:"header" binding:"required"`
	Content *FileContent `json:"content,omitempty"`
	Footer  *FileFooter  `json:"footer,omitempty"`
}

// FileHeader represents file header
type FileHeader struct {
	DeviceID        int       `json:"deviceID" binding:"required"`
	FiscalDayNo     int       `json:"fiscalDayNo" binding:"required"`
	FiscalDayOpened time.Time `json:"fiscalDayOpened" binding:"required"`
	FileSequence    int       `json:"fileSequence" binding:"required"`
}

// FileContent represents file content
type FileContent struct {
	Receipts []Receipt `json:"receipts" binding:"required"`
}

// FileFooter represents file footer
type FileFooter struct {
	FiscalCounters           []FiscalDayCounter `json:"fiscalCounters,omitempty"`
	FiscalDayDeviceSignature SignatureData      `json:"fiscalDayDeviceSignature" binding:"required"`
	ReceiptCounter           int                `json:"receiptCounter" binding:"required"`
	FiscalDayClosed          time.Time          `json:"fiscalDayClosed" binding:"required"`
}

// GetFileStatusRequest represents file status request
type GetFileStatusRequest struct {
	DeviceID         int     `json:"deviceID" binding:"required"`
	OperationID      *string `json:"operationID,omitempty"`
	FileUploadedFrom string  `json:"fileUploadedFrom" binding:"required"` // Date format
	FileUploadedTill string  `json:"fileUploadedTill" binding:"required"` // Date format
}

// GetFileStatusResponse represents file status response
type GetFileStatusResponse struct {
	OperationID string       `json:"operationID"`
	FileStatus  []FileStatus `json:"fileStatus"`
}

// FileStatus represents status of an uploaded file
type FileStatus struct {
	OperationID              string                 `json:"operationID"`
	FileUploadDate           time.Time              `json:"fileUploadDate"`
	DeviceID                 int                    `json:"deviceId"`
	FileName                 string                 `json:"fileName"`
	FileProcessingDate       *time.Time             `json:"fileProcessingDate,omitempty"`
	FileProcessingStatus     FileProcessingStatus   `json:"fileProcessingStatus"`
	FileProcessingErrorCode  []FileProcessingError  `json:"fileProcessingErrorCode,omitempty"`
	FiscalDayNo              int                    `json:"fiscalDayNo"`
	FiscalDayOpenedAt        time.Time              `json:"fiscalDayOpenedAt"`
	FileSequence             int                    `json:"fileSequence"`
	IPAddress                string                 `json:"ipAddress"`
}
