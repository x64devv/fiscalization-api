package service

import (
	"fmt"
	"time"

	"fiscalization-api/internal/models"
	"fiscalization-api/internal/repository"
	"fiscalization-api/internal/utils"

	"go.uber.org/zap"
)

type ReceiptService struct {
	receiptRepo    repository.ReceiptRepository
	fiscalDayRepo  repository.FiscalDayRepository
	deviceRepo     repository.DeviceRepository
	validationSvc  *ValidationService
	cryptoSvc      *CryptoService
	logger         *zap.Logger
}

func NewReceiptService(
	receiptRepo repository.ReceiptRepository,
	fiscalDayRepo repository.FiscalDayRepository,
	deviceRepo repository.DeviceRepository,
	validationSvc *ValidationService,
	cryptoSvc *CryptoService,
	logger *zap.Logger,
) *ReceiptService {
	return &ReceiptService{
		receiptRepo:   receiptRepo,
		fiscalDayRepo: fiscalDayRepo,
		deviceRepo:    deviceRepo,
		validationSvc: validationSvc,
		cryptoSvc:     cryptoSvc,
		logger:        logger,
	}
}

// SubmitReceipt submits a receipt in online mode
func (s *ReceiptService) SubmitReceipt(req models.SubmitReceiptRequest) (*models.SubmitReceiptResponse, error) {
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
		return nil, models.NewAPIError(422, "No fiscal day opened", models.ErrCodeRCPT01)
	}

	// Check fiscal day status
	if fiscalDay.Status != models.FiscalDayStatusOpened && fiscalDay.Status != models.FiscalDayStatusCloseFailed {
		return nil, models.NewAPIError(422, "Submitting receipt is not allowed", models.ErrCodeRCPT01)
	}

	// Set fiscal day ID
	req.Receipt.FiscalDayID = fiscalDay.ID

	// Check for duplicate (same deviceID, receiptGlobalNo, and hash)
	existing, err := s.receiptRepo.GetByGlobalNo(req.DeviceID, req.Receipt.ReceiptGlobalNo)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Return existing receipt signature
		s.logger.Info("Duplicate receipt detected, returning existing signature",
			zap.Int("deviceID", req.DeviceID),
			zap.Int("globalNo", req.Receipt.ReceiptGlobalNo),
		)

		return &models.SubmitReceiptResponse{
			OperationID:            generateOperationID(),
			ReceiptID:              existing.ReceiptID,
			ServerDate:             *existing.ServerDate,
			ReceiptServerSignature: *existing.ReceiptServerSignature,
		}, nil
	}

	// Get previous receipt for validation
	var previousReceipt *models.Receipt
	if req.Receipt.ReceiptCounter > 1 {
		previousReceipt, err = s.receiptRepo.GetPreviousReceipt(
			req.DeviceID,
			fiscalDay.ID,
			req.Receipt.ReceiptGlobalNo,
		)
		if err != nil {
			return nil, err
		}
	}

	// Get taxpayer
	taxpayer, err := s.deviceRepo.GetTaxpayer(device.TaxpayerID)
	if err != nil {
		return nil, err
	}

	// Get applicable taxes
	applicableTaxes, err := s.deviceRepo.GetApplicableTaxes()
	if err != nil {
		return nil, err
	}

	// Validate receipt
	validationResult := s.validationSvc.ValidateReceipt(
		&req.Receipt,
		previousReceipt,
		taxpayer,
		applicableTaxes,
		fiscalDay.FiscalDayOpened,
		taxpayer.TaxPayerDayMaxHrs,
	)

	// Validate credit/debit note if applicable
	if req.Receipt.ReceiptType == models.ReceiptTypeCreditNote || req.Receipt.ReceiptType == models.ReceiptTypeDebitNote {
		if req.Receipt.CreditDebitNote != nil && req.Receipt.CreditDebitNote.ReceiptID != nil {
			originalReceipt, err := s.receiptRepo.GetByReceiptID(*req.Receipt.CreditDebitNote.ReceiptID)
			if err != nil {
				return nil, err
			}

			creditNotes, debitNotes, err := s.receiptRepo.GetCreditDebitNotes(*req.Receipt.CreditDebitNote.ReceiptID)
			if err != nil {
				return nil, err
			}

			cdValidation := s.validationSvc.ValidateCreditDebitNote(
				&req.Receipt,
				originalReceipt,
				creditNotes,
				debitNotes,
			)

			// Merge validation results
			validationResult.IsValid = validationResult.IsValid && cdValidation.IsValid
			validationResult.Errors = append(validationResult.Errors, cdValidation.Errors...)
			if cdValidation.Color != nil {
				if validationResult.Color == nil || *cdValidation.Color == models.ValidationColorRed {
					validationResult.Color = cdValidation.Color
				}
			}
		}
	}

	// Verify receipt signature
	var previousHash []byte
	if previousReceipt != nil {
		previousHash = previousReceipt.ReceiptHash
	}

	receiptHash, err := utils.GenerateReceiptHash(&req.Receipt, previousHash)
	if err != nil {
		s.logger.Error("Failed to generate receipt hash", zap.Error(err))
		return nil, fmt.Errorf("failed to generate receipt hash: %w", err)
	}

	// Store the hash
	req.Receipt.ReceiptHash = receiptHash

	// Set validation results
	req.Receipt.ValidationColor = validationResult.Color
	req.Receipt.ValidationErrors = validationResult.Errors

	// Generate server signature
	serverDate := time.Now()
	req.Receipt.ServerDate = &serverDate

	serverSignature, err := s.generateServerSignature(&req.Receipt, serverDate)
	if err != nil {
		s.logger.Error("Failed to generate server signature", zap.Error(err))
		return nil, fmt.Errorf("failed to generate server signature: %w", err)
	}

	req.Receipt.ReceiptServerSignature = serverSignature

	// Save receipt to database
	if err := s.receiptRepo.CreateWithLines(&req.Receipt); err != nil {
		s.logger.Error("Failed to save receipt", zap.Error(err))
		return nil, fmt.Errorf("failed to save receipt: %w", err)
	}

	// Update fiscal day last receipt number
	fiscalDay.LastReceiptGlobalNo = &req.Receipt.ReceiptGlobalNo
	if err := s.fiscalDayRepo.Update(fiscalDay); err != nil {
		s.logger.Warn("Failed to update fiscal day", zap.Error(err))
	}

	s.logger.Info("Receipt submitted successfully",
		zap.Int64("receiptID", req.Receipt.ReceiptID),
		zap.Int("deviceID", req.DeviceID),
		zap.Int("globalNo", req.Receipt.ReceiptGlobalNo),
	)

	return &models.SubmitReceiptResponse{
		OperationID:            generateOperationID(),
		ReceiptID:              req.Receipt.ReceiptID,
		ServerDate:             serverDate,
		ReceiptServerSignature: *serverSignature,
	}, nil
}

// generateServerSignature generates FDMS signature for receipt
func (s *ReceiptService) generateServerSignature(receipt *models.Receipt, serverDate time.Time) (*models.SignatureDataEx, error) {
	// Build signature data: receiptDeviceSignature + receiptID + serverDate
	signatureData := fmt.Sprintf("%s%d%s",
		utils.EncodeBase64(receipt.ReceiptDeviceSignature.Signature),
		receipt.ReceiptID,
		serverDate.Format("2006-01-02T15:04:05"),
	)

	// Sign with server private key
	signature, err := s.cryptoSvc.SignData([]byte(signatureData))
	if err != nil {
		return nil, err
	}

	// Get certificate thumbprint
	// In production, this would come from the server certificate
	thumbprint := make([]byte, 20) // Placeholder

	return &models.SignatureDataEx{
		SignatureData: models.SignatureData{
			Hash:      receipt.ReceiptHash,
			Signature: signature,
		},
		CertificateThumbprint: thumbprint,
	}, nil
}
