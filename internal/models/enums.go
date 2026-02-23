package models

// DeviceOperatingMode specifies allowed receipt processing modes
type DeviceOperatingMode int

const (
	DeviceOperatingModeOnline DeviceOperatingMode = iota
	DeviceOperatingModeOffline
)

func (m DeviceOperatingMode) String() string {
	return [...]string{"Online", "Offline"}[m]
}

// FiscalDayStatus represents the status of a fiscal day
type FiscalDayStatus int

const (
	FiscalDayStatusClosed FiscalDayStatus = iota
	FiscalDayStatusOpened
	FiscalDayStatusCloseInitiated
	FiscalDayStatusCloseFailed
)

func (s FiscalDayStatus) String() string {
	return [...]string{"FiscalDayClosed", "FiscalDayOpened", "FiscalDayCloseInitiated", "FiscalDayCloseFailed"}[s]
}

// FiscalDayReconciliationMode defines how fiscal day was closed
type FiscalDayReconciliationMode int

const (
	FiscalDayReconciliationModeAuto FiscalDayReconciliationMode = iota
	FiscalDayReconciliationModeManual
)

func (m FiscalDayReconciliationMode) String() string {
	return [...]string{"Auto", "Manual"}[m]
}

// FiscalCounterType represents type of fiscal counter
type FiscalCounterType int

const (
	FiscalCounterTypeSaleByTax FiscalCounterType = iota
	FiscalCounterTypeSaleTaxByTax
	FiscalCounterTypeCreditNoteByTax
	FiscalCounterTypeCreditNoteTaxByTax
	FiscalCounterTypeDebitNoteByTax
	FiscalCounterTypeDebitNoteTaxByTax
	FiscalCounterTypeBalanceByMoneyType
)

func (t FiscalCounterType) String() string {
	return [...]string{
		"SaleByTax",
		"SaleTaxByTax",
		"CreditNoteByTax",
		"CreditNoteTaxByTax",
		"DebitNoteByTax",
		"DebitNoteTaxByTax",
		"BalanceByMoneyType",
	}[t]
}

// MoneyType represents payment method
type MoneyType int

const (
	MoneyTypeCash MoneyType = iota
	MoneyTypeCard
	MoneyTypeMobileWallet
	MoneyTypeCoupon
	MoneyTypeCredit
	MoneyTypeBankTransfer
	MoneyTypeOther
)

func (t MoneyType) String() string {
	return [...]string{"Cash", "Card", "MobileWallet", "Coupon", "Credit", "BankTransfer", "Other"}[t]
}

// ReceiptType represents type of receipt
type ReceiptType int

const (
	ReceiptTypeFiscalInvoice ReceiptType = iota
	ReceiptTypeCreditNote
	ReceiptTypeDebitNote
)

func (t ReceiptType) String() string {
	return [...]string{"FiscalInvoice", "CreditNote", "DebitNote"}[t]
}

// ReceiptLineType represents type of receipt line
type ReceiptLineType int

const (
	ReceiptLineTypeSale ReceiptLineType = iota
	ReceiptLineTypeDiscount
)

func (t ReceiptLineType) String() string {
	return [...]string{"Sale", "Discount"}[t]
}

// ReceiptPrintForm represents the format of printed invoice
type ReceiptPrintForm int

const (
	ReceiptPrintFormReceipt48 ReceiptPrintForm = iota
	ReceiptPrintFormInvoiceA4
)

func (f ReceiptPrintForm) String() string {
	return [...]string{"Receipt48", "InvoiceA4"}[f]
}

// FiscalDayProcessingError represents errors during fiscal day closure
type FiscalDayProcessingError int

const (
	FiscalDayProcessingErrorBadCertificateSignature FiscalDayProcessingError = iota
	FiscalDayProcessingErrorMissingReceipts
	FiscalDayProcessingErrorReceiptsWithValidationErrors
	FiscalDayProcessingErrorCountersMismatch
)

func (e FiscalDayProcessingError) String() string {
	return [...]string{
		"BadCertificateSignature",
		"MissingReceipts",
		"ReceiptsWithValidationErrors",
		"CountersMismatch",
	}[e]
}

// FileProcessingStatus represents file processing status
type FileProcessingStatus int

const (
	FileProcessingStatusInProgress FileProcessingStatus = iota
	FileProcessingStatusSuccessful
	FileProcessingStatusWithErrors
	FileProcessingStatusWaitingForPreviousFile
)

func (s FileProcessingStatus) String() string {
	return [...]string{
		"FileProcessingInProgress",
		"FileProcessingIsSuccessful",
		"FileProcessingWithErrors",
		"WaitingForPreviousFile",
	}[s]
}

// FileProcessingError represents errors during file processing
type FileProcessingError int

const (
	FileProcessingErrorIncorrectFileFormat FileProcessingError = iota
	FileProcessingErrorFileSentForClosedDay
	FileProcessingErrorBadCertificateSignature
	FileProcessingErrorMissingReceipts
	FileProcessingErrorReceiptsWithValidationErrors
	FileProcessingErrorCountersMismatch
	FileProcessingErrorFileExceededAllowedWaitingTime
)

func (e FileProcessingError) String() string {
	return [...]string{
		"IncorrectFileFormat",
		"FileSentForClosedDay",
		"BadCertificateSignature",
		"MissingReceipts",
		"ReceiptsWithValidationErrors",
		"CountersMismatch",
		"FileExceededAllowedWaitingTime",
	}[e]
}

// UserStatus represents user status
type UserStatus int

const (
	UserStatusActive UserStatus = iota
	UserStatusBlocked
	UserStatusNotConfirmed
)

func (s UserStatus) String() string {
	return [...]string{"Active", "Blocked", "NotConfirmed"}[s]
}

// SendSecurityCodeTo represents channels for security code
type SendSecurityCodeTo int

const (
	SendSecurityCodeToEmail SendSecurityCodeTo = iota
	SendSecurityCodeToPhoneNumber
)

func (t SendSecurityCodeTo) String() string {
	return [...]string{"Email", "PhoneNumber"}[t]
}

// ValidationColor represents validation error severity
type ValidationColor string

const (
	ValidationColorGrey   ValidationColor = "Grey"
	ValidationColorYellow ValidationColor = "Yellow"
	ValidationColorRed    ValidationColor = "Red"
)
