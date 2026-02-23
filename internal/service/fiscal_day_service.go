package service

import (
	"fmt"
	"time"

	"fiscalization-api/internal/models"
	"fiscalization-api/internal/repository"
	"fiscalization-api/internal/utils"

	"go.uber.org/zap"
)

type FiscalDayService struct {
	fiscalDayRepo repository.FiscalDayRepository
	receiptRepo   repository.ReceiptRepository
	deviceRepo    repository.DeviceRepository
	cryptoSvc     *CryptoService
	logger        *zap.Logger
}

func NewFiscalDayService(
	fiscalDayRepo repository.FiscalDayRepository,
	receiptRepo repository.ReceiptRepository,
	deviceRepo repository.DeviceRepository,
	cryptoSvc *CryptoService,
	logger *zap.Logger,
) *FiscalDayService {
	return &FiscalDayService{
		fiscalDayRepo: fiscalDayRepo,
		receiptRepo:   receiptRepo,
		deviceRepo:    deviceRepo,
		cryptoSvc:     cryptoSvc,
		logger:        logger,
	}
}

// OpenFiscalDay opens a new fiscal day
func (s *FiscalDayService) OpenFiscalDay(req models.OpenFiscalDayRequest) (*models.OpenFiscalDayResponse, error) {
	// Get device
	device, err := s.deviceRepo.GetByDeviceID(req.DeviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, models.NewAPIError(422, "Device not found", models.ErrCodeDEV01)
	}

	// Check operating mode
	if device.OperatingMode == models.DeviceOperatingModeOffline {
		return nil, models.NewAPIError(422, "Device operating mode is Offline", models.ErrCodeDEV01)
	}

	// Get current fiscal day
	currentDay, err := s.fiscalDayRepo.GetCurrent(req.DeviceID)
	if err != nil {
		return nil, err
	}

	// Check if a day is already open
	if currentDay != nil && currentDay.Status == models.FiscalDayStatusOpened {
		return nil, models.NewAPIError(422, "Fiscal day is already opened", models.ErrCodeFISC01)
	}

	// Check if previous day is closed
	if currentDay != nil && currentDay.Status != models.FiscalDayStatusClosed {
		return nil, models.NewAPIError(422, "Previous fiscal day is not closed", models.ErrCodeFISC02)
	}

	// Determine next fiscal day number
	nextDayNo := 1
	if currentDay != nil {
		nextDayNo = currentDay.FiscalDayNo + 1
	}

	// Create new fiscal day
	fiscalDay := &models.FiscalDay{
		DeviceID:        req.DeviceID,
		FiscalDayNo:     nextDayNo,
		FiscalDayOpened: time.Now(),
		Status:          models.FiscalDayStatusOpened,
	}

	if err := s.fiscalDayRepo.Create(fiscalDay); err != nil {
		s.logger.Error("Failed to create fiscal day", zap.Error(err))
		return nil, fmt.Errorf("failed to create fiscal day: %w", err)
	}

	s.logger.Info("Fiscal day opened",
		zap.Int("deviceID", req.DeviceID),
		zap.Int("fiscalDayNo", nextDayNo),
	)

	return &models.OpenFiscalDayResponse{
		OperationID: generateOperationID(),
		FiscalDayNo: nextDayNo,
	}, nil
}

