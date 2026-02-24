package repository

import (
	"database/sql"
	"fmt"
	"time"

	"fiscalization-api/internal/models"

	"github.com/jmoiron/sqlx"
)

// AdminRepository handles all system-owner operations across all tenants
type AdminRepository interface {
	// Taxpayer (company) management
	CreateTaxpayer(tp *models.Taxpayer) error
	GetTaxpayerByID(id int64) (*models.Taxpayer, error)
	GetTaxpayerByTIN(tin string) (*models.Taxpayer, error)
	ListTaxpayers(offset, limit int, search string) (int, []models.Taxpayer, error)
	UpdateTaxpayer(tp *models.Taxpayer) error
	SetTaxpayerStatus(id int64, status string) error

	// Device management across tenants
	CreateDevice(device *models.Device) error
	ListDevicesByTaxpayer(taxpayerID int64) ([]models.Device, error)
	ListAllDevices(offset, limit int) (int, []models.Device, error)
	GetDeviceByID(deviceID int) (*models.Device, error)
	UpdateDeviceStatus(deviceID int, status string) error
	UpdateDeviceMode(deviceID int, mode int) error

	// Cross-tenant fiscal day overview
	ListFiscalDays(taxpayerID *int64, deviceID *int, offset, limit int) (int, []models.FiscalDay, error)

	// Cross-tenant receipt overview
	ListReceipts(taxpayerID *int64, deviceID *int, from, to *time.Time, offset, limit int) (int, []models.AdminReceiptRow, error)

	// Audit logs
	ListAuditLogs(entityType string, entityID *int64, offset, limit int) (int, []models.AuditLog, error)
	InsertAuditLog(entityType, action string, entityID *int64, deviceID *int, ipAddress, details string) error

	// System stats
	GetSystemStats() (*models.SystemStats, error)
}

type adminRepository struct {
	db *sqlx.DB
}

func NewAdminRepository(db *sqlx.DB) AdminRepository {
	return &adminRepository{db: db}
}

// ─── Taxpayer ─────────────────────────────────────────────────────────────────

func (r *adminRepository) CreateTaxpayer(tp *models.Taxpayer) error {
	query := `
		INSERT INTO taxpayers (tin, name, vat_number, status, taxpayer_day_max_hrs, taxpayer_day_end_notification_hrs, qr_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query,
		tp.TIN, tp.Name, tp.VATNumber, tp.Status,
		tp.TaxPayerDayMaxHrs, tp.TaxpayerDayEndNotificationHrs, tp.QrURL,
	).Scan(&tp.ID, &tp.CreatedAt, &tp.UpdatedAt)
}

func (r *adminRepository) GetTaxpayerByID(id int64) (*models.Taxpayer, error) {
	var tp models.Taxpayer
	err := r.db.Get(&tp, `SELECT * FROM taxpayers WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tp, err
}

func (r *adminRepository) GetTaxpayerByTIN(tin string) (*models.Taxpayer, error) {
	var tp models.Taxpayer
	err := r.db.Get(&tp, `SELECT * FROM taxpayers WHERE tin = $1`, tin)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tp, err
}

func (r *adminRepository) ListTaxpayers(offset, limit int, search string) (int, []models.Taxpayer, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argc := 0

	if search != "" {
		argc++
		where += fmt.Sprintf(" AND (name ILIKE $%d OR tin ILIKE $%d)", argc, argc)
		args = append(args, "%"+search+"%")
	}

	var total int
	if err := r.db.Get(&total, "SELECT COUNT(*) FROM taxpayers "+where, args...); err != nil {
		return 0, nil, err
	}

	argc++
	argc2 := argc + 1
	args = append(args, limit, offset)
	var rows []models.Taxpayer
	err := r.db.Select(&rows,
		fmt.Sprintf("SELECT * FROM taxpayers %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d", where, argc, argc2),
		args...)
	return total, rows, err
}

func (r *adminRepository) UpdateTaxpayer(tp *models.Taxpayer) error {
	_, err := r.db.Exec(`
		UPDATE taxpayers SET
			name = $1, vat_number = $2, status = $3,
			taxpayer_day_max_hrs = $4, taxpayer_day_end_notification_hrs = $5, qr_url = $6
		WHERE id = $7`,
		tp.Name, tp.VATNumber, tp.Status,
		tp.TaxPayerDayMaxHrs, tp.TaxpayerDayEndNotificationHrs, tp.QrURL, tp.ID)
	return err
}

