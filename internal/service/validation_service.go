package service

import (
	"fmt"
	"strings"
	"time"

	"fiscalization-api/internal/models"
)

type ValidationService struct {
	// Can add configuration if needed
}

func NewValidationService() *ValidationService {
	return &ValidationService{}
}

// ValidationResult holds validation errors and color
type ValidationResult struct {
	IsValid bool
	Color   *models.ValidationColor
	Errors  []string
}

// ValidateReceipt performs comprehensive validation on a receipt
func (s *ValidationService) ValidateReceipt(
	receipt *models.Receipt,
	previousReceipt *models.Receipt,
	taxpayer *models.Taxpayer,
	applicableTaxes []models.Tax,
	fiscalDayOpened time.Time,
	fiscalDayMaxHrs int,
) ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Errors:  make([]string, 0),
	}

	// Currency validation (RCPT010)
	if !s.isValidCurrency(receipt.ReceiptCurrency) {
		result.addError("RCPT010", "Wrong currency code is used", models.ValidationColorRed)
	}

	// Receipt counter validation (RCPT011) - requires previous receipt
	if previousReceipt == nil {
		// Missing previous receipt - mark as grey
		if receipt.ReceiptCounter != 1 {
			result.addError("RCPT011", "Receipt counter is not sequential", models.ValidationColorGrey)
		}
	} else {
		if receipt.ReceiptCounter != previousReceipt.ReceiptCounter+1 {
			result.addError("RCPT011", "Receipt counter is not sequential", models.ValidationColorRed)
		}
	}

	// Receipt global number validation (RCPT012) - requires previous receipt
	if previousReceipt == nil {
		if receipt.ReceiptGlobalNo != 1 && receipt.ReceiptCounter != 1 {
			result.addError("RCPT012", "Receipt global number is not sequential", models.ValidationColorGrey)
		}
	} else {
		if receipt.ReceiptGlobalNo != previousReceipt.ReceiptGlobalNo+1 && receipt.ReceiptGlobalNo != 1 {
			result.addError("RCPT012", "Receipt global number is not sequential", models.ValidationColorRed)
		}
	}

	// Invoice number uniqueness (RCPT013) - checked at repository level

	// Receipt date validation (RCPT014)
	if receipt.ReceiptDate.Before(fiscalDayOpened) {
		result.addError("RCPT014", "Receipt date is earlier than fiscal day opening date", models.ValidationColorYellow)
	}

	// Credit/Debit note validation (RCPT015)
	if receipt.ReceiptType == models.ReceiptTypeCreditNote || receipt.ReceiptType == models.ReceiptTypeDebitNote {
		if receipt.CreditDebitNote == nil {
			result.addError("RCPT015", "Credited/debited invoice data is not provided", models.ValidationColorRed)
		}
	}

	// Receipt lines validation (RCPT016)
	if len(receipt.ReceiptLines) == 0 {
		result.addError("RCPT016", "No receipt lines provided", models.ValidationColorRed)
	}

	// Taxes validation (RCPT017)
	if len(receipt.ReceiptTaxes) == 0 {
		result.addError("RCPT017", "Taxes information is not provided", models.ValidationColorRed)
	}

	// Payment validation (RCPT018)
	if len(receipt.ReceiptPayments) == 0 {
		result.addError("RCPT018", "Payment information is not provided", models.ValidationColorRed)
	}

	// Receipt total validation (RCPT019, RCPT037, RCPT038, RCPT039, RCPT040)
	s.validateReceiptTotals(receipt, &result)

	// Signature validation (RCPT020) - done separately

	// VAT taxpayer validation (RCPT021)
	if taxpayer.VATNumber == nil {
		for _, tax := range receipt.ReceiptTaxes {
			if tax.TaxPercent != nil && *tax.TaxPercent > 0 {
				result.addError("RCPT021", "VAT tax is used while taxpayer is not VAT taxpayer", models.ValidationColorRed)
			}
		}
	}

	// Line price validation (RCPT022)
	for _, line := range receipt.ReceiptLines {
		if line.ReceiptLinePrice != nil {
			if receipt.ReceiptType == models.ReceiptTypeFiscalInvoice || receipt.ReceiptType == models.ReceiptTypeDebitNote {
				if line.ReceiptLineType == models.ReceiptLineTypeSale && *line.ReceiptLinePrice <= 0 {
					result.addError("RCPT022", "Invoice sales line price must be greater than 0", models.ValidationColorRed)
					break
				}
				if line.ReceiptLineType == models.ReceiptLineTypeDiscount && *line.ReceiptLinePrice >= 0 {
					result.addError("RCPT022", "Discount line price must be less than 0", models.ValidationColorRed)
					break
				}
			} else if receipt.ReceiptType == models.ReceiptTypeCreditNote {
				if line.ReceiptLineType == models.ReceiptLineTypeSale && *line.ReceiptLinePrice >= 0 {
					result.addError("RCPT022", "Credit note sales line price must be less than 0", models.ValidationColorRed)
					break
				}
			}
		}
	}

	// Line quantity validation (RCPT023)
	for _, line := range receipt.ReceiptLines {
		if line.ReceiptLineQuantity <= 0 {
			result.addError("RCPT023", "Invoice line quantity must be greater than 0", models.ValidationColorRed)
			break
		}
	}

	// Line total validation (RCPT024)
	for _, line := range receipt.ReceiptLines {
		if line.ReceiptLinePrice != nil {
			expectedTotal := *line.ReceiptLinePrice * line.ReceiptLineQuantity
			if !floatEquals(line.ReceiptLineTotal, expectedTotal, 0.01) {
				result.addError("RCPT024", "Invoice line total is not equal to unit price * quantity", models.ValidationColorRed)
				break
			}
		}
	}

	// Tax validation (RCPT025)
	s.validateTaxes(receipt, applicableTaxes, &result)

	// Tax amount validation (RCPT026)
	s.validateTaxAmounts(receipt, &result)

	// Sales amount validation (RCPT027)
	s.validateSalesAmounts(receipt, &result)

	// Payment amount validation (RCPT028)
	for _, payment := range receipt.ReceiptPayments {
		if receipt.ReceiptType == models.ReceiptTypeFiscalInvoice || receipt.ReceiptType == models.ReceiptTypeDebitNote {
			if payment.PaymentAmount < 0 {
				result.addError("RCPT028", "Payment amount must be >= 0 for invoices", models.ValidationColorRed)
				break
			}
		} else if receipt.ReceiptType == models.ReceiptTypeCreditNote {
			if payment.PaymentAmount > 0 {
				result.addError("RCPT028", "Payment amount must be <= 0 for credit notes", models.ValidationColorRed)
				break
			}
		}
	}

	// Receipt date vs previous (RCPT030) - requires previous receipt
	if previousReceipt != nil {
		if receipt.ReceiptDate.Before(previousReceipt.ReceiptDate) {
			result.addError("RCPT030", "Invoice date is earlier than previously submitted receipt date", models.ValidationColorRed)
		}
	}

	// Future date validation (RCPT031)
	if receipt.ReceiptDate.After(time.Now().Add(5 * time.Minute)) {
		result.addError("RCPT031", "Invoice is submitted with the future date", models.ValidationColorYellow)
	}

	// Fiscal day end validation (RCPT041)
	fiscalDayEnd := fiscalDayOpened.Add(time.Duration(fiscalDayMaxHrs) * time.Hour)
	if receipt.ReceiptDate.After(fiscalDayEnd) {
		result.addError("RCPT041", "Invoice is issued after fiscal day end", models.ValidationColorYellow)
	}

	// Buyer data validation (RCPT043)
	if receipt.BuyerData != nil {
		if receipt.BuyerData.BuyerRegisterName == "" || receipt.BuyerData.BuyerTIN == "" {
			result.addError("RCPT043", "Mandatory buyer data fields are not provided", models.ValidationColorRed)
		}
	}

	// HS Code validation for VAT payers (RCPT047, RCPT048)
	if taxpayer.VATNumber != nil {
		for _, line := range receipt.ReceiptLines {
			if line.ReceiptLineHSCode == nil || *line.ReceiptLineHSCode == "" {
				result.addError("RCPT047", "HS code must be sent if taxpayer is a VAT payer", models.ValidationColorRed)
				break
			}
			
			hsCodeLen := len(*line.ReceiptLineHSCode)
			if line.TaxPercent != nil && *line.TaxPercent > 0 {
				if hsCodeLen != 4 && hsCodeLen != 8 {
					result.addError("RCPT048", "HS code length must be 4 or 8 digits for VAT items", models.ValidationColorRed)
					break
				}
			} else {
				if hsCodeLen != 8 {
					result.addError("RCPT048", "HS code length must be 8 digits for exempt/zero-rated items", models.ValidationColorRed)
					break
				}
			}
		}
	}

	return result
}

