package service

import (
	"strings"
	"time"

	"fiscalization-api/internal/models"
	"fiscalization-api/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DeviceService struct {
	deviceRepo  repository.DeviceRepository
	cryptoSvc   *CryptoService
	logger      *zap.Logger
}

func NewDeviceService(
	deviceRepo repository.DeviceRepository,
	cryptoSvc *CryptoService,
	logger *zap.Logger,
) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		cryptoSvc:  cryptoSvc,
		logger:     logger,
	}
}

// VerifyTaxpayer verifies taxpayer information before device registration
func (s *DeviceService) VerifyTaxpayer(
	req models.VerifyTaxpayerRequest,
	modelName, modelVersion string,
) (*models.VerifyTaxpayerResponse, error) {
	// Check if device model is blacklisted
	blacklisted, err := s.deviceRepo.IsBlacklisted(modelName, modelVersion)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, models.NewAPIError(422, "Device model is blacklisted", models.ErrCodeDEV04)
	}

	// Get device by ID
	device, err := s.deviceRepo.GetByDeviceID(req.DeviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, models.NewAPIError(422, "Device not found", models.ErrCodeDEV01)
	}

	// Verify activation key (case-insensitive)
	if !strings.EqualFold(device.ActivationKey, req.ActivationKey) {
		return nil, models.NewAPIError(422, "Activation key is incorrect", models.ErrCodeDEV02)
	}

	// Verify device serial number matches
	if device.DeviceSerialNo != req.DeviceSerialNo {
		return nil, models.NewAPIError(422, "Device serial number mismatch", models.ErrCodeDEV01)
	}

	// Get taxpayer information
	taxpayer, err := s.deviceRepo.GetTaxpayer(device.TaxpayerID)
	if err != nil {
		return nil, err
	}
	if taxpayer == nil {
		return nil, models.NewAPIError(422, "Taxpayer not found", models.ErrCodeDEV05)
	}

	// Check taxpayer status
	if taxpayer.Status != "Active" {
		return nil, models.NewAPIError(422, "Taxpayer is not active", models.ErrCodeDEV05)
	}

	// Build response
	resp := &models.VerifyTaxpayerResponse{
		OperationID:          generateOperationID(),
		TaxPayerName:         taxpayer.Name,
		TaxPayerTIN:          taxpayer.TIN,
		DeviceBranchName:     device.BranchName,
		DeviceBranchAddress:  device.BranchAddress,
		DeviceBranchContacts: device.BranchContacts,
	}

	if taxpayer.VATNumber != nil {
		resp.VATNumber = *taxpayer.VATNumber
	}

	return resp, nil
}

// RegisterDevice registers a new device and issues certificate
func (s *DeviceService) RegisterDevice(
	req models.DeviceRegistrationRequest,
	modelName, modelVersion string,
) (*models.DeviceRegistrationResponse, error) {
	// Check if device model is blacklisted
	blacklisted, err := s.deviceRepo.IsBlacklisted(modelName, modelVersion)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, models.NewAPIError(422, "Device model is blacklisted", models.ErrCodeDEV04)
	}

	// Get device
	device, err := s.deviceRepo.GetByDeviceID(req.DeviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, models.NewAPIError(422, "Device not found", models.ErrCodeDEV01)
	}

	// Verify activation key
	if !strings.EqualFold(device.ActivationKey, req.ActivationKey) {
		return nil, models.NewAPIError(422, "Activation key is incorrect", models.ErrCodeDEV02)
	}

	// Check device status
	if device.Status != "Active" {
		return nil, models.NewAPIError(422, "Device is not active", models.ErrCodeDEV01)
	}

	// Check taxpayer status
	taxpayer, err := s.deviceRepo.GetTaxpayer(device.TaxpayerID)
	if err != nil {
		return nil, err
	}
	if taxpayer.Status != "Active" {
		return nil, models.NewAPIError(422, "Taxpayer is not active", models.ErrCodeDEV05)
	}

	// Issue certificate
	certPEM, thumbprint, validTill, err := s.cryptoSvc.IssueCertificate(
		[]byte(req.CertificateRequest),
		req.DeviceID,
		device.DeviceSerialNo,
	)
	if err != nil {
		s.logger.Error("Failed to issue certificate", zap.Error(err))
		return nil, models.NewAPIError(422, "Certificate request is invalid", models.ErrCodeDEV03)
	}

	// Update device with certificate
	err = s.deviceRepo.UpdateCertificate(req.DeviceID, certPEM, thumbprint, validTill)
	if err != nil {
		return nil, err
	}

	// Save certificate history
	err = s.deviceRepo.SaveCertificateHistory(req.DeviceID, certPEM, thumbprint, validTill)
	if err != nil {
		s.logger.Warn("Failed to save certificate history", zap.Error(err))
	}

	return &models.DeviceRegistrationResponse{
		OperationID: generateOperationID(),
		Certificate: certPEM,
	}, nil
}

