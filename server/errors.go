package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RespondError(c *gin.Context, status int, code, message string, details any) {
	c.AbortWithStatusJSON(status, gin.H{
		"ok":      false,
		"error":   message,
		"code":    code,
		"details": details,
	})
}

func NotFoundHandler(c *gin.Context) {
	RespondError(c, http.StatusNotFound, "not_found", "endpoint not found", nil)
}
func MethodNotAllowedHandler(c *gin.Context) {
	RespondError(c, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
}
