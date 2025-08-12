package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func RespondWithError(c *gin.Context, statusCode int, message ...string) {
	response := ErrorResponse{
		Message: http.StatusText(statusCode),
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	c.AbortWithStatusJSON(statusCode, response)
}

func InternalError(c *gin.Context, message ...string) {
	RespondWithError(c, http.StatusInternalServerError, message...)
}

func BadRequest(c *gin.Context, message ...string) {
	RespondWithError(c, http.StatusBadRequest, message...)
}

func Unauthorized(c *gin.Context, message ...string) {
	RespondWithError(c, http.StatusUnauthorized, message...)
}

func Forbidden(c *gin.Context, message ...string) {
	RespondWithError(c, http.StatusForbidden, message...)
}

func NotFound(c *gin.Context, message ...string) {
	RespondWithError(c, http.StatusNotFound, message...)
}
