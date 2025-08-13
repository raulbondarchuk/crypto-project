package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg        Config
	engine     *gin.Engine
	root       *gin.RouterGroup
	httpServer *http.Server

	accessOut io.WriteCloser
	errorOut  io.WriteCloser

	beforeStart    []func(*gin.Engine) error
	beforeStop     []func(*gin.Engine)
	routeRegs      []RouteRegistrar
	engineMutators []func(*gin.Engine)

	tokenValidator TokenValidator
	auth           *Auth

	startTime time.Time
}

func New(cfg Config, opts ...Option) (*Server, error) {
	if cfg.Addr <= 0 {
		cfg.Addr = 8080
	}
	if cfg.ShutdownWait <= 0 {
		cfg.ShutdownWait = 10 * time.Second
	}
	if cfg.Timeouts.ReadTimeout == 0 {
		cfg.Timeouts.ReadTimeout = 10 * time.Second
	}
	if cfg.Timeouts.ReadHeaderTimeout == 0 {
		cfg.Timeouts.ReadHeaderTimeout = 5 * time.Second
	}
	if cfg.Timeouts.WriteTimeout == 0 {
		cfg.Timeouts.WriteTimeout = 15 * time.Second
	}
	if cfg.Timeouts.IdleTimeout == 0 {
		cfg.Timeouts.IdleTimeout = 60 * time.Second
	}

	if cfg.Release {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{cfg: cfg, startTime: time.Now()}

	var err error
	s.accessOut, err = newRotatingWriter(cfg.Log.AccessFile, cfg.Log.RotateMaxSizeBytes, cfg.Log.RotateBackups)
	if err != nil {
		return nil, err
	}
	if cfg.Log.AccessFile == "" {
		s.accessOut = nopCloser{Writer: os.Stdout}
	}

	s.errorOut, err = newRotatingWriter(cfg.Log.ErrorFile, cfg.Log.RotateMaxSizeBytes, cfg.Log.RotateBackups)
	if err != nil {
		return nil, err
	}
	if cfg.Log.ErrorFile == "" {
		s.errorOut = nopCloser{Writer: os.Stderr}
	}

	s.engine = gin.New()
	s.engine.Use(gin.LoggerWithWriter(s.accessOut))
	s.engine.Use(RecoveryJSON(s.errorOut))
	s.engine.Use(ErrorCapture(s.errorOut))
	s.engine.Use(RequestID("X-Request-Id"))
	s.engine.Use(CORSMiddleware(cfg.CORS))
	if cfg.PerRequest.RequestTimeout > 0 {
		s.engine.Use(TimeoutMiddleware(cfg.PerRequest))
	}

	// 404/405
	s.engine.NoRoute(NotFoundHandler)
	s.engine.NoMethod(MethodNotAllowedHandler)

	// базовая группа
	if cfg.BasePath != "" {
		s.root = s.engine.Group(cfg.BasePath)
	} else {
		s.root = &s.engine.RouterGroup
	}

	// применяем опции
	for _, o := range opts {
		o(s)
	}

	// auth (по желанию)
	if s.tokenValidator == nil {
		s.tokenValidator = StubValidator{}
	}
	s.auth = newAuth(cfg.Auth, s.tokenValidator)
	if cfg.Auth.EnableAccessMiddleware {
		s.root.Use(s.auth.AccessMiddleware())
	}

	// мутации движка (pprof/метрики/и т.д.)
	for _, m := range s.engineMutators {
		m(s.engine)
	}

	// health на корне (по желанию оставить для k8s)
	live, ready := Health()
	s.GET("/livez", live)
	s.GET("/readyz", ready)

	// подключаем регистраторы
	for _, rr := range s.routeRegs {
		rr.Register(s.root)
	}

	// ✅ системные эндпоинты регистрируем АВТОМАТИЧЕСКИ
	SysEndpoints(s)

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Addr),
		Handler:           s.engine,
		ReadTimeout:       cfg.Timeouts.ReadTimeout,
		ReadHeaderTimeout: cfg.Timeouts.ReadHeaderTimeout,
		WriteTimeout:      cfg.Timeouts.WriteTimeout,
		IdleTimeout:       cfg.Timeouts.IdleTimeout,
	}

	// печать роутов при старте? (после SysEndpoints)
	if cfg.PrintRoutes {
		LogRoutes(s.engine)
	}

	return s, nil
}

func (s *Server) Engine() *gin.Engine    { return s.engine }
func (s *Server) Root() *gin.RouterGroup { return s.root }

// Sugar
func (s *Server) GET(path string, h ...gin.HandlerFunc)    { s.root.GET(path, h...) }
func (s *Server) POST(path string, h ...gin.HandlerFunc)   { s.root.POST(path, h...) }
func (s *Server) PUT(path string, h ...gin.HandlerFunc)    { s.root.PUT(path, h...) }
func (s *Server) PATCH(path string, h ...gin.HandlerFunc)  { s.root.PATCH(path, h...) }
func (s *Server) DELETE(path string, h ...gin.HandlerFunc) { s.root.DELETE(path, h...) }
func (s *Server) Group(path string, f func(g *gin.RouterGroup)) {
	g := s.root.Group(path)
	f(g)
}

func (s *Server) Start() error {
	for _, h := range s.beforeStart {
		if err := h(s.engine); err != nil {
			return err
		}
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch
		_ = s.Shutdown(context.Background())
	}()

	log.New(s.errorOut, "[server] ", log.LstdFlags|log.Lmsgprefix).Printf("listening on %s", fmt.Sprintf(":%d", s.cfg.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.New(s.errorOut, "[server] ", log.LstdFlags|log.Lmsgprefix).Printf("listen error: %v", err)
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	for _, h := range s.beforeStop {
		h(s.engine)
	}
	ctx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownWait)
	defer cancel()
	err := s.httpServer.Shutdown(ctx)
	_ = s.accessOut.Close()
	_ = s.errorOut.Close()
	return err
}
