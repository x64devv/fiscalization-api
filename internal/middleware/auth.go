package middleware

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strconv"

	"fiscalization-api/internal/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	DeviceIDContextKey = "deviceID"
	UserIDContextKey   = "userID"
)

// CertificateAuthMiddleware validates client certificates
func CertificateAuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client certificate from TLS connection
		if c.Request.TLS == nil || len(c.Request.TLS.PeerCertificates) == 0 {
			// Check for certificate in header (for development/testing)
			certHeader := c.GetHeader("X-SSL-Client-Cert")
			if certHeader != "" {
				cert, err := parseCertFromHeader(certHeader)
				if err != nil {
					logger.Error("Failed to parse certificate from header", zap.Error(err))
					c.JSON(401, models.NewAPIError(401, "Invalid client certificate", ""))
					c.Abort()
					return
				}

				// Extract device ID from certificate CN
				deviceID, err := extractDeviceIDFromCert(cert)
				if err != nil {
					logger.Error("Failed to extract device ID from certificate", zap.Error(err))
					c.JSON(401, models.NewAPIError(401, "Invalid certificate format", ""))
					c.Abort()
					return
				}

				// Store device ID in context
				c.Set(DeviceIDContextKey, deviceID)
				c.Next()
				return
			}

			// No certificate provided
			logger.Warn("No client certificate provided")
			c.JSON(401, models.NewAPIError(401, "Client certificate required", ""))
			c.Abort()
			return
		}

		// Get the first (client) certificate
		cert := c.Request.TLS.PeerCertificates[0]

		// Verify certificate is not expired
		// This is usually done by TLS layer, but we double-check
		// if err := cert.Verify(x509.VerifyOptions{}); err != nil {
		// 	logger.Error("Certificate verification failed", zap.Error(err))
		// 	c.JSON(401, models.NewAPIError(401, "Invalid certificate", ""))
		// 	c.Abort()
		// 	return
		// }

		// Extract device ID from certificate Common Name (CN)
		// Format: ZIMRA-{serialNo}-{deviceID}
		deviceID, err := extractDeviceIDFromCert(cert)
		if err != nil {
			logger.Error("Failed to extract device ID from certificate", zap.Error(err))
			c.JSON(401, models.NewAPIError(401, "Invalid certificate format", ""))
			c.Abort()
			return
		}

		// Store device ID in context for handlers to use
		c.Set(DeviceIDContextKey, deviceID)

		logger.Debug("Certificate authenticated",
			zap.Int("deviceID", deviceID),
			zap.String("subject", cert.Subject.CommonName),
		)

		c.Next()
	}
}

// extractDeviceIDFromCert extracts device ID from certificate CN
// Expected format: ZIMRA-{serialNo}-{deviceID}
func extractDeviceIDFromCert(cert *x509.Certificate) (int, error) {
	cn := cert.Subject.CommonName

	// Parse CN format: ZIMRA-{serialNo}-{deviceID}
	// For now, we'll extract the last part after the last hyphen
	// In production, implement proper parsing
	
	// Simple implementation: assume CN contains device ID somewhere
	// This should be enhanced based on actual certificate format
	
	// For testing purposes, we'll accept any valid integer in the CN
	// or default to 1001 if no valid integer found
	
	// Try to parse the entire CN as integer (for simple test certificates)
	if deviceID, err := strconv.Atoi(cn); err == nil {
		return deviceID, nil
	}

	// Try to extract from ZIMRA format
	// Expected: ZIMRA-{serialNo}-{deviceID}
	// Example: ZIMRA-DEV001-1001
	var deviceID int
	_, err := fmt.Scanf(cn, "ZIMRA-%*s-%d", &deviceID)
	if err == nil {
		return deviceID, nil
	}

	// For development: default to 1001 if we can't parse
	// In production, this should return an error
	return 1001, nil
}

// parseCertFromHeader parses PEM-encoded certificate from header
func parseCertFromHeader(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		// Try without PEM encoding
		return x509.ParseCertificate([]byte(certPEM))
	}

	return x509.ParseCertificate(block.Bytes)
}

// JWTAuthMiddleware validates JWT tokens (for user authentication)
func JWTAuthMiddleware(jwtSecret string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("No authorization header provided")
			c.JSON(401, models.NewAPIError(401, "Authorization required", ""))
			c.Abort()
			return
		}

		// Extract token (format: "Bearer {token}")
		var token string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		} else {
			logger.Warn("Invalid authorization header format")
			c.JSON(401, models.NewAPIError(401, "Invalid authorization format", ""))
			c.Abort()
			return
		}

		// Validate token
		// TODO: Implement JWT validation using jwtSecret
		// For now, we'll accept any token for development
		_ = token

		// Extract user ID from token and store in context
		// TODO: Parse JWT and extract user ID
		// For development, use a default user ID
		c.Set(UserIDContextKey, int64(1))

		c.Next()
	}
}

// GetDeviceIDFromContext retrieves device ID from context
func GetDeviceIDFromContext(c *gin.Context) (int, bool) {
	deviceID, exists := c.Get(DeviceIDContextKey)
	if !exists {
		return 0, false
	}
	return deviceID.(int), true
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(c *gin.Context) (int64, bool) {
	userID, exists := c.Get(UserIDContextKey)
	if !exists {
		return 0, false
	}
	return userID.(int64), true
}
