package repository

import (
	"database/sql"
	"fmt"

	"fiscalization-api/internal/models"

	"github.com/jmoiron/sqlx"
)

type ReceiptRepository interface {
	// Receipt operations
	Create(receipt *models.Receipt) error
	CreateWithLines(receipt *models.Receipt) error
	GetByID(id int64) (*models.Receipt, error)
	GetByReceiptID(receiptID int64) (*models.Receipt, error)
	GetByGlobalNo(deviceID, globalNo int) (*models.Receipt, error)
	GetPreviousReceipt(deviceID int, fiscalDayID int64, globalNo int) (*models.Receipt, error)
	Update(receipt *models.Receipt) error
	UpdateValidation(receiptID int64, color *models.ValidationColor, errors []string) error
	
	// Receipt lines
	CreateReceiptLines(receiptID int64, lines []models.ReceiptLine) error
	GetReceiptLines(receiptID int64) ([]models.ReceiptLine, error)
	
	// Receipt taxes
	CreateReceiptTaxes(receiptID int64, taxes []models.ReceiptTax) error
	GetReceiptTaxes(receiptID int64) ([]models.ReceiptTax, error)
	
	// Receipt payments
	CreateReceiptPayments(receiptID int64, payments []models.Payment) error
	GetReceiptPayments(receiptID int64) ([]models.Payment, error)
	
	// Validation and queries
	CheckInvoiceNoUnique(taxpayerID int64, invoiceNo string) (bool, error)
	GetMissingReceipts(deviceID int, fiscalDayID int64) ([]int, error)
	GetReceiptsWithValidationErrors(fiscalDayID int64) ([]models.Receipt, error)
	GetCreditDebitNotes(originalReceiptID int64) ([]*models.Receipt, []*models.Receipt, error)
}

type receiptRepository struct {
	db *sqlx.DB
}

func NewReceiptRepository(db *sqlx.DB) ReceiptRepository {
	return &receiptRepository{db: db}
}

func (r *receiptRepository) Create(receipt *models.Receipt) error {
	query := `
		INSERT INTO receipts (
			device_id, fiscal_day_id, receipt_type, receipt_currency, receipt_counter,
			receipt_global_no, invoice_no, buyer_data, receipt_notes, receipt_date,
			credit_debit_note, receipt_lines_tax_inclusive, receipt_total,
			receipt_print_form, receipt_device_signature, receipt_hash,
			username, user_name_surname
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		) RETURNING id, receipt_id, created_at, updated_at`

	return r.db.QueryRow(
		query,
		receipt.DeviceID,
		receipt.FiscalDayID,
		receipt.ReceiptType,
		receipt.ReceiptCurrency,
		receipt.ReceiptCounter,
		receipt.ReceiptGlobalNo,
		receipt.InvoiceNo,
		receipt.BuyerData,
		receipt.ReceiptNotes,
		receipt.ReceiptDate,
		receipt.CreditDebitNote,
		receipt.ReceiptLinesTaxInclusive,
		receipt.ReceiptTotal,
		receipt.ReceiptPrintForm,
		receipt.ReceiptDeviceSignature,
		receipt.ReceiptHash,
		receipt.Username,
		receipt.UserNameSurname,
	).Scan(&receipt.ID, &receipt.ReceiptID, &receipt.CreatedAt, &receipt.UpdatedAt)
}