// IssueCertificate renews device certificate
func (s *DeviceService) IssueCertificate(
	req models.IssueCertificateRequest,
) (*models.IssueCertificateResponse, error) {
	// Get device
	device, err := s.deviceRepo.GetByDeviceID(req.DeviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, models.NewAPIError(422, "Device not found", models.ErrCodeDEV01)
	}

	// Check device status
	if device.Status != "Active" {
		return nil, models.NewAPIError(422, "Device is not active", models.ErrCodeDEV01)
	}

	// Issue new certificate
	certPEM, thumbprint, validTill, err := s.cryptoSvc.IssueCertificate(
		[]byte(req.CertificateRequest),
		req.DeviceID,
		device.DeviceSerialNo,
	)
	if err != nil {
		s.logger.Error("Failed to issue certificate", zap.Error(err))
		return nil, models.NewAPIError(422, "Certificate request is invalid", models.ErrCodeDEV03)
	}

	// Update device with new certificate
	err = s.deviceRepo.UpdateCertificate(req.DeviceID, certPEM, thumbprint, validTill)
	if err != nil {
		return nil, err
	}

	// Save certificate history
	err = s.deviceRepo.SaveCertificateHistory(req.DeviceID, certPEM, thumbprint, validTill)
	if err != nil {
		s.logger.Warn("Failed to save certificate history", zap.Error(err))
	}

	return &models.IssueCertificateResponse{
		OperationID: generateOperationID(),
		Certificate: certPEM,
	}, nil
}

// GetConfig retrieves device configuration
func (s *DeviceService) GetConfig(deviceID int) (*models.GetConfigResponse, error) {
	// Get device
	device, err := s.deviceRepo.GetByDeviceID(deviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, models.NewAPIError(422, "Device not found", models.ErrCodeDEV01)
	}

	// Get taxpayer
	taxpayer, err := s.deviceRepo.GetTaxpayer(device.TaxpayerID)
	if err != nil {
		return nil, err
	}

	// Get applicable taxes
	taxes, err := s.deviceRepo.GetApplicableTaxes()
	if err != nil {
		return nil, err
	}

	// Build response
	resp := &models.GetConfigResponse{
		OperationID:                   generateOperationID(),
		TaxPayerName:                  taxpayer.Name,
		TaxPayerTIN:                   taxpayer.TIN,
		DeviceSerialNo:                device.DeviceSerialNo,
		DeviceBranchName:              device.BranchName,
		DeviceBranchAddress:           device.BranchAddress,
		DeviceBranchContacts:          device.BranchContacts,
		DeviceOperatingMode:           device.OperatingMode,
		TaxPayerDayMaxHrs:             taxpayer.TaxPayerDayMaxHrs,
		TaxpayerDayEndNotificationHrs: taxpayer.TaxpayerDayEndNotificationHrs,
		ApplicableTaxes:               taxes,
		CertificateValidTill:          device.CertificateValidTill,
		QrURL:                         taxpayer.QrURL,
	}

	if taxpayer.VATNumber != nil {
		resp.VATNumber = *taxpayer.VATNumber
	}

	return resp, nil
}

