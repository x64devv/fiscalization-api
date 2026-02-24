package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"fiscalization-api/internal/models"
	"fiscalization-api/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type AdminService struct {
	adminRepo repository.AdminRepository
	jwtSecret string
	logger    *zap.Logger
}

func (s *AdminService) audit(entityType, action string, entityID *int64, deviceID *int, details string) {
	if err := s.adminRepo.InsertAuditLog(entityType, action, entityID, deviceID, "system", details); err != nil {
		s.logger.Warn("Failed to write audit log", zap.Error(err))
	}
}

func NewAdminService(adminRepo repository.AdminRepository, jwtSecret string, logger *zap.Logger) *AdminService {
	return &AdminService{adminRepo: adminRepo, jwtSecret: jwtSecret, logger: logger}
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

// Login validates hardcoded admin credentials (replace with DB-backed admin users in production)
func (s *AdminService) Login(req models.AdminLoginRequest) (*models.AdminLoginResponse, error) {
	// In production: look up admin user table, verify bcrypt hash
	// For now: env-based superadmin account
	if req.Username != "superadmin" {
		return nil, models.NewAPIError(401, "Invalid credentials", "ADM01")
	}
	// TODO: replace hardcoded password with bcrypt comparison from admin_users table
	if req.Password != "ZimraAdmin2024!" {
		return nil, models.NewAPIError(401, "Invalid credentials", "ADM01")
	}

	expiresAt := time.Now().Add(8 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  req.Username,
		"role": "superadmin",
		"exp":  expiresAt.Unix(),
		"iat":  time.Now().Unix(),
	})

	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &models.AdminLoginResponse{
		Token:     signed,
		ExpiresAt: expiresAt,
		Username:  req.Username,
		Role:      "superadmin",
	}, nil
}

// ─── Taxpayer (Company) ───────────────────────────────────────────────────────

func (s *AdminService) CreateTaxpayer(req models.CreateTaxpayerRequest) (*models.Taxpayer, error) {
	// Check TIN uniqueness
	existing, err := s.adminRepo.GetTaxpayerByTIN(req.TIN)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, models.NewAPIError(422, fmt.Sprintf("Taxpayer with TIN %s already exists", req.TIN), "ADM10")
	}

	if req.Status == "" {
		req.Status = "Active"
	}
	if req.TaxPayerDayMaxHrs == 0 {
		req.TaxPayerDayMaxHrs = 24
	}
	if req.TaxpayerDayEndNotificationHrs == 0 {
		req.TaxpayerDayEndNotificationHrs = 2
	}
	if req.QrURL == "" {
		req.QrURL = "https://receipt.zimra.co.zw"
	}

	tp := &models.Taxpayer{
		TIN:                          req.TIN,
		Name:                         req.Name,
		VATNumber:                    req.VATNumber,
		Status:                       req.Status,
		TaxPayerDayMaxHrs:            req.TaxPayerDayMaxHrs,
		TaxpayerDayEndNotificationHrs: req.TaxpayerDayEndNotificationHrs,
		QrURL:                        req.QrURL,
	}

	if err := s.adminRepo.CreateTaxpayer(tp); err != nil {
		return nil, err
	}

	s.logger.Info("Taxpayer created", zap.String("tin", tp.TIN), zap.Int64("id", tp.ID))
	return tp, nil
}

func (s *AdminService) ListTaxpayers(offset, limit int, search string) (*models.ListTaxpayersResponse, error) {
	total, rows, err := s.adminRepo.ListTaxpayers(offset, limit, search)
	if err != nil {
		return nil, err
	}
	return &models.ListTaxpayersResponse{Total: total, Rows: rows}, nil
}

func (s *AdminService) GetTaxpayer(id int64) (*models.Taxpayer, error) {
	tp, err := s.adminRepo.GetTaxpayerByID(id)
	if err != nil {
		return nil, err
	}
	if tp == nil {
		return nil, models.NewAPIError(404, "Taxpayer not found", "ADM11")
	}
	return tp, nil
}

func (s *AdminService) UpdateTaxpayer(req models.UpdateTaxpayerRequest) (*models.Taxpayer, error) {
	tp, err := s.adminRepo.GetTaxpayerByID(req.ID)
	if err != nil || tp == nil {
		return nil, models.NewAPIError(404, "Taxpayer not found", "ADM11")
	}
	tp.Name = req.Name
	tp.VATNumber = req.VATNumber
	tp.Status = req.Status
	tp.TaxPayerDayMaxHrs = req.TaxPayerDayMaxHrs
	tp.TaxpayerDayEndNotificationHrs = req.TaxpayerDayEndNotificationHrs
	tp.QrURL = req.QrURL

	if err := s.adminRepo.UpdateTaxpayer(tp); err != nil {
		return nil, err
	}
	return tp, nil
}

