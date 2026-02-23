-- migrations/000001_initial_schema.up.sql
-- Create taxpayers table
CREATE TABLE IF NOT EXISTS taxpayers (
    id BIGSERIAL PRIMARY KEY,
    tin VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(250) NOT NULL,
    vat_number VARCHAR(9) UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'Active',
    taxpayer_day_max_hrs INTEGER NOT NULL DEFAULT 24,
    taxpayer_day_end_notification_hrs INTEGER NOT NULL DEFAULT 2,
    qr_url VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_taxpayers_tin ON taxpayers(tin);
CREATE INDEX idx_taxpayers_status ON taxpayers(status);

-- Create devices table
CREATE TABLE IF NOT EXISTS devices (
    id BIGSERIAL PRIMARY KEY,
    device_id INTEGER UNIQUE NOT NULL,
    taxpayer_id BIGINT NOT NULL REFERENCES taxpayers(id),
    device_serial_no VARCHAR(20) NOT NULL,
    device_model_name VARCHAR(100) NOT NULL,
    device_model_version VARCHAR(50) NOT NULL,
    activation_key VARCHAR(8) NOT NULL,
    certificate TEXT,
    certificate_thumbprint BYTEA,
    certificate_valid_till TIMESTAMP,
    operating_mode INTEGER NOT NULL DEFAULT 0, -- 0=Online, 1=Offline
    status VARCHAR(20) NOT NULL DEFAULT 'Active',
    branch_name VARCHAR(250) NOT NULL,
    branch_address JSONB NOT NULL,
    branch_contacts JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_devices_device_id ON devices(device_id);
CREATE INDEX idx_devices_taxpayer_id ON devices(taxpayer_id);
CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_thumbprint ON devices(certificate_thumbprint);

-- Create fiscal_days table
CREATE TABLE IF NOT EXISTS fiscal_days (
    id BIGSERIAL PRIMARY KEY,
    device_id INTEGER NOT NULL REFERENCES devices(device_id),
    fiscal_day_no INTEGER NOT NULL,
    fiscal_day_opened TIMESTAMP NOT NULL,
    fiscal_day_closed TIMESTAMP,
    status INTEGER NOT NULL DEFAULT 1, -- 0=Closed, 1=Opened, 2=CloseInitiated, 3=CloseFailed
    reconciliation_mode INTEGER,
    fiscal_day_device_signature JSONB,
    fiscal_day_server_signature JSONB,
    closing_error_code INTEGER,
    last_receipt_global_no INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, fiscal_day_no)
);

CREATE INDEX idx_fiscal_days_device_id ON fiscal_days(device_id);
CREATE INDEX idx_fiscal_days_status ON fiscal_days(status);
CREATE INDEX idx_fiscal_days_opened ON fiscal_days(fiscal_day_opened);

-- Create fiscal_counters table
CREATE TABLE IF NOT EXISTS fiscal_counters (
    id BIGSERIAL PRIMARY KEY,
    fiscal_day_id BIGINT NOT NULL REFERENCES fiscal_days(id) ON DELETE CASCADE,
    fiscal_counter_type INTEGER NOT NULL,
    fiscal_counter_currency VARCHAR(3) NOT NULL,
    fiscal_counter_tax_id INTEGER,
    fiscal_counter_tax_percent DECIMAL(5,2),
    fiscal_counter_money_type INTEGER,
    fiscal_counter_value DECIMAL(19,2) NOT NULL
);

CREATE INDEX idx_fiscal_counters_fiscal_day_id ON fiscal_counters(fiscal_day_id);

-- Create receipts table
CREATE TABLE IF NOT EXISTS receipts (
    id BIGSERIAL PRIMARY KEY,
    receipt_id BIGSERIAL UNIQUE NOT NULL,
    device_id INTEGER NOT NULL REFERENCES devices(device_id),
    fiscal_day_id BIGINT NOT NULL REFERENCES fiscal_days(id),
    receipt_type INTEGER NOT NULL, -- 0=FiscalInvoice, 1=CreditNote, 2=DebitNote
    receipt_currency VARCHAR(3) NOT NULL,
    receipt_counter INTEGER NOT NULL,
    receipt_global_no INTEGER NOT NULL,
    invoice_no VARCHAR(50) NOT NULL,
    buyer_data JSONB,
    receipt_notes TEXT,
    receipt_date TIMESTAMP NOT NULL,
    credit_debit_note JSONB,
    receipt_lines_tax_inclusive BOOLEAN NOT NULL,
    receipt_total DECIMAL(21,2) NOT NULL,
    receipt_print_form INTEGER DEFAULT 0,
    receipt_device_signature JSONB NOT NULL,
    receipt_server_signature JSONB,
    receipt_hash BYTEA,
    username VARCHAR(100),
    user_name_surname VARCHAR(250),
    validation_color VARCHAR(10),
    validation_errors TEXT[],
    server_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, receipt_global_no)
);

CREATE INDEX idx_receipts_device_id ON receipts(device_id);
CREATE INDEX idx_receipts_fiscal_day_id ON receipts(fiscal_day_id);
CREATE INDEX idx_receipts_receipt_id ON receipts(receipt_id);
CREATE INDEX idx_receipts_invoice_no ON receipts(invoice_no);
CREATE INDEX idx_receipts_receipt_date ON receipts(receipt_date);
CREATE INDEX idx_receipts_validation_color ON receipts(validation_color);