// GetStatus retrieves device and fiscal day status
func (s *DeviceService) GetStatus(deviceID int) (*models.GetStatusResponse, error) {
	// Validate operating mode
	device, err := s.deviceRepo.GetByDeviceID(deviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, models.NewAPIError(422, "Device not found", models.ErrCodeDEV01)
	}

	if device.OperatingMode == models.DeviceOperatingModeOffline {
		return nil, models.NewAPIError(422, "Device operating mode is Offline", models.ErrCodeDEV01)
	}

	// Get current fiscal day
	fiscalDay, err := s.deviceRepo.GetCurrentFiscalDay(deviceID)
	if err != nil {
		return nil, err
	}

	resp := &models.GetStatusResponse{
		OperationID: generateOperationID(),
	}

	if fiscalDay == nil {
		// No fiscal day yet
		resp.FiscalDayStatus = models.FiscalDayStatusClosed.String()
		return resp, nil
	}

	// Convert enum to string
	resp.FiscalDayStatus = fiscalDay.Status.String()
	resp.FiscalDayNo = &fiscalDay.FiscalDayNo
	
	// Convert reconciliation mode enum to string pointer
	if fiscalDay.ReconciliationMode != nil {
		mode := fiscalDay.ReconciliationMode.String()
		resp.FiscalDayReconciliationMode = &mode
	}
	
	resp.FiscalDayServerSignature = fiscalDay.FiscalDayServerSignature
	resp.FiscalDayClosed = fiscalDay.FiscalDayClosed
	resp.LastReceiptGlobalNo = fiscalDay.LastReceiptGlobalNo

	if fiscalDay.Status == models.FiscalDayStatusClosed {
		// Get counters and document quantities if manually closed
		if fiscalDay.ReconciliationMode != nil && *fiscalDay.ReconciliationMode == models.FiscalDayReconciliationModeManual {
			counters, err := s.deviceRepo.GetFiscalDayCounters(fiscalDay.ID)
			if err != nil {
				s.logger.Warn("Failed to get fiscal day counters", zap.Error(err))
			} else {
				resp.FiscalDayCounters = counters
			}

			docQuantities, err := s.deviceRepo.GetFiscalDayDocumentQuantities(fiscalDay.ID)
			if err != nil {
				s.logger.Warn("Failed to get document quantities", zap.Error(err))
			} else {
				resp.FiscalDayDocumentQuantities = docQuantities
			}
		}
	}

	return resp, nil
}

// Ping handles device heartbeat
func (s *DeviceService) Ping(deviceID int) (*models.PingResponse, error) {
	// Update last ping time
	err := s.deviceRepo.UpdateLastPing(deviceID, time.Now())
	if err != nil {
		return nil, err
	}

	return &models.PingResponse{
		OperationID:        generateOperationID(),
		ReportingFrequency: 5, // 5 minutes default
	}, nil
}

// GetStockList retrieves stock items
func (s *DeviceService) GetStockList(req models.GetStockListRequest) (*models.GetStockListResponse, error) {
	// Get device to validate taxpayer
	device, err := s.deviceRepo.GetByDeviceID(req.DeviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, models.NewAPIError(422, "Device not found", models.ErrCodeDEV01)
	}

	// Get stock items
	total, items, err := s.deviceRepo.GetStockList(
		device.TaxpayerID,
		device.ID,
		req.HSCode,
		req.GoodName,
		req.Sort,
		req.Order,
		req.Offset,
		req.Limit,
		req.Operator,
	)
	if err != nil {
		return nil, err
	}

	return &models.GetStockListResponse{
		Total: total,
		Rows:  items,
	}, nil
}

// GetServerCertificate returns FDMS server certificate
func (s *DeviceService) GetServerCertificate(thumbprint []byte) (*models.GetServerCertificateResponse, error) {
	chain, validTill, err := s.cryptoSvc.GetServerCertificate(thumbprint)
	if err != nil {
		return nil, err
	}

	return &models.GetServerCertificateResponse{
		Certificate:          chain,
		CertificateValidTill: validTill,
	}, nil
}

// Helper functions

func generateOperationID() string {
	return uuid.New().String()
}