func (s *AdminService) SetTaxpayerStatus(id int64, status string) error {
	return s.adminRepo.SetTaxpayerStatus(id, status)
}

// ─── Device ───────────────────────────────────────────────────────────────────

// generateActivationKey creates a cryptographically random 8-char uppercase activation key
func generateActivationKey() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:8], nil
}

func (s *AdminService) CreateDevice(req models.AdminCreateDeviceRequest) (*models.Device, error) {
	// Verify taxpayer exists
	tp, err := s.adminRepo.GetTaxpayerByID(req.TaxpayerID)
	if err != nil || tp == nil {
		return nil, models.NewAPIError(422, "Taxpayer not found", "ADM11")
	}

	// Use provided activation key or generate one
	activationKey := req.ActivationKey
	if activationKey == "" {
		var genErr error
		activationKey, genErr = generateActivationKey()
		if genErr != nil {
			return nil, genErr
		}
	}

	device := &models.Device{
		DeviceID:           req.DeviceID,
		TaxpayerID:         req.TaxpayerID,
		DeviceSerialNo:     req.DeviceSerialNo,
		DeviceModelName:    req.DeviceModelName,
		DeviceModelVersion: req.DeviceModelVersion,
		ActivationKey:      activationKey,
		OperatingMode:      req.OperatingMode,
		Status:             "Active",
		BranchName:         req.BranchName,
		BranchAddress:      req.BranchAddress,
		BranchContacts:     req.BranchContacts,
	}

	if err := s.adminRepo.CreateDevice(device); err != nil {
		return nil, err
	}

	// Store activation key in response (only time it's shown)
	device.ActivationKey = activationKey
	s.logger.Info("Device provisioned",
		zap.Int("deviceID", device.DeviceID),
		zap.Int64("taxpayerID", req.TaxpayerID),
	)
	return device, nil
}

func (s *AdminService) ListDevicesByTaxpayer(taxpayerID int64) ([]models.Device, error) {
	return s.adminRepo.ListDevicesByTaxpayer(taxpayerID)
}

func (s *AdminService) ListAllDevices(offset, limit int) (*models.ListDevicesResponse, error) {
	total, rows, err := s.adminRepo.ListAllDevices(offset, limit)
	if err != nil {
		return nil, err
	}
	return &models.ListDevicesResponse{Total: total, Rows: rows}, nil
}

func (s *AdminService) UpdateDeviceStatus(deviceID int, status string) error {
	return s.adminRepo.UpdateDeviceStatus(deviceID, status)
}

func (s *AdminService) UpdateDeviceMode(deviceID int, mode int) error {
	return s.adminRepo.UpdateDeviceMode(deviceID, mode)
}

// ─── Overview Queries ─────────────────────────────────────────────────────────

func (s *AdminService) ListFiscalDays(taxpayerID *int64, deviceID *int, offset, limit int) (*models.ListFiscalDaysResponse, error) {
	total, rows, err := s.adminRepo.ListFiscalDays(taxpayerID, deviceID, offset, limit)
	if err != nil {
		return nil, err
	}
	return &models.ListFiscalDaysResponse{Total: total, Rows: rows}, nil
}

func (s *AdminService) ListReceipts(taxpayerID *int64, deviceID *int, from, to *time.Time, offset, limit int) (*models.ListReceiptsResponse, error) {
	total, rows, err := s.adminRepo.ListReceipts(taxpayerID, deviceID, from, to, offset, limit)
	if err != nil {
		return nil, err
	}
	return &models.ListReceiptsResponse{Total: total, Rows: rows}, nil
}

func (s *AdminService) GetSystemStats() (*models.SystemStats, error) {
	return s.adminRepo.GetSystemStats()
}

func (s *AdminService) ListAuditLogs(entityType string, entityID *int64, offset, limit int) (*models.ListAuditLogsResponse, error) {
	total, rows, err := s.adminRepo.ListAuditLogs(entityType, entityID, offset, limit)
	if err != nil {
		return nil, err
	}
	return &models.ListAuditLogsResponse{Total: total, Rows: rows}, nil
}