// CloseFiscalDay closes the current fiscal day
func (s *FiscalDayService) CloseFiscalDay(req models.CloseFiscalDayRequest) (*models.CloseFiscalDayResponse, error) {
	// Get device
	device, err := s.deviceRepo.GetByDeviceID(req.DeviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, models.NewAPIError(422, "Device not found", models.ErrCodeDEV01)
	}

	// Check operating mode
	if device.OperatingMode == models.DeviceOperatingModeOffline {
		return nil, models.NewAPIError(422, "Device operating mode is Offline", models.ErrCodeDEV01)
	}

	// Get current fiscal day
	fiscalDay, err := s.fiscalDayRepo.GetCurrent(req.DeviceID)
	if err != nil {
		return nil, err
	}
	if fiscalDay == nil {
		return nil, models.NewAPIError(422, "No fiscal day to close", models.ErrCodeFISC03)
	}

	// Check if day is opened or close failed
	if fiscalDay.Status != models.FiscalDayStatusOpened && fiscalDay.Status != models.FiscalDayStatusCloseFailed {
		return nil, models.NewAPIError(422, "Fiscal day cannot be closed", models.ErrCodeFISC03)
	}

	// Check for validation errors that block closing
	receiptsWithErrors, err := s.receiptRepo.GetReceiptsWithValidationErrors(fiscalDay.ID)
	if err != nil {
		return nil, err
	}

	hasBlockingErrors := false
	for _, receipt := range receiptsWithErrors {
		if receipt.ValidationColor != nil && *receipt.ValidationColor == models.ValidationColorRed {
			hasBlockingErrors = true
			break
		}
		if receipt.ValidationColor != nil && *receipt.ValidationColor == models.ValidationColorGrey {
			hasBlockingErrors = true
			break
		}
	}

	if hasBlockingErrors {
		return nil, models.NewAPIError(422, "Fiscal day has receipts with validation errors", models.ErrCodeFISC04)
	}

	// Determine reconciliation mode
	reconciliationMode := models.FiscalDayReconciliationModeAuto
	if len(req.FiscalDayCounters) > 0 {
		reconciliationMode = models.FiscalDayReconciliationModeManual
	}

	// Validate counters if manual mode
	var counters []models.FiscalDayCounter
	if reconciliationMode == models.FiscalDayReconciliationModeManual {
		counters = req.FiscalDayCounters

		// Validate submitted counters match actual
		valid, err := s.fiscalDayRepo.ValidateCounters(fiscalDay.ID, counters)
		if err != nil {
			s.logger.Error("Failed to validate counters", zap.Error(err))
			return nil, fmt.Errorf("failed to validate counters: %w", err)
		}

		if !valid {
			return nil, models.NewAPIError(422, "Submitted counters do not match actual values", models.ErrCodeFISC04)
		}
	} else {
		// Calculate counters automatically
		counters, err = s.fiscalDayRepo.GetCounters(fiscalDay.ID)
		if err != nil {
			return nil, err
		}
	}

	// Generate fiscal day hash
	fiscalDayDate := fiscalDay.FiscalDayOpened.Format("2006-01-02")
	_, err = utils.GenerateFiscalDayHash(
		req.DeviceID,
		fiscalDay.FiscalDayNo,
		fiscalDayDate,
		counters,
	)
	if err != nil {
		s.logger.Error("Failed to generate fiscal day hash", zap.Error(err))
		return nil, fmt.Errorf("failed to generate fiscal day hash: %w", err)
	}

	// Verify device signature if auto mode
	if reconciliationMode == models.FiscalDayReconciliationModeAuto {
		if req.FiscalDayDeviceSignature == nil {
			return nil, models.NewAPIError(422, "Device signature required for auto reconciliation", models.ErrCodeFISC04)
		}

		// In production, verify the signature here
		// For now, just store it
	}

	// Generate server signature
	closedAt := time.Now()
	serverSignature, err := s.generateFiscalDayServerSignature(
		req.DeviceID,
		fiscalDay.FiscalDayNo,
		fiscalDayDate,
		closedAt,
		reconciliationMode,
		counters,
		req.FiscalDayDeviceSignature,
	)
	if err != nil {
		s.logger.Error("Failed to generate server signature", zap.Error(err))
		return nil, fmt.Errorf("failed to generate server signature: %w", err)
	}

	// Update fiscal day
	fiscalDay.FiscalDayClosed = &closedAt
	fiscalDay.Status = models.FiscalDayStatusClosed
	fiscalDay.ReconciliationMode = &reconciliationMode
	fiscalDay.FiscalDayDeviceSignature = req.FiscalDayDeviceSignature
	fiscalDay.FiscalDayServerSignature = serverSignature

	if err := s.fiscalDayRepo.Update(fiscalDay); err != nil {
		s.logger.Error("Failed to update fiscal day", zap.Error(err))
		return nil, fmt.Errorf("failed to update fiscal day: %w", err)
	}

	// Save counters
	if err := s.fiscalDayRepo.CreateCounters(fiscalDay.ID, counters); err != nil {
		s.logger.Warn("Failed to save counters", zap.Error(err))
	}

	s.logger.Info("Fiscal day closed",
		zap.Int("deviceID", req.DeviceID),
		zap.Int("fiscalDayNo", fiscalDay.FiscalDayNo),
		zap.String("reconciliationMode", string(rune(reconciliationMode))),
	)

	resp := &models.CloseFiscalDayResponse{
		OperationID:                generateOperationID(),
		FiscalDayServerSignature:   *serverSignature,
		FiscalDayCounters:          counters,
		FiscalDayDocumentQuantities: make([]models.FiscalDayDocumentQuantity, 0),
	}

	// Get document quantities
	docQuantities, err := s.deviceRepo.GetFiscalDayDocumentQuantities(fiscalDay.ID)
	if err != nil {
		s.logger.Warn("Failed to get document quantities", zap.Error(err))
	} else {
		resp.FiscalDayDocumentQuantities = docQuantities
	}

	return resp, nil
}

