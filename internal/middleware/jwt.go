package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/pvz-service/internal/dto"
	"github.com/senyabanana/pvz-service/internal/infrastructure/jwtutil"
)

const (
	authHeader   = "Authorization"
	bearerPrefix = "Bearer "
	userIDKey    = "user_id"
	userRoleKey  = "user_role"
)

func RequireRole(secretKey string, log *logrus.Logger, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(authHeader)
		if header == "" {
			log.Warn("missing Authorization header")
			dto.Unauthorized(c, "missing Authorization header")
			return
		}

		tokenString := strings.TrimPrefix(header, bearerPrefix)
		if tokenString == header {
			log.Warn("invalid bearer format")
			dto.Unauthorized(c, "invalid bearer format")
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &jwtutil.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		if err != nil || !token.Valid {
			log.Warnf("invalid token: %v", err)
			dto.Unauthorized(c, "invalid token")
			return
		}

		claims, ok := token.Claims.(*jwtutil.JWTClaims)
		if !ok {
			dto.Unauthorized(c, "claims cannot be read")
			return
		}

		for _, role := range allowedRoles {
			if claims.Role == role {
				c.Set(userIDKey, claims.UserID)
				c.Set(userRoleKey, claims.Role)
				c.Next()
				return
			}
		}

		log.Infof("forbidden access: user role=%s not in allowedRoles=%v", claims.Role, allowedRoles)
		dto.Forbidden(c, "insufficient access rights")
	}
}
