package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware — мягкий пер‑запросный timeout.
// Инжектит дедлайн в c.Request.Context(); если истёк — возвращает 504 и прерывает цепочку.
func TimeoutMiddleware(cfg TimeoutConfig) gin.HandlerFunc {
	d := cfg.RequestTimeout
	if d <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	status := cfg.GatewayTimeoutStatus
	if status == 0 {
		status = http.StatusGatewayTimeout
	}
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()

		done := make(chan struct{})
		req := c.Request.Clone(ctx) // копия с новым контекстом
		c.Request = req

		go func() { c.Next(); close(done) }()

		select {
		case <-done:
			return
		case <-ctx.Done():
			// если хендлеры не успели ответить
			if !c.Writer.Written() {
				RespondError(c, status, "timeout", "request timed out", nil)
			}
			c.Abort()
			return
		}
	}
}
