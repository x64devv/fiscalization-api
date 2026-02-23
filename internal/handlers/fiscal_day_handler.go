package handlers

import (
	"fiscalization-api/internal/models"
	"fiscalization-api/internal/service"
	"fiscalization-api/pkg/api"

	"github.com/gin-gonic/gin"
)

type FiscalDayHandler struct {
	fiscalDayService *service.FiscalDayService
}

func NewFiscalDayHandler(fiscalDayService *service.FiscalDayService) *FiscalDayHandler {
	return &FiscalDayHandler{
		fiscalDayService: fiscalDayService,
	}
}

// OpenFiscalDay handles POST /api/v1/fiscal-day/open
func (h *FiscalDayHandler) OpenFiscalDay(c *gin.Context) {
	deviceID, exists := api.GetDeviceIDFromContext(c)
	if !exists {
		api.UnauthorizedResponse(c, "Device ID not found in context")
		return
	}

	req := models.OpenFiscalDayRequest{
		DeviceID: deviceID,
	}

	resp, err := h.fiscalDayService.OpenFiscalDay(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// CloseFiscalDay handles POST /api/v1/fiscal-day/close
func (h *FiscalDayHandler) CloseFiscalDay(c *gin.Context) {
	deviceID, exists := api.GetDeviceIDFromContext(c)
	if !exists {
		api.UnauthorizedResponse(c, "Device ID not found in context")
		return
	}

	var req models.CloseFiscalDayRequest
	if !api.BindJSON(c, &req) {
		return
	}

	req.DeviceID = deviceID

	resp, err := h.fiscalDayService.CloseFiscalDay(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// GetStatus handles GET /api/v1/fiscal-day/status
func (h *FiscalDayHandler) GetStatus(c *gin.Context) {
	deviceID, exists := api.GetDeviceIDFromContext(c)
	if !exists {
		api.UnauthorizedResponse(c, "Device ID not found in context")
		return
	}

	resp, err := h.fiscalDayService.GetFiscalDayStatus(deviceID)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}
