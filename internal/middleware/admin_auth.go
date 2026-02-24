package middleware

import (
	"strings"

	"fiscalization-api/pkg/api"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AdminAuthMiddleware validates the admin JWT token on protected admin routes
func AdminAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			api.UnauthorizedResponse(c, "Missing Authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			api.UnauthorizedResponse(c, "Invalid Authorization header format")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			api.UnauthorizedResponse(c, "Invalid or expired token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			api.UnauthorizedResponse(c, "Invalid token claims")
			c.Abort()
			return
		}

		role, _ := claims["role"].(string)
		if role != "superadmin" {
			api.UnauthorizedResponse(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Set("admin_username", claims["sub"])
		c.Set("admin_role", role)
		c.Next()
	}
}