// GetFiscalDayStatus gets the status of the current fiscal day
func (s *FiscalDayService) GetFiscalDayStatus(deviceID int) (*models.GetFiscalDayStatusResponse, error) {
	// Get current fiscal day
	fiscalDay, err := s.fiscalDayRepo.GetCurrent(deviceID)
	if err != nil {
		return nil, err
	}

	resp := &models.GetFiscalDayStatusResponse{
		OperationID: generateOperationID(),
	}

	if fiscalDay == nil {
		resp.FiscalDayStatus = models.FiscalDayStatusClosed.String()
		return resp, nil
	}

	mode := fiscalDay.ReconciliationMode.String()

	resp.FiscalDayStatus = fiscalDay.Status.String()
	resp.FiscalDayNo = &fiscalDay.FiscalDayNo
	resp.FiscalDayReconciliationMode = &mode
	resp.FiscalDayServerSignature = fiscalDay.FiscalDayServerSignature
	resp.FiscalDayClosed = fiscalDay.FiscalDayClosed
	resp.LastReceiptGlobalNo = fiscalDay.LastReceiptGlobalNo

	return resp, nil
}

// generateFiscalDayServerSignature generates FDMS signature for fiscal day
func (s *FiscalDayService) generateFiscalDayServerSignature(
	deviceID int,
	fiscalDayNo int,
	fiscalDayDate string,
	fiscalDayUpdated time.Time,
	reconciliationMode models.FiscalDayReconciliationMode,
	counters []models.FiscalDayCounter,
	deviceSignature *models.SignatureData,
) (*models.SignatureDataEx, error) {
	// Generate hash
	var deviceSig []byte
	if deviceSignature != nil {
		deviceSig = deviceSignature.Signature
	}

	hash, err := utils.GenerateFiscalDayServerHash(
		deviceID,
		fiscalDayNo,
		fiscalDayDate,
		fiscalDayUpdated.Format("2006-01-02T15:04:05"),
		reconciliationMode,
		counters,
		deviceSig,
	)
	if err != nil {
		return nil, err
	}

	// Sign with server private key
	signature, err := s.cryptoSvc.SignData(hash)
	if err != nil {
		return nil, err
	}

	// Get certificate thumbprint (placeholder)
	thumbprint := make([]byte, 20)

	return &models.SignatureDataEx{
		SignatureData: models.SignatureData{
			Hash:      hash,
			Signature: signature,
		},
		CertificateThumbprint: thumbprint,
	}, nil
}