func (r *adminRepository) SetTaxpayerStatus(id int64, status string) error {
	_, err := r.db.Exec(`UPDATE taxpayers SET status = $1 WHERE id = $2`, status, id)
	return err
}

// ─── Device ───────────────────────────────────────────────────────────────────

func (r *adminRepository) CreateDevice(device *models.Device) error {
	query := `
		INSERT INTO devices (
			device_id, taxpayer_id, device_serial_no, device_model_name,
			device_model_version, activation_key, operating_mode, status,
			branch_name, branch_address, branch_contacts
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query,
		device.DeviceID, device.TaxpayerID, device.DeviceSerialNo,
		device.DeviceModelName, device.DeviceModelVersion, device.ActivationKey,
		device.OperatingMode, device.Status, device.BranchName,
		device.BranchAddress, device.BranchContacts,
	).Scan(&device.ID, &device.CreatedAt, &device.UpdatedAt)
}

func (r *adminRepository) ListDevicesByTaxpayer(taxpayerID int64) ([]models.Device, error) {
	var rows []models.Device
	err := r.db.Select(&rows,
		`SELECT * FROM devices WHERE taxpayer_id = $1 ORDER BY device_id`, taxpayerID)
	return rows, err
}

func (r *adminRepository) ListAllDevices(offset, limit int) (int, []models.Device, error) {
	var total int
	if err := r.db.Get(&total, `SELECT COUNT(*) FROM devices`); err != nil {
		return 0, nil, err
	}
	var rows []models.Device
	err := r.db.Select(&rows,
		`SELECT * FROM devices ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	return total, rows, err
}