func (r *receiptRepository) CreateWithLines(receipt *models.Receipt) error {
	// Start transaction
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create receipt
	query := `
		INSERT INTO receipts (
			device_id, fiscal_day_id, receipt_type, receipt_currency, receipt_counter,
			receipt_global_no, invoice_no, buyer_data, receipt_notes, receipt_date,
			credit_debit_note, receipt_lines_tax_inclusive, receipt_total,
			receipt_print_form, receipt_device_signature, receipt_hash,
			username, user_name_surname
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		) RETURNING id, receipt_id, created_at, updated_at`

	err = tx.QueryRow(
		query,
		receipt.DeviceID,
		receipt.FiscalDayID,
		receipt.ReceiptType,
		receipt.ReceiptCurrency,
		receipt.ReceiptCounter,
		receipt.ReceiptGlobalNo,
		receipt.InvoiceNo,
		receipt.BuyerData,
		receipt.ReceiptNotes,
		receipt.ReceiptDate,
		receipt.CreditDebitNote,
		receipt.ReceiptLinesTaxInclusive,
		receipt.ReceiptTotal,
		receipt.ReceiptPrintForm,
		receipt.ReceiptDeviceSignature,
		receipt.ReceiptHash,
		receipt.Username,
		receipt.UserNameSurname,
	).Scan(&receipt.ID, &receipt.ReceiptID, &receipt.CreatedAt, &receipt.UpdatedAt)
	if err != nil {
		return err
	}

	// Create receipt lines
	for _, line := range receipt.ReceiptLines {
		lineQuery := `
			INSERT INTO receipt_lines (
				receipt_id, receipt_line_type, receipt_line_no, receipt_line_hs_code,
				receipt_line_name, receipt_line_price, receipt_line_quantity,
				receipt_line_total, tax_code, tax_percent, tax_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

		_, err = tx.Exec(
			lineQuery,
			receipt.ID,
			line.ReceiptLineType,
			line.ReceiptLineNo,
			line.ReceiptLineHSCode,
			line.ReceiptLineName,
			line.ReceiptLinePrice,
			line.ReceiptLineQuantity,
			line.ReceiptLineTotal,
			line.TaxCode,
			line.TaxPercent,
			line.TaxID,
		)
		if err != nil {
			return err
		}
	}

	// Create receipt taxes
	for _, tax := range receipt.ReceiptTaxes {
		taxQuery := `
			INSERT INTO receipt_taxes (
				receipt_id, tax_code, tax_percent, tax_id, tax_amount, sales_amount_with_tax
			) VALUES ($1, $2, $3, $4, $5, $6)`

		_, err = tx.Exec(
			taxQuery,
			receipt.ID,
			tax.TaxCode,
			tax.TaxPercent,
			tax.TaxID,
			tax.TaxAmount,
			tax.SalesAmountWithTax,
		)
		if err != nil {
			return err
		}
	}

	// Create receipt payments
	for _, payment := range receipt.ReceiptPayments {
		paymentQuery := `
			INSERT INTO receipt_payments (
				receipt_id, money_type_code, payment_amount
			) VALUES ($1, $2, $3)`

		_, err = tx.Exec(paymentQuery, receipt.ID, payment.MoneyTypeCode, payment.PaymentAmount)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *receiptRepository) GetByID(id int64) (*models.Receipt, error) {
	var receipt models.Receipt
	query := `SELECT * FROM receipts WHERE id = $1`

	err := r.db.Get(&receipt, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load related data
	if err := r.loadReceiptRelations(&receipt); err != nil {
		return nil, err
	}

	return &receipt, nil
}

func (r *receiptRepository) GetByReceiptID(receiptID int64) (*models.Receipt, error) {
	var receipt models.Receipt
	query := `SELECT * FROM receipts WHERE receipt_id = $1`

	err := r.db.Get(&receipt, query, receiptID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := r.loadReceiptRelations(&receipt); err != nil {
		return nil, err
	}

	return &receipt, nil
}

func (r *receiptRepository) GetByGlobalNo(deviceID, globalNo int) (*models.Receipt, error) {
	var receipt models.Receipt
	query := `SELECT * FROM receipts WHERE device_id = $1 AND receipt_global_no = $2`

	err := r.db.Get(&receipt, query, deviceID, globalNo)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := r.loadReceiptRelations(&receipt); err != nil {
		return nil, err
	}

	return &receipt, nil
}

func (r *receiptRepository) GetPreviousReceipt(deviceID int, fiscalDayID int64, globalNo int) (*models.Receipt, error) {
	var receipt models.Receipt
	query := `
		SELECT * FROM receipts
		WHERE device_id = $1 AND fiscal_day_id = $2 AND receipt_global_no < $3
		ORDER BY receipt_global_no DESC
		LIMIT 1`

	err := r.db.Get(&receipt, query, deviceID, fiscalDayID, globalNo)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := r.loadReceiptRelations(&receipt); err != nil {
		return nil, err
	}

	return &receipt, nil
}

func (r *receiptRepository) Update(receipt *models.Receipt) error {
	query := `
		UPDATE receipts SET
			receipt_server_signature = $1,
			server_date = $2,
			validation_color = $3,
			validation_errors = $4
		WHERE id = $5`

	_, err := r.db.Exec(
		query,
		receipt.ReceiptServerSignature,
		receipt.ServerDate,
		receipt.ValidationColor,
		receipt.ValidationErrors,
		receipt.ID,
	)
	return err
}

func (r *receiptRepository) UpdateValidation(receiptID int64, color *models.ValidationColor, errors []string) error {
	query := `
		UPDATE receipts SET
			validation_color = $1,
			validation_errors = $2
		WHERE id = $3`

	_, err := r.db.Exec(query, color, errors, receiptID)
	return err
}

func (r *receiptRepository) CreateReceiptLines(receiptID int64, lines []models.ReceiptLine) error {
	for _, line := range lines {
		query := `
			INSERT INTO receipt_lines (
				receipt_id, receipt_line_type, receipt_line_no, receipt_line_hs_code,
				receipt_line_name, receipt_line_price, receipt_line_quantity,
				receipt_line_total, tax_code, tax_percent, tax_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

		_, err := r.db.Exec(
			query,
			receiptID,
			line.ReceiptLineType,
			line.ReceiptLineNo,
			line.ReceiptLineHSCode,
			line.ReceiptLineName,
			line.ReceiptLinePrice,
			line.ReceiptLineQuantity,
			line.ReceiptLineTotal,
			line.TaxCode,
			line.TaxPercent,
			line.TaxID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *receiptRepository) GetReceiptLines(receiptID int64) ([]models.ReceiptLine, error) {
	var lines []models.ReceiptLine
	query := `SELECT * FROM receipt_lines WHERE receipt_id = $1 ORDER BY receipt_line_no`

	err := r.db.Select(&lines, query, receiptID)
	return lines, err
}

func (r *receiptRepository) CreateReceiptTaxes(receiptID int64, taxes []models.ReceiptTax) error {
	for _, tax := range taxes {
		query := `
			INSERT INTO receipt_taxes (
				receipt_id, tax_code, tax_percent, tax_id, tax_amount, sales_amount_with_tax
			) VALUES ($1, $2, $3, $4, $5, $6)`

		_, err := r.db.Exec(
			query,
			receiptID,
			tax.TaxCode,
			tax.TaxPercent,
			tax.TaxID,
			tax.TaxAmount,
			tax.SalesAmountWithTax,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *receiptRepository) GetReceiptTaxes(receiptID int64) ([]models.ReceiptTax, error) {
	var taxes []models.ReceiptTax
	query := `SELECT * FROM receipt_taxes WHERE receipt_id = $1 ORDER BY tax_id`

	err := r.db.Select(&taxes, query, receiptID)
	return taxes, err
}

func (r *receiptRepository) CreateReceiptPayments(receiptID int64, payments []models.Payment) error {
	for _, payment := range payments {
		query := `
			INSERT INTO receipt_payments (receipt_id, money_type_code, payment_amount)
			VALUES ($1, $2, $3)`

		_, err := r.db.Exec(query, receiptID, payment.MoneyTypeCode, payment.PaymentAmount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *receiptRepository) GetReceiptPayments(receiptID int64) ([]models.Payment, error) {
	var payments []models.Payment
	query := `SELECT * FROM receipt_payments WHERE receipt_id = $1`

	err := r.db.Select(&payments, query, receiptID)
	return payments, err
}

func (r *receiptRepository) CheckInvoiceNoUnique(taxpayerID int64, invoiceNo string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM receipts r
		JOIN devices d ON r.device_id = d.device_id
		WHERE d.taxpayer_id = $1 AND r.invoice_no = $2`

	err := r.db.Get(&count, query, taxpayerID, invoiceNo)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (r *receiptRepository) GetMissingReceipts(deviceID int, fiscalDayID int64) ([]int, error) {
	// Find gaps in receipt_global_no sequence
	query := `
		WITH receipt_sequence AS (
			SELECT 
				receipt_global_no,
				LAG(receipt_global_no) OVER (ORDER BY receipt_global_no) as prev_no
			FROM receipts
			WHERE device_id = $1 AND fiscal_day_id = $2
			ORDER BY receipt_global_no
		)
		SELECT prev_no + 1 as missing_no
		FROM receipt_sequence
		WHERE receipt_global_no - prev_no > 1`

	var missing []int
	err := r.db.Select(&missing, query, deviceID, fiscalDayID)
	return missing, err
}

func (r *receiptRepository) GetReceiptsWithValidationErrors(fiscalDayID int64) ([]models.Receipt, error) {
	var receipts []models.Receipt
	query := `
		SELECT * FROM receipts
		WHERE fiscal_day_id = $1
		  AND (validation_color = 'Red' OR validation_color = 'Grey')
		ORDER BY receipt_global_no`

	err := r.db.Select(&receipts, query, fiscalDayID)
	return receipts, err
}

func (r *receiptRepository) GetCreditDebitNotes(originalReceiptID int64) ([]*models.Receipt, []*models.Receipt, error) {
	var creditNotes []*models.Receipt
	var debitNotes []*models.Receipt

	// Get credit notes
	creditQuery := `
		SELECT * FROM receipts
		WHERE receipt_type = $1
		  AND credit_debit_note->>'receiptID' = $2
		ORDER BY receipt_date`

	err := r.db.Select(&creditNotes, creditQuery, models.ReceiptTypeCreditNote, fmt.Sprintf("%d", originalReceiptID))
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}

	// Get debit notes
	debitQuery := `
		SELECT * FROM receipts
		WHERE receipt_type = $1
		  AND credit_debit_note->>'receiptID' = $2
		ORDER BY receipt_date`

	err = r.db.Select(&debitNotes, debitQuery, models.ReceiptTypeDebitNote, fmt.Sprintf("%d", originalReceiptID))
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}

	return creditNotes, debitNotes, nil
}

// Helper method to load all related data for a receipt
func (r *receiptRepository) loadReceiptRelations(receipt *models.Receipt) error {
	// Load lines
	lines, err := r.GetReceiptLines(receipt.ID)
	if err != nil {
		return err
	}
	receipt.ReceiptLines = lines

	// Load taxes
	taxes, err := r.GetReceiptTaxes(receipt.ID)
	if err != nil {
		return err
	}
	receipt.ReceiptTaxes = taxes

	// Load payments
	payments, err := r.GetReceiptPayments(receipt.ID)
	if err != nil {
		return err
	}
	receipt.ReceiptPayments = payments

	return nil
}
