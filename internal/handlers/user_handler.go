package handlers

import (
	"fiscalization-api/internal/models"
	"fiscalization-api/internal/service"
	"fiscalization-api/pkg/api"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Login handles POST /api/v1/users/login
func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if !api.BindJSON(c, &req) {
		return
	}

	resp, err := h.userService.Login(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// ListUsers handles GET /api/v1/users/list
func (h *UserHandler) ListUsers(c *gin.Context) {
	deviceID, exists := api.GetDeviceIDFromContext(c)
	if !exists {
		api.UnauthorizedResponse(c, "Device ID not found in context")
		return
	}

	pagination := api.GetPaginationParams(c, 20, 100)

	resp, err := h.userService.ListUsers(deviceID, pagination.Offset, pagination.Limit)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// CreateUserBegin handles POST /api/v1/users/create-begin
func (h *UserHandler) CreateUserBegin(c *gin.Context) {
	deviceID, exists := api.GetDeviceIDFromContext(c)
	if !exists {
		api.UnauthorizedResponse(c, "Device ID not found in context")
		return
	}

	var req models.CreateUserBeginRequest
	if !api.BindJSON(c, &req) {
		return
	}

	req.DeviceID = deviceID

	resp, err := h.userService.CreateUserBegin(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// CreateUserConfirm handles POST /api/v1/users/create-confirm
func (h *UserHandler) CreateUserConfirm(c *gin.Context) {
	var req models.CreateUserConfirmRequest
	if !api.BindJSON(c, &req) {
		return
	}

	resp, err := h.userService.CreateUserConfirm(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// UpdateUser handles PUT /api/v1/users/update
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req models.UpdateUserRequest
	if !api.BindJSON(c, &req) {
		return
	}

	resp, err := h.userService.UpdateUser(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}

// ChangePassword handles PUT /api/v1/users/change-password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if !api.BindJSON(c, &req) {
		return
	}

	resp, err := h.userService.ChangePassword(req)
	if err != nil {
		api.ErrorResponse(c, err)
		return
	}

	api.SuccessResponse(c, resp)
}