// ValidateCreditDebitNote validates credit/debit note specific rules
func (s *ValidationService) ValidateCreditDebitNote(
	note *models.Receipt,
	originalReceipt *models.Receipt,
	allCreditNotes []*models.Receipt,
	allDebitNotes []*models.Receipt,
) ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Errors:  make([]string, 0),
	}

	if originalReceipt == nil {
		result.addError("RCPT032", "Credit/debit note refers to non-existing invoice", models.ValidationColorRed)
		return result
	}

	// Date validation (RCPT033)
	twelveMonthsAgo := note.ReceiptDate.AddDate(0, -12, 0)
	if originalReceipt.ReceiptDate.Before(twelveMonthsAgo) {
		result.addError("RCPT033", "Credited/debited invoice is issued more than 12 months ago", models.ValidationColorRed)
	}

	// Notes mandatory (RCPT034)
	if note.ReceiptNotes == nil || *note.ReceiptNotes == "" {
		result.addError("RCPT034", "Note for credit/debit note is not provided", models.ValidationColorRed)
	}

	// Amount validation (RCPT035)
	totalCreditAmount := 0.0
	for _, cn := range allCreditNotes {
		totalCreditAmount += cn.ReceiptTotal
	}
	totalDebitAmount := 0.0
	for _, dn := range allDebitNotes {
		totalDebitAmount += dn.ReceiptTotal
	}
	
	remainingAmount := originalReceipt.ReceiptTotal - totalCreditAmount + totalDebitAmount
	if note.ReceiptType == models.ReceiptTypeCreditNote {
		if remainingAmount+note.ReceiptTotal < 0 {
			result.addError("RCPT035", "Total credit note amount exceeds original invoice amount", models.ValidationColorRed)
		}
	}

	// Tax validation (RCPT036)
	originalTaxIDs := make(map[int]bool)
	for _, tax := range originalReceipt.ReceiptTaxes {
		originalTaxIDs[tax.TaxID] = true
	}
	
	for _, tax := range note.ReceiptTaxes {
		if !originalTaxIDs[tax.TaxID] {
			result.addError("RCPT036", "Credit/debit note uses other taxes than are used in the original invoice", models.ValidationColorRed)
			break
		}
	}

	// Currency validation (RCPT042)
	if note.ReceiptCurrency != originalReceipt.ReceiptCurrency {
		result.addError("RCPT042", "Credit/debit note uses other currency than is used in the original invoice", models.ValidationColorRed)
	}

	// FiscalInvoice with credit/debit note data (RCPT029)
	if note.ReceiptType == models.ReceiptTypeFiscalInvoice && note.CreditDebitNote != nil {
		result.addError("RCPT029", "Credited/debited invoice information provided for regular invoice", models.ValidationColorRed)
	}

	return result
}

