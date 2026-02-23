package api

import (
	"net/http"
	"strconv"

	"fiscalization-api/internal/models"

	"github.com/gin-gonic/gin"
)

// SuccessResponse sends a successful JSON response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// CreatedResponse sends a 201 created response
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// NoContentResponse sends a 204 no content response
func NoContentResponse(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, err error) {
	if apiErr, ok := err.(*models.APIError); ok {
		c.JSON(apiErr.Status, apiErr)
		return
	}
	
	// Default to 500 internal server error
	c.JSON(http.StatusInternalServerError, models.APIError{
		Type:   "about:blank",
		Title:  "Internal server error",
		Status: http.StatusInternalServerError,
	})
}

// ValidationErrorResponse sends a 400 bad request response
func ValidationErrorResponse(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, models.APIError{
		Type:   "about:blank",
		Title:  message,
		Status: http.StatusBadRequest,
	})
}

// UnauthorizedResponse sends a 401 unauthorized response
func UnauthorizedResponse(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, models.APIError{
		Type:   "about:blank",
		Title:  message,
		Status: http.StatusUnauthorized,
	})
}

// ForbiddenResponse sends a 403 forbidden response
func ForbiddenResponse(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, models.APIError{
		Type:   "about:blank",
		Title:  message,
		Status: http.StatusForbidden,
	})
}

// NotFoundResponse sends a 404 not found response
func NotFoundResponse(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, models.APIError{
		Type:   "about:blank",
		Title:  message,
		Status: http.StatusNotFound,
	})
}

// ConflictResponse sends a 409 conflict response
func ConflictResponse(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, models.APIError{
		Type:   "about:blank",
		Title:  message,
		Status: http.StatusConflict,
	})
}

// UnprocessableEntityResponse sends a 422 unprocessable entity response
func UnprocessableEntityResponse(c *gin.Context, message, errorCode string) {
	c.JSON(http.StatusUnprocessableEntity, models.APIError{
		Type:      "about:blank",
		Title:     message,
		Status:    http.StatusUnprocessableEntity,
		ErrorCode: errorCode,
	})
}

// GetDeviceIDFromContext gets device ID from gin context
func GetDeviceIDFromContext(c *gin.Context) (int, bool) {
	if deviceID, exists := c.Get("deviceID"); exists {
		if id, ok := deviceID.(int); ok {
			return id, true
		}
	}
	return 0, false
}

// GetUserIDFromContext gets user ID from gin context
func GetUserIDFromContext(c *gin.Context) (int64, bool) {
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(int64); ok {
			return id, true
		}
	}
	return 0, false
}

// ValidateHeaders validates required headers
func ValidateHeaders(c *gin.Context, required ...string) bool {
	for _, header := range required {
		if c.GetHeader(header) == "" {
			ValidationErrorResponse(c, "Missing required header: "+header)
			return false
		}
	}
	return true
}

// BindJSON binds JSON and handles errors
func BindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		ValidationErrorResponse(c, "Invalid request body: "+err.Error())
		return false
	}
	return true
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Offset int
	Limit  int
	Sort   string
	Order  string
}

// GetPaginationParams extracts pagination params from query
func GetPaginationParams(c *gin.Context, defaultLimit int, maxLimit int) PaginationParams {
	offset := 0
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if val, err := strconv.Atoi(offsetParam); err == nil && val >= 0 {
			offset = val
		}
	}
	
	limit := defaultLimit
	if limitParam := c.Query("limit"); limitParam != "" {
		if val, err := strconv.Atoi(limitParam); err == nil && val > 0 {
			limit = min(val, maxLimit)
		}
	}
	
	sort := c.DefaultQuery("sort", "")
	order := c.DefaultQuery("order", "asc")
	
	return PaginationParams{
		Offset: offset,
		Limit:  limit,
		Sort:   sort,
		Order:  order,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
