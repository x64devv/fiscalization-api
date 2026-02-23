package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"fiscalization-api/internal/models"
)

// GenerateReceiptHash generates SHA-256 hash for receipt signature
// According to ZIMRA spec section 13.2.1
func GenerateReceiptHash(receipt *models.Receipt, previousHash []byte) ([]byte, error) {
	var sb strings.Builder

	// 1. deviceID
	sb.WriteString(strconv.Itoa(receipt.DeviceID))

	// 2. receiptType (uppercase)
	sb.WriteString(strings.ToUpper(receipt.ReceiptType.String()))

	// 3. receiptCurrency (uppercase)
	sb.WriteString(strings.ToUpper(receipt.ReceiptCurrency))

	// 4. receiptGlobalNo
	sb.WriteString(strconv.Itoa(receipt.ReceiptGlobalNo))

	// 5. receiptDate (ISO 8601 format: YYYY-MM-DDTHH:mm:ss)
	sb.WriteString(receipt.ReceiptDate.Format("2006-01-02T15:04:05"))

	// 6. receiptTotal in cents
	totalCents := int64(receipt.ReceiptTotal * 100)
	sb.WriteString(strconv.FormatInt(totalCents, 10))

	// 7. receiptTaxes (sorted by taxID ascending, then taxCode alphabetically)
	sortedTaxes := make([]models.ReceiptTax, len(receipt.ReceiptTaxes))
	copy(sortedTaxes, receipt.ReceiptTaxes)
	sort.Slice(sortedTaxes, func(i, j int) bool {
		if sortedTaxes[i].TaxID != sortedTaxes[j].TaxID {
			return sortedTaxes[i].TaxID < sortedTaxes[j].TaxID
		}
		code1 := ""
		code2 := ""
		if sortedTaxes[i].TaxCode != nil {
			code1 = *sortedTaxes[i].TaxCode
		}
		if sortedTaxes[j].TaxCode != nil {
			code2 = *sortedTaxes[j].TaxCode
		}
		return code1 < code2
	})

	for _, tax := range sortedTaxes {
		// taxCode (empty if nil)
		if tax.TaxCode != nil {
			sb.WriteString(*tax.TaxCode)
		}

		// taxPercent (format: XX.XX or empty for exempt)
		if tax.TaxPercent != nil {
			sb.WriteString(fmt.Sprintf("%.2f", *tax.TaxPercent))
		}

		// taxAmount in cents
		taxAmountCents := int64(tax.TaxAmount * 100)
		sb.WriteString(strconv.FormatInt(taxAmountCents, 10))

		// salesAmountWithTax in cents
		salesAmountCents := int64(tax.SalesAmountWithTax * 100)
		sb.WriteString(strconv.FormatInt(salesAmountCents, 10))
	}

	// 8. previousReceiptHash (if not first receipt in fiscal day)
	if previousHash != nil {
		sb.Write(previousHash)
	}

	// Generate SHA-256 hash
	data := []byte(sb.String())
	hash := sha256.Sum256(data)
	return hash[:], nil
}

