package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"fiscalization-api/internal/models"

	"github.com/jmoiron/sqlx"
)

type DeviceRepository interface {
	// Device operations
	Create(device *models.Device) error
	GetByDeviceID(deviceID int) (*models.Device, error)
	GetBySerialNo(serialNo string) (*models.Device, error)
	Update(device *models.Device) error
	UpdateCertificate(deviceID int, cert string, thumbprint []byte, validTill time.Time) error
	UpdateLastPing(deviceID int, lastPing time.Time) error
	IsBlacklisted(modelName, modelVersion string) (bool, error)
	
	// Taxpayer operations
	GetTaxpayer(taxpayerID int64) (*models.Taxpayer, error)
	
	// Tax operations
	GetApplicableTaxes() ([]models.Tax, error)
	
	// Fiscal day operations
	GetCurrentFiscalDay(deviceID int) (*models.FiscalDay, error)
	GetFiscalDayCounters(fiscalDayID int64) ([]models.FiscalDayCounter, error)
	GetFiscalDayDocumentQuantities(fiscalDayID int64) ([]models.FiscalDayDocumentQuantity, error)
	
	// Certificate history
	SaveCertificateHistory(deviceID int, cert string, thumbprint []byte, validTill time.Time) error
	
	// Stock operations
	GetStockList(
		taxpayerID int64,
		branchID int64,
		hsCode *string,
		goodName *string,
		sort *string,
		order *string,
		offset int,
		limit int,
		operator *string,
	) (int, []models.Good, error)
}

type deviceRepository struct {
	db *sqlx.DB
}

func NewDeviceRepository(db *sqlx.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) Create(device *models.Device) error {
	query := `
		INSERT INTO devices (
			device_id, taxpayer_id, device_serial_no, device_model_name, device_model_version,
			activation_key, operating_mode, status, branch_name, branch_address, branch_contacts
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id, created_at, updated_at`

	return r.db.QueryRow(
		query,
		device.DeviceID,
		device.TaxpayerID,
		device.DeviceSerialNo,
		device.DeviceModelName,
		device.DeviceModelVersion,
		device.ActivationKey,
		device.OperatingMode,
		device.Status,
		device.BranchName,
		device.BranchAddress,
		device.BranchContacts,
	).Scan(&device.ID, &device.CreatedAt, &device.UpdatedAt)
}

