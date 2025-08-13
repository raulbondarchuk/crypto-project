package main

import (
	"net/http"
	"time"

	"api/server"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := server.Config{
		Addr:        8080,
		BasePath:    "/api/v1",
		PrintRoutes: true,
		CORS: server.CORSConfig{
			AllowedOrigins:   []string{"http://localhost:3000", "https://*.example.com"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		},
		Timeouts: server.HTTPTimeouts{
			ReadTimeout:       10 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		PerRequest: server.TimeoutConfig{
			RequestTimeout:       3 * time.Second, // пер‑запросный timeout
			GatewayTimeoutStatus: 504,
		},
		Log: server.LogConfig{
			AccessFile:         "logs/access.log",
			ErrorFile:          "logs/error.log",
			RotateMaxSizeBytes: 10 << 20, // 10MB
			RotateBackups:      5,
		},
		Auth: server.AuthConfig{
			AuthHeader:             "Authorization",
			BearerPrefix:           "Bearer ",
			AccessCookie:           "access_token",
			RefreshCookie:          "refresh_token",
			EnableAccessMiddleware: true, // подключить мидлвар в корневую группу
		},
	}

	// свой валидатор токенов (замени на PASETO)
	val := server.StubValidator{}

	srv, err := server.New(
		cfg,
		server.WithTokenValidator(val),
		server.WithRegistrar(server.HandlerFuncRegistrar(func(r *gin.RouterGroup) {
			// защищённые эндпоинты (access мидлвар уже повешен на root, см. cfg.Auth)
			r.GET("/me", func(c *gin.Context) {
				claims, _ := c.Get("access_claims")
				c.JSON(http.StatusOK, gin.H{"ok": true, "claims": claims})
			})
			// «длинная» операция (проверяем timeout)
			r.GET("/slow", func(c *gin.Context) {
				select {
				case <-time.After(5 * time.Second):
					c.JSON(http.StatusOK, gin.H{"ok": true})
				case <-c.Request.Context().Done():
					// хендлер уведет timeout mw
					return
				}
			})
			// обмен refresh -> access (вешай RefreshMiddleware только здесь, если нужно)
			r.POST("/auth/refresh", server.AuthOnly(val, false), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true, "new": "token"})
			})
		})),
	)
	if err != nil {
		panic(err)
	}
	if err := srv.Start(); err != nil {
		panic(err)
	}
}