// Helper methods

func (r *ValidationResult) addError(code, message string, color models.ValidationColor) {
	r.IsValid = false
	r.Errors = append(r.Errors, fmt.Sprintf("%s: %s", code, message))
	
	// Set the most severe color
	if r.Color == nil {
		r.Color = &color
	} else {
		if color == models.ValidationColorRed {
			r.Color = &color
		} else if color == models.ValidationColorYellow && *r.Color == models.ValidationColorGrey {
			r.Color = &color
		}
	}
}

func (s *ValidationService) validateReceiptTotals(receipt *models.Receipt, result *ValidationResult) {
	var linesTotalSum float64
	for _, line := range receipt.ReceiptLines {
		linesTotalSum += line.ReceiptLineTotal
	}

	var taxesSum float64
	var salesWithTaxSum float64
	for _, tax := range receipt.ReceiptTaxes {
		taxesSum += tax.TaxAmount
		salesWithTaxSum += tax.SalesAmountWithTax
	}

	var paymentsSum float64
	for _, payment := range receipt.ReceiptPayments {
		paymentsSum += payment.PaymentAmount
	}

	// RCPT019 / RCPT037
	if receipt.ReceiptLinesTaxInclusive {
		if !floatEquals(receipt.ReceiptTotal, linesTotalSum, 0.01) {
			result.addError("RCPT019", "Invoice total amount is not equal to sum of all invoice lines", models.ValidationColorRed)
		}
	} else {
		if !floatEquals(receipt.ReceiptTotal, linesTotalSum+taxesSum, 0.01) {
			result.addError("RCPT037", "Invoice total amount is not equal to sum of all invoice lines and taxes", models.ValidationColorRed)
		}
	}

	// RCPT038
	if !floatEquals(receipt.ReceiptTotal, salesWithTaxSum, 0.01) {
		result.addError("RCPT038", "Invoice total amount is not equal to sum of sales amount including tax", models.ValidationColorRed)
	}

	// RCPT039
	if !floatEquals(receipt.ReceiptTotal, paymentsSum, 0.01) {
		result.addError("RCPT039", "Invoice total amount is not equal to sum of all payment amounts", models.ValidationColorRed)
	}

	// RCPT040
	if receipt.ReceiptType == models.ReceiptTypeFiscalInvoice || receipt.ReceiptType == models.ReceiptTypeDebitNote {
		if receipt.ReceiptTotal < 0 {
			result.addError("RCPT040", "Invoice total amount must be >= 0 for invoices", models.ValidationColorRed)
		}
	} else if receipt.ReceiptType == models.ReceiptTypeCreditNote {
		if receipt.ReceiptTotal > 0 {
			result.addError("RCPT040", "Invoice total amount must be <= 0 for credit notes", models.ValidationColorRed)
		}
	}
}