-- Create receipt_lines table
CREATE TABLE IF NOT EXISTS receipt_lines (
    id BIGSERIAL PRIMARY KEY,
    receipt_id BIGINT NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
    receipt_line_type INTEGER NOT NULL,
    receipt_line_no INTEGER NOT NULL,
    receipt_line_hs_code VARCHAR(8),
    receipt_line_name VARCHAR(200) NOT NULL,
    receipt_line_price DECIMAL(25,6),
    receipt_line_quantity DECIMAL(25,6) NOT NULL,
    receipt_line_total DECIMAL(21,2) NOT NULL,
    tax_code VARCHAR(3),
    tax_percent DECIMAL(5,2),
    tax_id INTEGER NOT NULL
);

CREATE INDEX idx_receipt_lines_receipt_id ON receipt_lines(receipt_id);

-- Create receipt_taxes table
CREATE TABLE IF NOT EXISTS receipt_taxes (
    id BIGSERIAL PRIMARY KEY,
    receipt_id BIGINT NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
    tax_code VARCHAR(3),
    tax_percent DECIMAL(5,2),
    tax_id INTEGER NOT NULL,
    tax_amount DECIMAL(21,2) NOT NULL,
    sales_amount_with_tax DECIMAL(21,2) NOT NULL
);

CREATE INDEX idx_receipt_taxes_receipt_id ON receipt_taxes(receipt_id);

-- Create receipt_payments table
CREATE TABLE IF NOT EXISTS receipt_payments (
    id BIGSERIAL PRIMARY KEY,
    receipt_id BIGINT NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
    money_type_code INTEGER NOT NULL,
    payment_amount DECIMAL(21,2) NOT NULL
);

CREATE INDEX idx_receipt_payments_receipt_id ON receipt_payments(receipt_id);

-- Create taxes table
CREATE TABLE IF NOT EXISTS taxes (
    id SERIAL PRIMARY KEY,
    tax_id INTEGER UNIQUE NOT NULL,
    tax_name VARCHAR(50) NOT NULL,
    tax_percent DECIMAL(5,2),
    tax_valid_from DATE NOT NULL,
    tax_valid_till DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_taxes_tax_id ON taxes(tax_id);
CREATE INDEX idx_taxes_valid_from ON taxes(tax_valid_from);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    taxpayer_id BIGINT NOT NULL REFERENCES taxpayers(id),
    username VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    person_name VARCHAR(100) NOT NULL,
    person_surname VARCHAR(100) NOT NULL,
    user_role VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    phone_no VARCHAR(20) NOT NULL,
    status INTEGER NOT NULL DEFAULT 2, -- 0=Active, 1=Blocked, 2=NotConfirmed
    security_code VARCHAR(10),
    security_code_expiry TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(taxpayer_id, username)
);

CREATE INDEX idx_users_taxpayer_id ON users(taxpayer_id);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- Create file_uploads table
CREATE TABLE IF NOT EXISTS file_uploads (
    id BIGSERIAL PRIMARY KEY,
    operation_id VARCHAR(60) UNIQUE NOT NULL,
    device_id INTEGER NOT NULL REFERENCES devices(device_id),
    file_name VARCHAR(100) NOT NULL,
    file_upload_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    file_processing_date TIMESTAMP,
    file_processing_status INTEGER NOT NULL DEFAULT 0, -- 0=InProgress, 1=Successful, 2=WithErrors, 3=WaitingForPrevious
    file_processing_error_codes INTEGER[],
    fiscal_day_no INTEGER NOT NULL,
    fiscal_day_opened_at TIMESTAMP NOT NULL,
    file_sequence INTEGER NOT NULL,
    ip_address VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_file_uploads_device_id ON file_uploads(device_id);
CREATE INDEX idx_file_uploads_operation_id ON file_uploads(operation_id);
CREATE INDEX idx_file_uploads_upload_date ON file_uploads(file_upload_date);

-- Create stock table
CREATE TABLE IF NOT EXISTS stock (
    id BIGSERIAL PRIMARY KEY,
    taxpayer_id BIGINT NOT NULL REFERENCES taxpayers(id),
    branch_id BIGINT REFERENCES devices(id),
    hs_code VARCHAR(8) NOT NULL,
    good_name VARCHAR(200) NOT NULL,
    quantity DECIMAL(19,3) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_stock_taxpayer_id ON stock(taxpayer_id);
CREATE INDEX idx_stock_branch_id ON stock(branch_id);
CREATE INDEX idx_stock_hs_code ON stock(hs_code);
CREATE INDEX idx_stock_good_name ON stock(good_name);

-- Create certificates_history table for certificate versioning
CREATE TABLE IF NOT EXISTS certificates_history (
    id BIGSERIAL PRIMARY KEY,
    device_id INTEGER NOT NULL REFERENCES devices(device_id),
    certificate TEXT NOT NULL,
    certificate_thumbprint BYTEA NOT NULL,
    issued_at TIMESTAMP NOT NULL,
    valid_till TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_certificates_history_device_id ON certificates_history(device_id);
CREATE INDEX idx_certificates_history_thumbprint ON certificates_history(certificate_thumbprint);

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT,
    action VARCHAR(50) NOT NULL,
    user_id BIGINT,
    device_id INTEGER,
    ip_address VARCHAR(100),
    details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_entity_type ON audit_logs(entity_type);
CREATE INDEX idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at
CREATE TRIGGER update_taxpayers_updated_at BEFORE UPDATE ON taxpayers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_devices_updated_at BEFORE UPDATE ON devices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_fiscal_days_updated_at BEFORE UPDATE ON fiscal_days
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_receipts_updated_at BEFORE UPDATE ON receipts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_file_uploads_updated_at BEFORE UPDATE ON file_uploads
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_stock_updated_at BEFORE UPDATE ON stock
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
