package handlers

import (
	"fiscalization-api/internal/models"
	"fiscalization-api/internal/service"
	"fiscalization-api/pkg/api"

	"github.com/gin-gonic/gin"
)

type ReceiptHandler struct {
	receiptService *service.ReceiptService
}

func NewReceiptHandler(receiptService *service.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{
		receiptService: receiptService,
	}
}

// SubmitReceipt handles POST /api/v1/receipt/submit
func (h *ReceiptHandler) SubmitReceipt(c *gin.Context) {
	deviceID, exists := api.GetDeviceIDFromContext(c)
	if !exists {
		api.UnauthorizedResponse(c, "Device ID not found in context")
		return
	}

	var req models.SubmitReceiptRequest
	if !api.BindJSON(c, &req) {
		return
	}

	req.DeviceID = deviceID

	resp, err := h.receiptService.SubmitReceipt(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// TODO: Implement file upload and status endpoints when needed
// SubmitFileUpload handles POST /api/v1/receipt/file
// GetFileUploadStatus handles GET /api/v1/receipt/file-status