func (s *ValidationService) validateTaxes(receipt *models.Receipt, applicableTaxes []models.Tax, result *ValidationResult) {
	validTaxMap := make(map[string]models.Tax)
	for _, tax := range applicableTaxes {
		key := fmt.Sprintf("%d_%.2f", tax.TaxID, getTaxPercent(tax.TaxPercent))
		validTaxMap[key] = tax
	}

	for _, receiptTax := range receipt.ReceiptTaxes {
		key := fmt.Sprintf("%d_%.2f", receiptTax.TaxID, getTaxPercent(receiptTax.TaxPercent))
		if tax, exists := validTaxMap[key]; !exists {
			result.addError("RCPT025", "Invalid tax is used", models.ValidationColorRed)
		} else {
			// Check if receipt date is within tax validity period
			if receipt.ReceiptDate.Before(tax.TaxValidFrom) ||
				(tax.TaxValidTill != nil && receipt.ReceiptDate.After(*tax.TaxValidTill)) {
				result.addError("RCPT025", "Receipt date is not in tax valid period", models.ValidationColorRed)
			}
		}
	}
}

func (s *ValidationService) validateTaxAmounts(receipt *models.Receipt, result *ValidationResult) {
	// Group lines by tax
	lineTotalsByTax := make(map[string]float64)
	for _, line := range receipt.ReceiptLines {
		key := getTaxKey(line.TaxCode, line.TaxPercent)
		lineTotalsByTax[key] += line.ReceiptLineTotal
	}

	// Validate tax amounts
	for _, tax := range receipt.ReceiptTaxes {
		key := getTaxKey(tax.TaxCode, tax.TaxPercent)
		lineTotal := lineTotalsByTax[key]

		var expectedTaxAmount float64
		if receipt.ReceiptLinesTaxInclusive {
			if tax.TaxPercent != nil {
				expectedTaxAmount = lineTotal * (*tax.TaxPercent / 100.0) / (1 + (*tax.TaxPercent / 100.0))
			}
		} else {
			if tax.TaxPercent != nil {
				expectedTaxAmount = lineTotal * (*tax.TaxPercent / 100.0)
			}
		}

		if !floatEquals(tax.TaxAmount, expectedTaxAmount, 0.01) {
			result.addError("RCPT026", "Incorrectly calculated tax amount", models.ValidationColorRed)
			break
		}
	}
}

func (s *ValidationService) validateSalesAmounts(receipt *models.Receipt, result *ValidationResult) {
	lineTotalsByTax := make(map[string]float64)
	for _, line := range receipt.ReceiptLines {
		key := getTaxKey(line.TaxCode, line.TaxPercent)
		lineTotalsByTax[key] += line.ReceiptLineTotal
	}

	for _, tax := range receipt.ReceiptTaxes {
		key := getTaxKey(tax.TaxCode, tax.TaxPercent)
		lineTotal := lineTotalsByTax[key]

		var expectedSalesAmount float64
		if receipt.ReceiptLinesTaxInclusive {
			expectedSalesAmount = lineTotal
		} else {
			if tax.TaxPercent != nil {
				expectedSalesAmount = lineTotal * (1 + (*tax.TaxPercent / 100.0))
			} else {
				expectedSalesAmount = lineTotal
			}
		}

		if !floatEquals(tax.SalesAmountWithTax, expectedSalesAmount, 0.01) {
			result.addError("RCPT027", "Incorrectly calculated total sales amount", models.ValidationColorRed)
			break
		}
	}
}

func (s *ValidationService) isValidCurrency(currency string) bool {
	validCurrencies := []string{"USD", "ZWL", "EUR", "GBP", "ZAR"}
	for _, valid := range validCurrencies {
		if strings.ToUpper(currency) == valid {
			return true
		}
	}
	return false
}

func floatEquals(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

func getTaxPercent(percent *float64) float64 {
	if percent == nil {
		return 0
	}
	return *percent
}

func getTaxKey(code *string, percent *float64) string {
	taxCode := ""
	if code != nil {
		taxCode = *code
	}
	taxPercent := getTaxPercent(percent)
	return fmt.Sprintf("%s_%.2f", taxCode, taxPercent)
}
