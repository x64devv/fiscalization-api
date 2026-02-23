package service

import (
	"testing"
	"time"

	"fiscalization-api/internal/models"
)

func TestValidationService_ValidateReceipt(t *testing.T) {
	svc := NewValidationService()

	tests := []struct {
		name           string
		receipt        *models.Receipt
		previousReceipt *models.Receipt
		taxpayer       *models.Taxpayer
		taxes          []models.Tax
		wantValid      bool
		wantColor      *models.ValidationColor
	}{
		{
			name: "Valid receipt",
			receipt: &models.Receipt{
				ReceiptType:     models.ReceiptTypeFiscalInvoice,
				ReceiptCurrency: "USD",
				ReceiptCounter:  1,
				ReceiptGlobalNo: 1,
				ReceiptDate:     time.Now(),
				ReceiptTotal:    100.00,
				ReceiptLines: []models.ReceiptLine{
					{
						ReceiptLineType:     models.ReceiptLineTypeSale,
						ReceiptLineNo:       1,
						ReceiptLineName:     "Test Product",
						ReceiptLineQuantity: 1,
						ReceiptLineTotal:    100.00,
						TaxID:               1,
					},
				},
				ReceiptTaxes: []models.ReceiptTax{
					{
						TaxID:              1,
						TaxPercent:         float64Ptr(15.0),
						TaxAmount:          13.04,
						SalesAmountWithTax: 100.00,
					},
				},
				ReceiptPayments: []models.Payment{
					{
						MoneyTypeCode: models.MoneyTypeCash,
						PaymentAmount: 100.00,
					},
				},
				ReceiptLinesTaxInclusive: true,
			},
			previousReceipt: nil,
			taxpayer: &models.Taxpayer{
				VATNumber: stringPtr("123456789"),
			},
			taxes: []models.Tax{
				{
					TaxID:      1,
					TaxPercent: float64Ptr(15.0),
					TaxValidFrom: time.Now().AddDate(0, 0, -30),
				},
			},
			wantValid: true,
			wantColor: nil,
		},
		{
			name: "Invalid currency",
			receipt: &models.Receipt{
				ReceiptType:     models.ReceiptTypeFiscalInvoice,
				ReceiptCurrency: "XXX",
				ReceiptCounter:  1,
				ReceiptGlobalNo: 1,
				ReceiptDate:     time.Now(),
				ReceiptTotal:    100.00,
				ReceiptLines: []models.ReceiptLine{
					{
						ReceiptLineType:     models.ReceiptLineTypeSale,
						ReceiptLineNo:       1,
						ReceiptLineName:     "Test Product",
						ReceiptLineQuantity: 1,
						ReceiptLineTotal:    100.00,
					},
				},
				ReceiptTaxes: []models.ReceiptTax{
					{
						TaxID:              1,
						TaxAmount:          15.00,
						SalesAmountWithTax: 100.00,
					},
				},
				ReceiptPayments: []models.Payment{
					{
						MoneyTypeCode: models.MoneyTypeCash,
						PaymentAmount: 100.00,
					},
				},
				ReceiptLinesTaxInclusive: true,
			},
			previousReceipt: nil,
			taxpayer: &models.Taxpayer{},
			taxes:    []models.Tax{},
			wantValid: false,
			wantColor: colorPtr(models.ValidationColorRed),
		},
		{
			name: "Missing receipt lines",
			receipt: &models.Receipt{
				ReceiptType:     models.ReceiptTypeFiscalInvoice,
				ReceiptCurrency: "USD",
				ReceiptCounter:  1,
				ReceiptGlobalNo: 1,
				ReceiptDate:     time.Now(),
				ReceiptTotal:    100.00,
				ReceiptLines:    []models.ReceiptLine{},
				ReceiptTaxes: []models.ReceiptTax{
					{
						TaxID:              1,
						TaxAmount:          15.00,
						SalesAmountWithTax: 100.00,
					},
				},
				ReceiptPayments: []models.Payment{
					{
						MoneyTypeCode: models.MoneyTypeCash,
						PaymentAmount: 100.00,
					},
				},
			},
			previousReceipt: nil,
			taxpayer: &models.Taxpayer{},
			taxes:    []models.Tax{},
			wantValid: false,
			wantColor: colorPtr(models.ValidationColorRed),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.ValidateReceipt(
				tt.receipt,
				tt.previousReceipt,
				tt.taxpayer,
				tt.taxes,
				time.Now(),
				24,
			)

			if result.IsValid != tt.wantValid {
				t.Errorf("ValidateReceipt() IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}

			if tt.wantColor != nil {
				if result.Color == nil {
					t.Errorf("ValidateReceipt() Color = nil, want %v", *tt.wantColor)
				} else if *result.Color != *tt.wantColor {
					t.Errorf("ValidateReceipt() Color = %v, want %v", *result.Color, *tt.wantColor)
				}
			}
		})
	}
}

func TestValidationService_ValidateCurrency(t *testing.T) {
	svc := NewValidationService()

	tests := []struct {
		currency string
		want     bool
	}{
		{"USD", true},
		{"ZWL", true},
		{"EUR", true},
		{"GBP", true},
		{"ZAR", true},
		{"XXX", false},
		{"usd", true}, // Should handle lowercase
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.currency, func(t *testing.T) {
			if got := svc.isValidCurrency(tt.currency); got != tt.want {
				t.Errorf("isValidCurrency(%q) = %v, want %v", tt.currency, got, tt.want)
			}
		})
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}

func colorPtr(c models.ValidationColor) *models.ValidationColor {
	return &c
}
