package repository

import (
	"database/sql"
	"fmt"
	"time"

	"fiscalization-api/internal/models"

	"github.com/jmoiron/sqlx"
)

type FiscalDayRepository interface {
	// Fiscal day operations
	Create(fiscalDay *models.FiscalDay) error
	GetByID(id int64) (*models.FiscalDay, error)
	GetCurrent(deviceID int) (*models.FiscalDay, error)
	GetByDayNo(deviceID, fiscalDayNo int) (*models.FiscalDay, error)
	Update(fiscalDay *models.FiscalDay) error
	UpdateStatus(id int64, status models.FiscalDayStatus) error
	Close(id int64, closedAt time.Time, signature *models.SignatureData) error
	
	// Counter operations
	CreateCounters(fiscalDayID int64, counters []models.FiscalDayCounter) error
	GetCounters(fiscalDayID int64) ([]models.FiscalDayCounter, error)
	UpdateCounters(fiscalDayID int64, counters []models.FiscalDayCounter) error
	
	// Validation
	ValidateCounters(fiscalDayID int64, submittedCounters []models.FiscalDayCounter) (bool, error)
	GetLastClosedDay(deviceID int) (*models.FiscalDay, error)
}

type fiscalDayRepository struct {
	db *sqlx.DB
}

func NewFiscalDayRepository(db *sqlx.DB) FiscalDayRepository {
	return &fiscalDayRepository{db: db}
}

func (r *fiscalDayRepository) Create(fiscalDay *models.FiscalDay) error {
	query := `
		INSERT INTO fiscal_days (
			device_id, fiscal_day_no, fiscal_day_opened, status
		) VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(
		query,
		fiscalDay.DeviceID,
		fiscalDay.FiscalDayNo,
		fiscalDay.FiscalDayOpened,
		fiscalDay.Status,
	).Scan(&fiscalDay.ID, &fiscalDay.CreatedAt, &fiscalDay.UpdatedAt)
}

func (r *fiscalDayRepository) GetByID(id int64) (*models.FiscalDay, error) {
	var fiscalDay models.FiscalDay
	query := `SELECT * FROM fiscal_days WHERE id = $1`

	err := r.db.Get(&fiscalDay, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &fiscalDay, nil
}

func (r *fiscalDayRepository) GetCurrent(deviceID int) (*models.FiscalDay, error) {
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

func (r *fiscalDayRepository) GetByDayNo(deviceID, fiscalDayNo int) (*models.FiscalDay, error) {
	var fiscalDay models.FiscalDay
	query := `
		SELECT * FROM fiscal_days
		WHERE device_id = $1 AND fiscal_day_no = $2`

	err := r.db.Get(&fiscalDay, query, deviceID, fiscalDayNo)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &fiscalDay, nil
}

func (r *fiscalDayRepository) Update(fiscalDay *models.FiscalDay) error {
	query := `
		UPDATE fiscal_days SET
			fiscal_day_closed = $1,
			status = $2,
			reconciliation_mode = $3,
			fiscal_day_device_signature = $4,
			fiscal_day_server_signature = $5,
			closing_error_code = $6,
			last_receipt_global_no = $7
		WHERE id = $8`

	_, err := r.db.Exec(
		query,
		fiscalDay.FiscalDayClosed,
		fiscalDay.Status,
		fiscalDay.ReconciliationMode,
		fiscalDay.FiscalDayDeviceSignature,
		fiscalDay.FiscalDayServerSignature,
		fiscalDay.ClosingErrorCode,
		fiscalDay.LastReceiptGlobalNo,
		fiscalDay.ID,
	)

	return err
}

func (r *fiscalDayRepository) UpdateStatus(id int64, status models.FiscalDayStatus) error {
	query := `UPDATE fiscal_days SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *fiscalDayRepository) Close(id int64, closedAt time.Time, signature *models.SignatureData) error {
	query := `
		UPDATE fiscal_days SET
			fiscal_day_closed = $1,
			status = $2,
			fiscal_day_device_signature = $3
		WHERE id = $4`

	_, err := r.db.Exec(query, closedAt, models.FiscalDayStatusCloseInitiated, signature, id)
	return err
}

