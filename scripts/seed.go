package main

import (
	"log"

	"fiscalization-api/internal/config"
	"fiscalization-api/internal/database"
	"fiscalization-api/internal/models"
	"fiscalization-api/internal/utils"

	"github.com/jmoiron/sqlx"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Starting database seeding...")

	// Seed taxpayers
	if err := seedTaxpayers(db); err != nil {
		log.Fatalf("Failed to seed taxpayers: %v", err)
	}

	// Seed taxes
	if err := seedTaxes(db); err != nil {
		log.Fatalf("Failed to seed taxes: %v", err)
	}

	// Seed devices
	if err := seedDevices(db); err != nil {
		log.Fatalf("Failed to seed devices: %v", err)
	}

	log.Println("Database seeding completed successfully!")
}

func seedTaxpayers(db *sqlx.DB) error {
	log.Println("Seeding taxpayers...")

	taxpayers := []struct {
		TIN       string
		Name      string
		VATNumber *string
	}{
		{
			TIN:       "1234567890",
			Name:      "ABC Retail Store Ltd",
			VATNumber: stringPtr("123456789"),
		},
		{
			TIN:       "0987654321",
			Name:      "XYZ Supermarket",
			VATNumber: stringPtr("987654321"),
		},
		{
			TIN:       "1122334455",
			Name:      "Small Shop",
			VATNumber: nil, // Non-VAT taxpayer
		},
	}

	for _, tp := range taxpayers {
		query := `
			INSERT INTO taxpayers (tin, name, vat_number, status, taxpayer_day_max_hrs, taxpayer_day_end_notification_hrs, qr_url)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (tin) DO NOTHING`

		_, err := db.Exec(query, tp.TIN, tp.Name, tp.VATNumber, "Active", 24, 2, "https://receipt.zimra.co.zw")
		if err != nil {
			return err
		}
	}

	log.Println("✓ Taxpayers seeded")
	return nil
}

func seedTaxes(db *sqlx.DB) error {
	log.Println("Seeding taxes...")

	taxes := []struct {
		TaxID      int
		TaxName    string
		TaxPercent *float64
		ValidFrom  string
		ValidTill  *string
	}{
		{
			TaxID:      1,
			TaxName:    "VAT 15%",
			TaxPercent: float64Ptr(15.00),
			ValidFrom:  "2020-01-01",
			ValidTill:  nil,
		},
		{
			TaxID:      2,
			TaxName:    "VAT 0%",
			TaxPercent: float64Ptr(0.00),
			ValidFrom:  "2020-01-01",
			ValidTill:  nil,
		},
		{
			TaxID:      3,
			TaxName:    "Exempt",
			TaxPercent: nil,
			ValidFrom:  "2020-01-01",
			ValidTill:  nil,
		},
	}

	for _, tax := range taxes {
		query := `
			INSERT INTO taxes (tax_id, tax_name, tax_percent, tax_valid_from, tax_valid_till)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (tax_id) DO NOTHING`

		_, err := db.Exec(query, tax.TaxID, tax.TaxName, tax.TaxPercent, tax.ValidFrom, tax.ValidTill)
		if err != nil {
			return err
		}
	}

	log.Println("✓ Taxes seeded")
	return nil
}

func seedDevices(db *sqlx.DB) error {
	log.Println("Seeding devices...")

	// Get first taxpayer ID
	var taxpayerID int64
	err := db.Get(&taxpayerID, "SELECT id FROM taxpayers LIMIT 1")
	if err != nil {
		return err
	}

	devices := []struct {
		DeviceID     int
		SerialNo     string
		ModelName    string
		ModelVersion string
		BranchName   string
	}{
		{
			DeviceID:     1001,
			SerialNo:     "DEV-001-2024",
			ModelName:    "ZIMRA-POS-2000",
			ModelVersion: "1.0",
			BranchName:   "Main Branch",
		},
		{
			DeviceID:     1002,
			SerialNo:     "DEV-002-2024",
			ModelName:    "ZIMRA-POS-2000",
			ModelVersion: "1.0",
			BranchName:   "Downtown Branch",
		},
	}

	for _, dev := range devices {
		// Generate activation key
		activationKey, err := utils.GenerateActivationKey()
		if err != nil {
			return err
		}

		query := `
			INSERT INTO devices (
				device_id, taxpayer_id, device_serial_no, device_model_name, device_model_version,
				activation_key, operating_mode, status, branch_name, branch_address
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (device_id) DO NOTHING`

		address := models.Address{
			Province: "Harare",
			City:     "Harare",
			Street:   "Samora Machel Avenue",
			HouseNo:  "123",
		}

		_, err = db.Exec(
			query,
			dev.DeviceID,
			taxpayerID,
			dev.SerialNo,
			dev.ModelName,
			dev.ModelVersion,
			activationKey,
			0, // Online mode
			"Active",
			dev.BranchName,
			address,
		)
		if err != nil {
			return err
		}

		log.Printf("  Device %d created with activation key: %s", dev.DeviceID, activationKey)
	}

	log.Println("✓ Devices seeded")
	return nil
}

func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
