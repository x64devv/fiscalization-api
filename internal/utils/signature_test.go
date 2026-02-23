package utils

import (
	"testing"
	"time"

	"fiscalization-api/internal/models"
)

func TestGenerateReceiptHash(t *testing.T) {
	receipt := &models.Receipt{
		DeviceID:        1001,
		ReceiptType:     models.ReceiptTypeFiscalInvoice,
		ReceiptCurrency: "USD",
		ReceiptGlobalNo: 1,
		ReceiptDate:     time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		ReceiptTotal:    100.00,
		ReceiptTaxes: []models.ReceiptTax{
			{
				TaxID:              1,
				TaxCode:            stringPtr("A"),
				TaxPercent:         float64Ptr(15.0),
				TaxAmount:          13.04,
				SalesAmountWithTax: 100.00,
			},
		},
	}

	hash1, err := GenerateReceiptHash(receipt, nil)
	if err != nil {
		t.Fatalf("GenerateReceiptHash() error = %v", err)
	}

	if len(hash1) != 32 { // SHA-256 produces 32 bytes
		t.Errorf("GenerateReceiptHash() hash length = %d, want 32", len(hash1))
	}

	// Same receipt should produce same hash
	hash2, err := GenerateReceiptHash(receipt, nil)
	if err != nil {
		t.Fatalf("GenerateReceiptHash() error = %v", err)
	}

	if !bytesEqual(hash1, hash2) {
		t.Error("Same receipt produced different hashes")
	}

	// Different receipt should produce different hash
	receipt2 := *receipt
	receipt2.ReceiptTotal = 200.00
	hash3, err := GenerateReceiptHash(&receipt2, nil)
	if err != nil {
		t.Fatalf("GenerateReceiptHash() error = %v", err)
	}

	if bytesEqual(hash1, hash3) {
		t.Error("Different receipts produced same hash")
	}
}

func TestGenerateFiscalDayHash(t *testing.T) {
	counters := []models.FiscalDayCounter{
		{
			//ToDo: Update to use actual enum values when available
			// FiscalCounterType:      models.FiscalCounterTypeSaleByTax,
			FiscalCounterType:      1,
			FiscalCounterCurrency:  "USD",
			FiscalCounterTaxID:     intPtr(1),
			FiscalCounterTaxPercent: float64Ptr(15.0),
			FiscalCounterValue:     100.00,
		},
	}

	hash1, err := GenerateFiscalDayHash(1001, 1, "2024-01-01", counters)
	if err != nil {
		t.Fatalf("GenerateFiscalDayHash() error = %v", err)
	}

	if len(hash1) != 32 {
		t.Errorf("GenerateFiscalDayHash() hash length = %d, want 32", len(hash1))
	}

	// Same data should produce same hash
	hash2, err := GenerateFiscalDayHash(1001, 1, "2024-01-01", counters)
	if err != nil {
		t.Fatalf("GenerateFiscalDayHash() error = %v", err)
	}

	if !bytesEqual(hash1, hash2) {
		t.Error("Same fiscal day data produced different hashes")
	}

	// Different day number should produce different hash
	hash3, err := GenerateFiscalDayHash(1001, 2, "2024-01-01", counters)
	if err != nil {
		t.Fatalf("GenerateFiscalDayHash() error = %v", err)
	}

	if bytesEqual(hash1, hash3) {
		t.Error("Different fiscal day numbers produced same hash")
	}
}

func TestGenerateQRCodeData(t *testing.T) {
	receipt := &models.Receipt{
		DeviceID:        1001,
		ReceiptDate:     time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		ReceiptGlobalNo: 123,
		ReceiptDeviceSignature: models.SignatureData{
			Hash:      []byte{0x01, 0x02, 0x03, 0x04},
			Signature: []byte{0xAA, 0xBB, 0xCC, 0xDD},
		},
	}

	qrURL := "https://receipt.zimra.co.zw"
	qrData := GenerateQRCodeData(receipt, qrURL)

	// Check format: <qrUrl>/<deviceID>/<date>/<globalNo>/<qrData>
	expected := "https://receipt.zimra.co.zw/0000001001/15012024/0000000123/"
	if len(qrData) < len(expected) {
		t.Errorf("QR code data too short: %s", qrData)
	}

	if qrData[:len(expected)] != expected {
		t.Errorf("QR code data = %s, want prefix %s", qrData, expected)
	}
}

func TestFormatQRCodeForDisplay(t *testing.T) {
	qrData := "https://receipt.zimra.co.zw/0000001001/15012024/0000000123/AABBCCDDEEFF0011"
	formatted := FormatQRCodeForDisplay(qrData)

	expected := "AABB-CCDD-EEFF-0011"
	if formatted != expected {
		t.Errorf("FormatQRCodeForDisplay() = %s, want %s", formatted, expected)
	}
}

func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