func (r *fiscalDayRepository) CreateCounters(fiscalDayID int64, counters []models.FiscalDayCounter) error {
	// Delete existing counters first
	deleteQuery := `DELETE FROM fiscal_counters WHERE fiscal_day_id = $1`
	_, err := r.db.Exec(deleteQuery, fiscalDayID)
	if err != nil {
		return err
	}

	// Insert new counters (only non-zero values)
	for _, counter := range counters {
		if counter.FiscalCounterValue == 0 {
			continue
		}

		query := `
			INSERT INTO fiscal_counters (
				fiscal_day_id, fiscal_counter_type, fiscal_counter_currency,
				fiscal_counter_tax_id, fiscal_counter_tax_percent,
				fiscal_counter_money_type, fiscal_counter_value
			) VALUES ($1, $2, $3, $4, $5, $6, $7)`

		_, err := r.db.Exec(
			query,
			fiscalDayID,
			counter.FiscalCounterType,
			counter.FiscalCounterCurrency,
			counter.FiscalCounterTaxID,
			counter.FiscalCounterTaxPercent,
			counter.FiscalCounterMoneyType,
			counter.FiscalCounterValue,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *fiscalDayRepository) GetCounters(fiscalDayID int64) ([]models.FiscalDayCounter, error) {
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

func (r *fiscalDayRepository) UpdateCounters(fiscalDayID int64, counters []models.FiscalDayCounter) error {
	return r.CreateCounters(fiscalDayID, counters)
}

func (r *fiscalDayRepository) ValidateCounters(fiscalDayID int64, submittedCounters []models.FiscalDayCounter) (bool, error) {
	// Calculate actual counters from receipts
	actualCounters, err := r.calculateActualCounters(fiscalDayID)
	if err != nil {
		return false, err
	}

	// Compare submitted vs actual
	return r.compareCounters(submittedCounters, actualCounters), nil
}

func (r *fiscalDayRepository) GetLastClosedDay(deviceID int) (*models.FiscalDay, error) {
	var fiscalDay models.FiscalDay
	query := `
		SELECT * FROM fiscal_days
		WHERE device_id = $1
		  AND status = $2
		ORDER BY fiscal_day_no DESC
		LIMIT 1`

	err := r.db.Get(&fiscalDay, query, deviceID, models.FiscalDayStatusClosed)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &fiscalDay, nil
}

// Helper methods

func (r *fiscalDayRepository) calculateActualCounters(fiscalDayID int64) ([]models.FiscalDayCounter, error) {
	counters := make([]models.FiscalDayCounter, 0)

	// Calculate SaleByTax counters
	saleByTaxQuery := `
		SELECT
			0 as fiscal_counter_type,
			rt.tax_id as fiscal_counter_tax_id,
			rt.tax_percent as fiscal_counter_tax_percent,
			r.receipt_currency as fiscal_counter_currency,
			SUM(rt.sales_amount_with_tax) as fiscal_counter_value
		FROM receipts r
		JOIN receipt_taxes rt ON r.id = rt.receipt_id
		WHERE r.fiscal_day_id = $1
		  AND r.receipt_type = $2
		GROUP BY rt.tax_id, rt.tax_percent, r.receipt_currency`

	var saleCounters []models.FiscalDayCounter
	err := r.db.Select(&saleCounters, saleByTaxQuery, fiscalDayID, models.ReceiptTypeFiscalInvoice)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	counters = append(counters, saleCounters...)

	// Calculate SaleTaxByTax counters
	saleTaxQuery := `
		SELECT
			1 as fiscal_counter_type,
			rt.tax_id as fiscal_counter_tax_id,
			rt.tax_percent as fiscal_counter_tax_percent,
			r.receipt_currency as fiscal_counter_currency,
			SUM(rt.tax_amount) as fiscal_counter_value
		FROM receipts r
		JOIN receipt_taxes rt ON r.id = rt.receipt_id
		WHERE r.fiscal_day_id = $1
		  AND r.receipt_type = $2
		GROUP BY rt.tax_id, rt.tax_percent, r.receipt_currency`

	var saleTaxCounters []models.FiscalDayCounter
	err = r.db.Select(&saleTaxCounters, saleTaxQuery, fiscalDayID, models.ReceiptTypeFiscalInvoice)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	counters = append(counters, saleTaxCounters...)

	// Calculate CreditNoteByTax counters
	creditQuery := `
		SELECT
			2 as fiscal_counter_type,
			rt.tax_id as fiscal_counter_tax_id,
			rt.tax_percent as fiscal_counter_tax_percent,
			r.receipt_currency as fiscal_counter_currency,
			SUM(rt.sales_amount_with_tax) as fiscal_counter_value
		FROM receipts r
		JOIN receipt_taxes rt ON r.id = rt.receipt_id
		WHERE r.fiscal_day_id = $1
		  AND r.receipt_type = $2
		GROUP BY rt.tax_id, rt.tax_percent, r.receipt_currency`

	var creditCounters []models.FiscalDayCounter
	err = r.db.Select(&creditCounters, creditQuery, fiscalDayID, models.ReceiptTypeCreditNote)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	counters = append(counters, creditCounters...)

	// Calculate DebitNoteByTax counters
	debitQuery := `
		SELECT
			4 as fiscal_counter_type,
			rt.tax_id as fiscal_counter_tax_id,
			rt.tax_percent as fiscal_counter_tax_percent,
			r.receipt_currency as fiscal_counter_currency,
			SUM(rt.sales_amount_with_tax) as fiscal_counter_value
		FROM receipts r
		JOIN receipt_taxes rt ON r.id = rt.receipt_id
		WHERE r.fiscal_day_id = $1
		  AND r.receipt_type = $2
		GROUP BY rt.tax_id, rt.tax_percent, r.receipt_currency`

	var debitCounters []models.FiscalDayCounter
	err = r.db.Select(&debitCounters, debitQuery, fiscalDayID, models.ReceiptTypeDebitNote)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	counters = append(counters, debitCounters...)

	// Calculate BalanceByMoneyType counters
	balanceQuery := `
		SELECT
			6 as fiscal_counter_type,
			rp.money_type_code as fiscal_counter_money_type,
			r.receipt_currency as fiscal_counter_currency,
			SUM(rp.payment_amount) as fiscal_counter_value
		FROM receipts r
		JOIN receipt_payments rp ON r.id = rp.receipt_id
		WHERE r.fiscal_day_id = $1
		GROUP BY rp.money_type_code, r.receipt_currency`

	var balanceCounters []models.FiscalDayCounter
	err = r.db.Select(&balanceCounters, balanceQuery, fiscalDayID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	counters = append(counters, balanceCounters...)

	return counters, nil
}

func (r *fiscalDayRepository) compareCounters(submitted, actual []models.FiscalDayCounter) bool {
	// Create maps for easy comparison
	submittedMap := make(map[string]float64)
	actualMap := make(map[string]float64)

	for _, c := range submitted {
		key := r.counterKey(c)
		submittedMap[key] = c.FiscalCounterValue
	}

	for _, c := range actual {
		key := r.counterKey(c)
		actualMap[key] = c.FiscalCounterValue
	}

	// Check if all keys match and values are close (allowing small rounding differences)
	if len(submittedMap) != len(actualMap) {
		return false
	}

	for key, submittedVal := range submittedMap {
		actualVal, exists := actualMap[key]
		if !exists {
			return false
		}

		// Allow 0.01 difference for rounding
		diff := submittedVal - actualVal
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.01 {
			return false
		}
	}

	return true
}

func (r *fiscalDayRepository) counterKey(c models.FiscalDayCounter) string {
	key := string(c.FiscalCounterType) + "_" + c.FiscalCounterCurrency + "_"

	if c.FiscalCounterTaxID != nil {
		key += string(*c.FiscalCounterTaxID) + "_"
	}
	if c.FiscalCounterTaxPercent != nil {
		key += fmt.Sprintf("%.2f", *c.FiscalCounterTaxPercent) + "_"
	}
	if c.FiscalCounterMoneyType != nil {
		key += string(*c.FiscalCounterMoneyType)
	}

	return key
}
