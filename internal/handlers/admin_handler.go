package handlers

import (
	"net/http"
	"strconv"
	"time"

	"fiscalization-api/internal/models"
	"fiscalization-api/internal/service"
	"fiscalization-api/pkg/api"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

// POST /api/admin/login
func (h *AdminHandler) Login(c *gin.Context) {
	var req models.AdminLoginRequest
	if !api.BindJSON(c, &req) {
		return
	}
	resp, err := h.adminService.Login(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, resp)
}

// ─── Stats ────────────────────────────────────────────────────────────────────

// GET /api/admin/stats
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.adminService.GetSystemStats()
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, stats)
}

// ─── Taxpayers (Companies) ────────────────────────────────────────────────────

// GET /api/admin/companies
func (h *AdminHandler) ListCompanies(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	resp, err := h.adminService.ListTaxpayers(offset, limit, search)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, resp)
}

// POST /api/admin/companies
func (h *AdminHandler) CreateCompany(c *gin.Context) {
	var req models.CreateTaxpayerRequest
	if !api.BindJSON(c, &req) {
		return
	}
	tp, err := h.adminService.CreateTaxpayer(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	c.JSON(http.StatusCreated, tp)
}

// GET /api/admin/companies/:id
func (h *AdminHandler) GetCompany(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ValidationErrorResponse(c, "Invalid company ID")
		return
	}
	tp, err := h.adminService.GetTaxpayer(id)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, tp)
}

// PUT /api/admin/companies/:id
func (h *AdminHandler) UpdateCompany(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ValidationErrorResponse(c, "Invalid company ID")
		return
	}
	var req models.UpdateTaxpayerRequest
	if !api.BindJSON(c, &req) {
		return
	}
	req.ID = id
	tp, err := h.adminService.UpdateTaxpayer(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, tp)
}

// PATCH /api/admin/companies/:id/status
func (h *AdminHandler) SetCompanyStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ValidationErrorResponse(c, "Invalid company ID")
		return
	}
	var body struct {
		Status string `json:"status" binding:"required,oneof=Active Inactive"`
	}
	if !api.BindJSON(c, &body) {
		return
	}
	if err := h.adminService.SetTaxpayerStatus(id, body.Status); err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, gin.H{"message": "Status updated"})
}

// GET /api/admin/companies/:id/devices
func (h *AdminHandler) ListCompanyDevices(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ValidationErrorResponse(c, "Invalid company ID")
		return
	}
	devices, err := h.adminService.ListDevicesByTaxpayer(id)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, gin.H{"total": len(devices), "rows": devices})
}

// ─── Devices ──────────────────────────────────────────────────────────────────

// GET /api/admin/devices
func (h *AdminHandler) ListAllDevices(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	resp, err := h.adminService.ListAllDevices(offset, limit)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, resp)
}

// POST /api/admin/devices
func (h *AdminHandler) ProvisionDevice(c *gin.Context) {
	var req models.AdminCreateDeviceRequest
	if !api.BindJSON(c, &req) {
		return
	}
	device, err := h.adminService.CreateDevice(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	c.JSON(http.StatusCreated, device)
}

// PATCH /api/admin/devices/:deviceID/status
func (h *AdminHandler) SetDeviceStatus(c *gin.Context) {
	deviceID, err := strconv.Atoi(c.Param("deviceID"))
	if err != nil {
		api.ValidationErrorResponse(c, "Invalid device ID")
		return
	}
	var body struct {
		Status string `json:"status" binding:"required,oneof=Active Blocked Revoked"`
	}
	if !api.BindJSON(c, &body) {
		return
	}
	if err := h.adminService.UpdateDeviceStatus(deviceID, body.Status); err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, gin.H{"message": "Device status updated"})
}

// PATCH /api/admin/devices/:deviceID/mode
func (h *AdminHandler) SetDeviceMode(c *gin.Context) {
	deviceID, err := strconv.Atoi(c.Param("deviceID"))
	if err != nil {
		api.ValidationErrorResponse(c, "Invalid device ID")
		return
	}
	var body struct {
		Mode int `json:"mode" binding:"oneof=0 1"`
	}
	if !api.BindJSON(c, &body) {
		return
	}
	if err := h.adminService.UpdateDeviceMode(deviceID, body.Mode); err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, gin.H{"message": "Device mode updated"})
}

// ─── Fiscal Days (cross-tenant) ───────────────────────────────────────────────

// GET /api/admin/fiscal-days
func (h *AdminHandler) ListFiscalDays(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var taxpayerID *int64
	var deviceID *int

	if v := c.Query("taxpayerID"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		taxpayerID = &id
	}
	if v := c.Query("deviceID"); v != "" {
		id, _ := strconv.Atoi(v)
		deviceID = &id
	}

	resp, err := h.adminService.ListFiscalDays(taxpayerID, deviceID, offset, limit)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, resp)
}

// ─── Receipts (cross-tenant) ──────────────────────────────────────────────────

// GET /api/admin/receipts
func (h *AdminHandler) ListReceipts(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var taxpayerID *int64
	var deviceID *int
	var from, to *time.Time

	if v := c.Query("taxpayerID"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		taxpayerID = &id
	}
	if v := c.Query("deviceID"); v != "" {
		id, _ := strconv.Atoi(v)
		deviceID = &id
	}
	if v := c.Query("from"); v != "" {
		t, _ := time.Parse(time.RFC3339, v)
		from = &t
	}
	if v := c.Query("to"); v != "" {
		t, _ := time.Parse(time.RFC3339, v)
		to = &t
	}

	resp, err := h.adminService.ListReceipts(taxpayerID, deviceID, from, to, offset, limit)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, resp)
}

// ─── Audit Logs ───────────────────────────────────────────────────────────────

// GET /api/admin/audit
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	entityType := c.Query("entityType")

	var entityID *int64
	if v := c.Query("entityID"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		entityID = &id
	}

	resp, err := h.adminService.ListAuditLogs(entityType, entityID, offset, limit)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}
	api.SuccessResponse(c, resp)
}
