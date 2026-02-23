package handlers

import (
	"fiscalization-api/internal/models"
	"fiscalization-api/internal/service"
	"fiscalization-api/pkg/api"

	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	deviceService *service.DeviceService
}

func NewDeviceHandler(deviceService *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

// VerifyTaxpayer handles POST /api/v1/device/verify-taxpayer
func (h *DeviceHandler) VerifyTaxpayer(c *gin.Context) {
	var req models.VerifyTaxpayerRequest
	if !api.BindJSON(c, &req) {
		return
	}

	// Get headers
	modelName := c.GetHeader("DeviceModelName")
	modelVersion := c.GetHeader("DeviceModelVersionNo")

	if modelName == "" || modelVersion == "" {
		api.ValidationErrorResponse(c, "DeviceModelName and DeviceModelVersionNo headers required")
		return
	}

	resp, err := h.deviceService.VerifyTaxpayer(req, modelName, modelVersion)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// RegisterDevice handles POST /api/v1/device/register
func (h *DeviceHandler) RegisterDevice(c *gin.Context) {
	var req models.DeviceRegistrationRequest
	if !api.BindJSON(c, &req) {
		return
	}

	// Get headers
	modelName := c.GetHeader("DeviceModelName")
	modelVersion := c.GetHeader("DeviceModelVersionNo")

	if modelName == "" || modelVersion == "" {
		api.ValidationErrorResponse(c, "DeviceModelName and DeviceModelVersionNo headers required")
		return
	}

	resp, err := h.deviceService.RegisterDevice(req, modelName, modelVersion)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// IssueCertificate handles POST /api/v1/device/issue-certificate
func (h *DeviceHandler) IssueCertificate(c *gin.Context) {
	deviceID, exists := api.GetDeviceIDFromContext(c)
	if !exists {
		api.UnauthorizedResponse(c, "Device ID not found in context")
		return
	}

	var req models.IssueCertificateRequest
	if !api.BindJSON(c, &req) {
		return
	}

	req.DeviceID = deviceID

	resp, err := h.deviceService.IssueCertificate(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// GetConfig handles GET /api/v1/device/config
func (h *DeviceHandler) GetConfig(c *gin.Context) {
	deviceID, exists := api.GetDeviceIDFromContext(c)
	if !exists {
		api.UnauthorizedResponse(c, "Device ID not found in context")
		return
	}

	resp, err := h.deviceService.GetConfig(deviceID)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// GetStatus handles GET /api/v1/device/status
func (h *DeviceHandler) GetStatus(c *gin.Context) {
	deviceID, exists := api.GetDeviceIDFromContext(c)
	if !exists {
		api.UnauthorizedResponse(c, "Device ID not found in context")
		return
	}

	resp, err := h.deviceService.GetStatus(deviceID)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// Ping handles POST /api/v1/device/ping
func (h *DeviceHandler) Ping(c *gin.Context) {
	deviceID, exists := api.GetDeviceIDFromContext(c)
	if !exists {
		api.UnauthorizedResponse(c, "Device ID not found in context")
		return
	}

	resp, err := h.deviceService.Ping(deviceID)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// GetServerCertificate handles GET /api/v1/server/certificate
func (h *DeviceHandler) GetServerCertificate(c *gin.Context) {
	// Optional thumbprint parameter
	thumbprintStr := c.Query("thumbprint")
	var thumbprint []byte
	if thumbprintStr != "" {
		// Decode thumbprint from hex or base64
		// For now, pass as is
	}

	resp, err := h.deviceService.GetServerCertificate(thumbprint)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// GetStockList handles GET /api/v1/stock/list
func (h *DeviceHandler) GetStockList(c *gin.Context) {
	deviceID, exists := api.GetDeviceIDFromContext(c)
	if !exists {
		api.UnauthorizedResponse(c, "Device ID not found in context")
		return
	}

	// Get pagination params
	pagination := api.GetPaginationParams(c, 20, 100)

	// Get query parameters
	hsCode := c.Query("hsCode")
	goodName := c.Query("goodName")
	operator := c.Query("operator")

	req := models.GetStockListRequest{
		DeviceID: deviceID,
		Offset:   pagination.Offset,
		Limit:    pagination.Limit,
		Sort:     &pagination.Sort,
		Order:    &pagination.Order,
	}

	if hsCode != "" {
		req.HSCode = &hsCode
	}
	if goodName != "" {
		req.GoodName = &goodName
	}
	if operator != "" {
		req.Operator = &operator
	}

	resp, err := h.deviceService.GetStockList(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}