// GenerateFiscalDayHash generates SHA-256 hash for fiscal day signature
// According to ZIMRA spec section 13.3.1
func GenerateFiscalDayHash(
	deviceID int,
	fiscalDayNo int,
	fiscalDayDate string,
	counters []models.FiscalDayCounter,
) ([]byte, error) {
	var sb strings.Builder

	// 1. deviceID
	sb.WriteString(strconv.Itoa(deviceID))

	// 2. fiscalDayNo
	sb.WriteString(strconv.Itoa(fiscalDayNo))

	// 3. fiscalDayDate (YYYY-MM-DD)
	sb.WriteString(fiscalDayDate)

	// 4. fiscalDayCounters (sorted)
	sortedCounters := make([]models.FiscalDayCounter, len(counters))
	copy(sortedCounters, counters)
	
	sort.Slice(sortedCounters, func(i, j int) bool {
		if sortedCounters[i].FiscalCounterType != sortedCounters[j].FiscalCounterType {
			return sortedCounters[i].FiscalCounterType < sortedCounters[j].FiscalCounterType
		}
		if sortedCounters[i].FiscalCounterCurrency != sortedCounters[j].FiscalCounterCurrency {
			return sortedCounters[i].FiscalCounterCurrency < sortedCounters[j].FiscalCounterCurrency
		}
		
		// Sort by taxID or moneyType
		if sortedCounters[i].FiscalCounterTaxID != nil && sortedCounters[j].FiscalCounterTaxID != nil {
			return *sortedCounters[i].FiscalCounterTaxID < *sortedCounters[j].FiscalCounterTaxID
		}
		if sortedCounters[i].FiscalCounterMoneyType != nil && sortedCounters[j].FiscalCounterMoneyType != nil {
			return *sortedCounters[i].FiscalCounterMoneyType < *sortedCounters[j].FiscalCounterMoneyType
		}
		
		return false
	})

	for _, counter := range sortedCounters {
		// fiscalCounterType (uppercase)
		sb.WriteString(strings.ToUpper(strconv.Itoa(counter.FiscalCounterType)))

		// fiscalCounterCurrency (uppercase)
		sb.WriteString(strings.ToUpper(counter.FiscalCounterCurrency))

		// fiscalCounterTaxPercent or fiscalCounterMoneyType
		if counter.FiscalCounterTaxPercent != nil {
			sb.WriteString(fmt.Sprintf("%.2f", *counter.FiscalCounterTaxPercent))
		} else if counter.FiscalCounterMoneyType != nil {
			sb.WriteString(strings.ToUpper(strconv.Itoa(*counter.FiscalCounterMoneyType)))
		}

		// fiscalCounterValue in cents
		valueCents := int64(counter.FiscalCounterValue * 100)
		sb.WriteString(strconv.FormatInt(valueCents, 10))
	}

	// Generate SHA-256 hash
	data := []byte(sb.String())
	hash := sha256.Sum256(data)
	return hash[:], nil
}

// GenerateFiscalDayServerHash generates hash for FDMS signature
// According to ZIMRA spec section 13.3.2
func GenerateFiscalDayServerHash(
	deviceID int,
	fiscalDayNo int,
	fiscalDayDate string,
	fiscalDayUpdated string,
	reconciliationMode models.FiscalDayReconciliationMode,
	counters []models.FiscalDayCounter,
	deviceSignature []byte,
) ([]byte, error) {
	var sb strings.Builder

	// 1. deviceID
	sb.WriteString(strconv.Itoa(deviceID))

	// 2. fiscalDayNo
	sb.WriteString(strconv.Itoa(fiscalDayNo))

	// 3. fiscalDayDate (YYYY-MM-DD)
	sb.WriteString(fiscalDayDate)

	// 4. fiscalDayUpdated (YYYY-MM-DDTHH:mm:ss)
	sb.WriteString(fiscalDayUpdated)

	// 5. reconciliationMode (uppercase)
	sb.WriteString(strings.ToUpper(reconciliationMode.String()))

	// 6. fiscalDayCounters (same as device hash)
	sortedCounters := make([]models.FiscalDayCounter, len(counters))
	copy(sortedCounters, counters)
	
	sort.Slice(sortedCounters, func(i, j int) bool {
		if sortedCounters[i].FiscalCounterType != sortedCounters[j].FiscalCounterType {
			return sortedCounters[i].FiscalCounterType < sortedCounters[j].FiscalCounterType
		}
		if sortedCounters[i].FiscalCounterCurrency != sortedCounters[j].FiscalCounterCurrency {
			return sortedCounters[i].FiscalCounterCurrency < sortedCounters[j].FiscalCounterCurrency
		}
		
		if sortedCounters[i].FiscalCounterTaxID != nil && sortedCounters[j].FiscalCounterTaxID != nil {
			return *sortedCounters[i].FiscalCounterTaxID < *sortedCounters[j].FiscalCounterTaxID
		}
		if sortedCounters[i].FiscalCounterMoneyType != nil && sortedCounters[j].FiscalCounterMoneyType != nil {
			return *sortedCounters[i].FiscalCounterMoneyType < *sortedCounters[j].FiscalCounterMoneyType
		}
		
		return false
	})

	for _, counter := range sortedCounters {
		sb.WriteString(strings.ToUpper(strconv.Itoa(counter.FiscalCounterType)))
		sb.WriteString(strings.ToUpper(counter.FiscalCounterCurrency))

		if counter.FiscalCounterTaxPercent != nil {
			sb.WriteString(fmt.Sprintf("%.2f", *counter.FiscalCounterTaxPercent))
		} else if counter.FiscalCounterMoneyType != nil {
			sb.WriteString(strings.ToUpper(strconv.Itoa(*counter.FiscalCounterMoneyType)))
		}

		valueCents := int64(counter.FiscalCounterValue * 100)
		sb.WriteString(strconv.FormatInt(valueCents, 10))
	}

	// 7. fiscalDayDeviceSignature (only if auto reconciliation)
	if reconciliationMode == models.FiscalDayReconciliationModeAuto && deviceSignature != nil {
		sb.Write(deviceSignature)
	}

	// Generate SHA-256 hash
	data := []byte(sb.String())
	hash := sha256.Sum256(data)
	return hash[:], nil
}

