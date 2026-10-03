package middleware

import (
	"strings"

	"siakad-mini/internal/config"
	"siakad-mini/internal/utils"

	"github.com/gin-gonic/gin"
)

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.JSONError(c, 401, "Token tidak ditemukan", nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.JSONError(c, 401, "Format token tidak valid", nil)
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(cfg.JWTSecret, parts[1])
		if err != nil {
			utils.JSONError(c, 401, "Token tidak valid atau kedaluwarsa", nil)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		roleStr, _ := role.(string)

		for _, r := range roles {
			if r == roleStr {
				c.Next()
				return
			}
		}

		utils.JSONError(c, 403, "Akses ditolak", nil)
		c.Abort()
	}
}
