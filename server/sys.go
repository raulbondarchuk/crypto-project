package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SysEndpoints(s *Server) {
	// системные эндпоинты
	sys := s.engine.Group("/sys")
	{
		sys.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true, "status": "live"})
		})

		sys.GET("/readyz", func(c *gin.Context) {
			// тут можно расширить логику готовности
			c.JSON(http.StatusOK, gin.H{"ok": true, "status": "ready"})
		})

		sys.GET("/info", func(c *gin.Context) {
			uptime := time.Since(s.startTime).Round(time.Second)
			c.JSON(http.StatusOK, gin.H{
				"ok":       true,
				"version":  "1.0.0", // можно из конфига или build-флага
				"started":  s.startTime.Format(time.RFC3339),
				"uptime":   uptime.String(),
				"addr":     s.cfg.Addr,
				"basePath": s.cfg.BasePath,
			})
		})

		sys.GET("/routes", func(c *gin.Context) {
			routes := s.engine.Routes()
			c.JSON(http.StatusOK, gin.H{"ok": true, "routes": routes})
		})

		sys.GET("/routes/table", func(c *gin.Context) {
			LogRoutes(s.engine) // печать в stdout
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
	}
}