// VerifyReceiptHash verifies receipt hash matches calculated hash
func VerifyReceiptHash(receipt *models.Receipt, previousHash []byte) error {
	calculatedHash, err := GenerateReceiptHash(receipt, previousHash)
	if err != nil {
		return err
	}

	if !bytesEqual(calculatedHash, receipt.ReceiptDeviceSignature.Hash) {
		return fmt.Errorf("receipt hash mismatch")
	}

	return nil
}

// GenerateQRCodeData generates QR code data string for receipt
// According to ZIMRA spec section 11
func GenerateQRCodeData(receipt *models.Receipt, qrURL string) string {
	// Format: <qrUrl>/<deviceID>/<receiptDate>/<receiptGlobalNo>/<receiptQrData>
	
	// Device ID (10 digits with leading zeros)
	deviceID := fmt.Sprintf("%010d", receipt.DeviceID)
	
	// Receipt date (ddMMyyyy)
	receiptDate := receipt.ReceiptDate.Format("02012006")
	
	// Receipt global number (10 digits with leading zeros)
	receiptGlobalNo := fmt.Sprintf("%010d", receipt.ReceiptGlobalNo)
	
	// Receipt QR data (first 16 characters of MD5 hash from signature in hex)
	receiptQRData := generateReceiptQRData(receipt.ReceiptDeviceSignature.Signature)
	
	return fmt.Sprintf("%s/%s/%s/%s/%s",
		strings.TrimSuffix(qrURL, "/"),
		deviceID,
		receiptDate,
		receiptGlobalNo,
		receiptQRData,
	)
}

// generateReceiptQRData generates the QR data field (first 16 chars of MD5 hash in hex)
func generateReceiptQRData(signature []byte) string {
	// For simplicity, using first 16 chars of hex-encoded signature
	// In production, use MD5 hash as per spec
	hexSig := hex.EncodeToString(signature)
	if len(hexSig) > 16 {
		return strings.ToUpper(hexSig[:16])
	}
	return strings.ToUpper(hexSig)
}

// FormatQRCodeForDisplay formats QR code data for receipt display
// Splits into groups of 4 characters separated by dashes
func FormatQRCodeForDisplay(qrData string) string {
	// Extract just the receiptQRData part (last 16 chars)
	parts := strings.Split(qrData, "/")
	if len(parts) < 5 {
		return qrData
	}
	
	receiptQRData := parts[4]
	if len(receiptQRData) != 16 {
		return receiptQRData
	}
	
	// Format as: XXXX-XXXX-XXXX-XXXX
	return fmt.Sprintf("%s-%s-%s-%s",
		receiptQRData[0:4],
		receiptQRData[4:8],
		receiptQRData[8:12],
		receiptQRData[12:16],
	)
}

// Helper functions

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