func (r *deviceRepository) GetByDeviceID(deviceID int) (*models.Device, error) {
	var device models.Device
	query := `SELECT * FROM devices WHERE device_id = $1`
	
	err := r.db.Get(&device, query, deviceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &device, nil
}

func (r *deviceRepository) GetBySerialNo(serialNo string) (*models.Device, error) {
	var device models.Device
	query := `SELECT * FROM devices WHERE device_serial_no = $1`
	
	err := r.db.Get(&device, query, serialNo)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &device, nil
}

func (r *deviceRepository) Update(device *models.Device) error {
	query := `
		UPDATE devices SET
			device_serial_no = $1,
			device_model_name = $2,
			device_model_version = $3,
			operating_mode = $4,
			status = $5,
			branch_name = $6,
			branch_address = $7,
			branch_contacts = $8
		WHERE device_id = $9`

	_, err := r.db.Exec(
		query,
		device.DeviceSerialNo,
		device.DeviceModelName,
		device.DeviceModelVersion,
		device.OperatingMode,
		device.Status,
		device.BranchName,
		device.BranchAddress,
		device.BranchContacts,
		device.DeviceID,
	)
	
	return err
}

func (r *deviceRepository) UpdateCertificate(deviceID int, cert string, thumbprint []byte, validTill time.Time) error {
	query := `
		UPDATE devices SET
			certificate = $1,
			certificate_thumbprint = $2,
			certificate_valid_till = $3
		WHERE device_id = $4`

	_, err := r.db.Exec(query, cert, thumbprint, validTill, deviceID)
	return err
}

func (r *deviceRepository) UpdateLastPing(deviceID int, lastPing time.Time) error {
	query := `UPDATE devices SET updated_at = $1 WHERE device_id = $2`
	_, err := r.db.Exec(query, lastPing, deviceID)
	return err
}

func (r *deviceRepository) IsBlacklisted(modelName, modelVersion string) (bool, error) {
	// This would check against a blacklist table
	// For now, return false
	return false, nil
}

func (r *deviceRepository) GetTaxpayer(taxpayerID int64) (*models.Taxpayer, error) {
	var taxpayer models.Taxpayer
	query := `SELECT * FROM taxpayers WHERE id = $1`
	
	err := r.db.Get(&taxpayer, query, taxpayerID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &taxpayer, nil
}

func (r *deviceRepository) GetApplicableTaxes() ([]models.Tax, error) {
	var taxes []models.Tax
	query := `
		SELECT tax_id, tax_name, tax_percent, tax_valid_from, tax_valid_till
		FROM taxes
		WHERE tax_valid_from <= CURRENT_DATE
		  AND (tax_valid_till IS NULL OR tax_valid_till >= CURRENT_DATE)
		ORDER BY tax_id`
	
	err := r.db.Select(&taxes, query)
	if err != nil {
		return nil, err
	}
	
	return taxes, nil
}

func (r *deviceRepository) GetCurrentFiscalDay(deviceID int) (*models.FiscalDay, error) {
	var fiscalDay models.FiscalDay
	query := `
		SELECT * FROM fiscal_days
		WHERE device_id = $1
		ORDER BY fiscal_day_no DESC
		LIMIT 1`
	
	err := r.db.Get(&fiscalDay, query, deviceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &fiscalDay, nil
}

func (r *deviceRepository) GetFiscalDayCounters(fiscalDayID int64) ([]models.FiscalDayCounter, error) {
	var counters []models.FiscalDayCounter
	query := `
		SELECT * FROM fiscal_counters
		WHERE fiscal_day_id = $1
		  AND fiscal_counter_value != 0
		ORDER BY fiscal_counter_type, fiscal_counter_currency, fiscal_counter_tax_id`
	
	err := r.db.Select(&counters, query, fiscalDayID)
	if err != nil {
		return nil, err
	}
	
	return counters, nil
}

func (r *deviceRepository) GetFiscalDayDocumentQuantities(fiscalDayID int64) ([]models.FiscalDayDocumentQuantity, error) {
	var quantities []models.FiscalDayDocumentQuantity
	query := `
		SELECT
			receipt_type,
			receipt_currency,
			COUNT(*) as receipt_quantity,
			SUM(receipt_total) as receipt_total_amount
		FROM receipts
		WHERE fiscal_day_id = $1
		GROUP BY receipt_type, receipt_currency
		ORDER BY receipt_type, receipt_currency`
	
	err := r.db.Select(&quantities, query, fiscalDayID)
	if err != nil {
		return nil, err
	}
	
	return quantities, nil
}

func (r *deviceRepository) SaveCertificateHistory(deviceID int, cert string, thumbprint []byte, validTill time.Time) error {
	query := `
		INSERT INTO certificates_history (
			device_id, certificate, certificate_thumbprint, issued_at, valid_till
		) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(query, deviceID, cert, thumbprint, time.Now(), validTill)
	return err
}

func (r *deviceRepository) GetStockList(
	taxpayerID int64,
	branchID int64,
	hsCode *string,
	goodName *string,
	sort *string,
	order *string,
	offset int,
	limit int,
	operator *string,
) (int, []models.Good, error) {
	// Build query with filters
	baseQuery := `
		FROM stock s
		INNER JOIN taxpayers t ON s.taxpayer_id = t.id
		LEFT JOIN devices d ON s.branch_id = d.id
		WHERE s.taxpayer_id = $1`
	
	args := []interface{}{taxpayerID}
	argCount := 1

	if hsCode != nil && *hsCode != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND s.hs_code = $%d", argCount)
		args = append(args, *hsCode)
	}

	if goodName != nil && *goodName != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND s.good_name ILIKE $%d", argCount)
		args = append(args, "%"+*goodName+"%")
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return 0, nil, err
	}

	// Get data
	selectQuery := `
		SELECT
			s.hs_code,
			s.good_name,
			s.quantity,
			t.id as taxpayer_id,
			t.name as taxpayer_name,
			d.id as branch_id,
			d.branch_name
		` + baseQuery

	// Add sorting
	sortField := "s.hs_code"
	if sort != nil && *sort != "" {
		sortField = *sort
	}
	sortOrder := "ASC"
	if order != nil && strings.ToUpper(*order) == "DESC" {
		sortOrder = "DESC"
	}
	selectQuery += fmt.Sprintf(" ORDER BY %s %s", sortField, sortOrder)

	// Add pagination
	selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	var items []models.Good
	err = r.db.Select(&items, selectQuery, args...)
	if err != nil {
		return 0, nil, err
	}

	return total, items, nil
}