func (r *adminRepository) GetDeviceByID(deviceID int) (*models.Device, error) {
	var d models.Device
	err := r.db.Get(&d, `SELECT * FROM devices WHERE device_id = $1`, deviceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &d, err
}

func (r *adminRepository) UpdateDeviceStatus(deviceID int, status string) error {
	_, err := r.db.Exec(`UPDATE devices SET status = $1 WHERE device_id = $2`, status, deviceID)
	return err
}

func (r *adminRepository) UpdateDeviceMode(deviceID int, mode int) error {
	_, err := r.db.Exec(`UPDATE devices SET operating_mode = $1 WHERE device_id = $2`, mode, deviceID)
	return err
}

// ─── Fiscal Days ──────────────────────────────────────────────────────────────

func (r *adminRepository) ListFiscalDays(taxpayerID *int64, deviceID *int, offset, limit int) (int, []models.FiscalDay, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argc := 0

	if taxpayerID != nil {
		argc++
		where += fmt.Sprintf(" AND d.taxpayer_id = $%d", argc)
		args = append(args, *taxpayerID)
	}
	if deviceID != nil {
		argc++
		where += fmt.Sprintf(" AND f.device_id = $%d", argc)
		args = append(args, *deviceID)
	}

	joinQ := `FROM fiscal_days f JOIN devices d ON f.device_id = d.device_id ` + where

	var total int
	if err := r.db.Get(&total, "SELECT COUNT(*) "+joinQ, args...); err != nil {
		return 0, nil, err
	}

	argc++
	argc2 := argc + 1
	args = append(args, limit, offset)
	var rows []models.FiscalDay
	err := r.db.Select(&rows,
		fmt.Sprintf("SELECT f.* "+joinQ+" ORDER BY f.fiscal_day_opened DESC LIMIT $%d OFFSET $%d", argc, argc2),
		args...)
	return total, rows, err
}

// ─── Receipts ─────────────────────────────────────────────────────────────────

func (r *adminRepository) ListReceipts(taxpayerID *int64, deviceID *int, from, to *time.Time, offset, limit int) (int, []models.AdminReceiptRow, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argc := 0

	if taxpayerID != nil {
		argc++
		where += fmt.Sprintf(" AND d.taxpayer_id = $%d", argc)
		args = append(args, *taxpayerID)
	}
	if deviceID != nil {
		argc++
		where += fmt.Sprintf(" AND r.device_id = $%d", argc)
		args = append(args, *deviceID)
	}
	if from != nil {
		argc++
		where += fmt.Sprintf(" AND r.receipt_date >= $%d", argc)
		args = append(args, *from)
	}
	if to != nil {
		argc++
		where += fmt.Sprintf(" AND r.receipt_date <= $%d", argc)
		args = append(args, *to)
	}

	joinQ := `FROM receipts r JOIN devices d ON r.device_id = d.device_id ` + where

	var total int
	if err := r.db.Get(&total, "SELECT COUNT(*) "+joinQ, args...); err != nil {
		return 0, nil, err
	}

	argc++
	argc2 := argc + 1
	args = append(args, limit, offset)

	// Use explicit column list so validation_color is included (models.Receipt has json:"-" on it)
	selectQ := fmt.Sprintf(`
		SELECT r.id, r.receipt_id, r.device_id, r.receipt_type, r.receipt_currency,
		       r.invoice_no, r.receipt_date, r.receipt_total, r.validation_color, r.server_date
		`+joinQ+` ORDER BY r.receipt_date DESC LIMIT $%d OFFSET $%d`, argc, argc2)

	var rows []models.AdminReceiptRow
	err := r.db.Select(&rows, selectQ, args...)
	return total, rows, err
}

// ─── Audit Logs ───────────────────────────────────────────────────────────────

func (r *adminRepository) ListAuditLogs(entityType string, entityID *int64, offset, limit int) (int, []models.AuditLog, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argc := 0

	if entityType != "" {
		argc++
		where += fmt.Sprintf(" AND entity_type = $%d", argc)
		args = append(args, entityType)
	}
	if entityID != nil {
		argc++
		where += fmt.Sprintf(" AND entity_id = $%d", argc)
		args = append(args, *entityID)
	}

	var total int
	if err := r.db.Get(&total, "SELECT COUNT(*) FROM audit_logs "+where, args...); err != nil {
		return 0, nil, err
	}

	argc++
	argc2 := argc + 1
	args = append(args, limit, offset)
	var rows []models.AuditLog
	err := r.db.Select(&rows,
		fmt.Sprintf("SELECT * FROM audit_logs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
			where, argc, argc2), args...)
	return total, rows, err
}

func (r *adminRepository) InsertAuditLog(entityType, action string, entityID *int64, deviceID *int, ipAddress, details string) error {
	_, err := r.db.Exec(`
		INSERT INTO audit_logs (entity_type, action, entity_id, device_id, ip_address, details)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		entityType, action, entityID, deviceID, ipAddress, details)
	return err
}
// ─── System Stats ─────────────────────────────────────────────────────────────

func (r *adminRepository) GetSystemStats() (*models.SystemStats, error) {
	stats := &models.SystemStats{}

	queries := []struct {
		dest  *int
		query string
	}{
		{&stats.TotalCompanies, `SELECT COUNT(*) FROM taxpayers`},
		{&stats.ActiveCompanies, `SELECT COUNT(*) FROM taxpayers WHERE status = 'Active'`},
		{&stats.TotalDevices, `SELECT COUNT(*) FROM devices`},
		{&stats.ActiveDevices, `SELECT COUNT(*) FROM devices WHERE status = 'Active'`},
		{&stats.TodayReceipts, `SELECT COUNT(*) FROM receipts WHERE receipt_date >= CURRENT_DATE`},
		{&stats.OpenFiscalDays, `SELECT COUNT(*) FROM fiscal_days WHERE status = 1`},
		{&stats.ValidationErrors, `SELECT COUNT(*) FROM receipts WHERE validation_color IS NOT NULL AND validation_color != '' AND receipt_date >= CURRENT_DATE`},
	}

	for _, q := range queries {
		if err := r.db.Get(q.dest, q.query); err != nil {
			*q.dest = 0
		}
	}

	// Today's revenue
	r.db.Get(&stats.TodayRevenue,
		`SELECT COALESCE(SUM(receipt_total), 0) FROM receipts WHERE receipt_date >= CURRENT_DATE AND receipt_type = 0`)

	return stats, nil
}
