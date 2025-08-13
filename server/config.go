package server

import "time"

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

type HTTPTimeouts struct {
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type LogConfig struct {
	AccessFile         string
	ErrorFile          string
	RotateMaxSizeBytes int64 // 0 — без ротации
	RotateBackups      int
}

type AuthConfig struct {
	// откуда брать токены
	AuthHeader    string // "Authorization"
	BearerPrefix  string // "Bearer "
	AccessCookie  string // например, "access_token"
	RefreshCookie string // например, "refresh_token"
	// включение стандартных мидлваров
	EnableAccessMiddleware  bool
	EnableRefreshMiddleware bool
}

type TimeoutConfig struct {
	// пер‑запросный timeout (добавляет ctx с дедлайном в c.Request.Context())
	// 0 — выключено.
	RequestTimeout time.Duration
	// что возвращать, если сработал timeout
	GatewayTimeoutStatus int // по умолчанию 504
}

type Config struct {
	Addr     int
	Release  bool
	BasePath string

	CORS       CORSConfig
	Timeouts   HTTPTimeouts
	Log        LogConfig
	Auth       AuthConfig
	PerRequest TimeoutConfig

	ShutdownWait time.Duration
	// печатать таблицу роутов при старте
	PrintRoutes bool
}
