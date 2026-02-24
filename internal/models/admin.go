package models

import "time"

// ─── System Stats ─────────────────────────────────────────────────────────────

type SystemStats struct {
	TotalCompanies   int     `json:"totalCompanies" db:"total_companies"`
	ActiveCompanies  int     `json:"activeCompanies" db:"active_companies"`
	TotalDevices     int     `json:"totalDevices" db:"total_devices"`
	ActiveDevices    int     `json:"activeDevices" db:"active_devices"`
	TodayReceipts    int     `json:"todayReceipts" db:"today_receipts"`
	OpenFiscalDays   int     `json:"openFiscalDays" db:"open_fiscal_days"`
	TodayRevenue     float64 `json:"todayRevenue" db:"today_revenue"`
	ValidationErrors int     `json:"validationErrors" db:"validation_errors"`
}

// ─── Audit Log ────────────────────────────────────────────────────────────────

type AuditLog struct {
	ID         int64      `json:"id" db:"id"`
	EntityType string     `json:"entityType" db:"entity_type"`
	EntityID   *int64     `json:"entityId,omitempty" db:"entity_id"`
	Action     string     `json:"action" db:"action"`
	UserID     *int64     `json:"userId,omitempty" db:"user_id"`
	DeviceID   *int       `json:"deviceId,omitempty" db:"device_id"`
	IPAddress  string     `json:"ipAddress" db:"ip_address"`
	Details    *string    `json:"details,omitempty" db:"details"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
}

// ─── Admin Taxpayer Requests ──────────────────────────────────────────────────

type CreateTaxpayerRequest struct {
	TIN                          string  `json:"tin" binding:"required,len=10"`
	Name                         string  `json:"name" binding:"required,max=250"`
	VATNumber                    *string `json:"vatNumber,omitempty"`
	Status                       string  `json:"status"`
	TaxPayerDayMaxHrs            int     `json:"taxPayerDayMaxHrs"`
	TaxpayerDayEndNotificationHrs int    `json:"taxpayerDayEndNotificationHrs"`
	QrURL                        string  `json:"qrUrl"`
}

type UpdateTaxpayerRequest struct {
	ID                           int64   `json:"id" binding:"required"`
	Name                         string  `json:"name" binding:"required,max=250"`
	VATNumber                    *string `json:"vatNumber,omitempty"`
	Status                       string  `json:"status" binding:"required"`
	TaxPayerDayMaxHrs            int     `json:"taxPayerDayMaxHrs"`
	TaxpayerDayEndNotificationHrs int    `json:"taxpayerDayEndNotificationHrs"`
	QrURL                        string  `json:"qrUrl"`
}

// ─── Admin Device Requests ────────────────────────────────────────────────────

type AdminCreateDeviceRequest struct {
	DeviceID           int                 `json:"deviceID" binding:"required"`
	TaxpayerID         int64               `json:"taxpayerID" binding:"required"`
	DeviceSerialNo     string              `json:"deviceSerialNo" binding:"required,max=20"`
	DeviceModelName    string              `json:"deviceModelName" binding:"required,max=100"`
	DeviceModelVersion string              `json:"deviceModelVersion" binding:"required,max=50"`
	ActivationKey      string              `json:"activationKey" binding:"required,len=8"`
	OperatingMode      DeviceOperatingMode `json:"operatingMode"`
	BranchName         string              `json:"branchName" binding:"required,max=250"`
	BranchAddress      Address             `json:"branchAddress" binding:"required"`
	BranchContacts     *Contacts           `json:"branchContacts,omitempty"`
}

type UpdateDeviceStatusRequest struct {
	DeviceID int    `json:"deviceID" binding:"required"`
	Status   string `json:"status" binding:"required,oneof=Active Blocked Revoked"`
}

type UpdateDeviceModeRequest struct {
	DeviceID int `json:"deviceID" binding:"required"`
	Mode     int `json:"mode" binding:"required,oneof=0 1"`
}

// ─── Admin List Responses ─────────────────────────────────────────────────────

type ListTaxpayersResponse struct {
	Total int         `json:"total"`
	Rows  []Taxpayer  `json:"rows"`
}

type ListDevicesResponse struct {
	Total int      `json:"total"`
	Rows  []Device `json:"rows"`
}

type ListFiscalDaysResponse struct {
	Total int          `json:"total"`
	Rows  []FiscalDay  `json:"rows"`
}

// AdminReceiptRow is a flat receipt DTO for the admin cross-tenant view.
// Uses its own struct so ValidationColor is exposed in JSON (models.Receipt hides it with json:"-").
type AdminReceiptRow struct {
	ID              int64     `json:"id" db:"id"`
	ReceiptID       int64     `json:"receiptID" db:"receipt_id"`
	DeviceID        int       `json:"deviceID" db:"device_id"`
	ReceiptType     int       `json:"receiptType" db:"receipt_type"`
	ReceiptCurrency string    `json:"receiptCurrency" db:"receipt_currency"`
	InvoiceNo       string    `json:"invoiceNo" db:"invoice_no"`
	ReceiptDate     time.Time `json:"receiptDate" db:"receipt_date"`
	ReceiptTotal    float64   `json:"receiptTotal" db:"receipt_total"`
	ValidationColor *string   `json:"validationColor,omitempty" db:"validation_color"`
	ServerDate      *time.Time `json:"serverDate,omitempty" db:"server_date"`
}

type ListReceiptsResponse struct {
	Total int               `json:"total"`
	Rows  []AdminReceiptRow `json:"rows"`
}

type ListAuditLogsResponse struct {
	Total int        `json:"total"`
	Rows  []AuditLog `json:"rows"`
}

// ─── Admin Auth ───────────────────────────────────────────────────────────────

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminLoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
}

type AdminCreateUserRequest struct {
	TaxpayerID    int64  `json:"taxpayerID"`
	Username      string `json:"username" binding:"required,max=100"`
	Password      string `json:"password" binding:"required,max=100"`
	PersonName    string `json:"personName" binding:"required,max=100"`
	PersonSurname string `json:"personSurname" binding:"required,max=100"`
	UserRole      string `json:"userRole" binding:"required,max=100"`
	Email         string `json:"email" binding:"required,max=100"`
	PhoneNo       string `json:"phoneNo" binding:"required,max=20"`
}

type AdminUserRow struct {
	ID            int64     `json:"id" db:"id"`
	Username      string    `json:"username" db:"username"`
	PersonName    string    `json:"personName" db:"person_name"`
	PersonSurname string    `json:"personSurname" db:"person_surname"`
	UserRole      string    `json:"userRole" db:"user_role"`
	Email         string    `json:"email" db:"email"`
	PhoneNo       string    `json:"phoneNo" db:"phone_no"`
	Status        int       `json:"status" db:"status"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}