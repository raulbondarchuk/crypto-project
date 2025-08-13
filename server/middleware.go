package server

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestID — простой ID для трассировки.
func RequestID(headerName string) gin.HandlerFunc {
	if headerName == "" {
		headerName = "X-Request-Id"
	}
	return func(c *gin.Context) {
		id := c.GetHeader(headerName)
		if id == "" {
			var b [12]byte
			_, _ = rand.Read(b[:])
			id = hex.EncodeToString(b[:])
		}
		c.Set("request_id", id)
		c.Writer.Header().Set(headerName, id)
		c.Next()
	}
}

// RecoveryJSON — перехватывает паники, пишет в лог и возвращает JSON 500.
func RecoveryJSON(errWriter io.Writer) gin.HandlerFunc {
	logger := log.New(errWriter, "[panic] ", log.LstdFlags|log.Lmsgprefix)
	return gin.CustomRecoveryWithWriter(errWriter, func(c *gin.Context, err interface{}) {
		logger.Printf("%s %s | panic: %v", c.Request.Method, c.Request.URL.Path, err)
		RespondError(c, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	})
}

// ErrorCapture — отправляйте ошибки через c.Error(err); мы их залогируем.
func ErrorCapture(errWriter io.Writer) gin.HandlerFunc {
	logger := log.New(errWriter, "[error] ", log.LstdFlags|log.Lmsgprefix)
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Printf("%s %s -> %d in %s | %v",
					c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start), e.Err)
			}
		}
	}
}

// Health — готовые health‑эндпоинты.
func Health() (live gin.HandlerFunc, ready gin.HandlerFunc) {
	live = func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true, "status": "live"}) }
	ready = func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true, "status": "ready"}) }
	return
}
