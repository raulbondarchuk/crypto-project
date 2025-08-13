package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Claims map[string]any

// TokenValidator — ваш код реализует эту штуку (PASETO/JWT/что угодно).
type TokenValidator interface {
	ValidateAccess(c *gin.Context, token string) (Claims, error)
	ValidateRefresh(c *gin.Context, token string) (Claims, error)
}

type Auth struct {
	cfg       AuthConfig
	validator TokenValidator
}

func newAuth(cfg AuthConfig, v TokenValidator) *Auth {
	if cfg.AuthHeader == "" {
		cfg.AuthHeader = "Authorization"
	}
	if cfg.BearerPrefix == "" {
		cfg.BearerPrefix = "Bearer "
	}
	return &Auth{cfg: cfg, validator: v}
}

func (a *Auth) AccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := a.pickToken(c, true)
		if tok == "" {
			RespondError(c, http.StatusUnauthorized, "no_token", "access token missing", nil)
			return
		}
		claims, err := a.validator.ValidateAccess(c, tok)
		if err != nil {
			_ = c.Error(err)
			RespondError(c, http.StatusUnauthorized, "invalid_token", "invalid access token", nil)
			return
		}
		c.Set("access_claims", claims)
		c.Next()
	}
}

func (a *Auth) RefreshMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := a.pickToken(c, false)
		if tok == "" {
			RespondError(c, http.StatusUnauthorized, "no_token", "refresh token missing", nil)
			return
		}
		claims, err := a.validator.ValidateRefresh(c, tok)
		if err != nil {
			_ = c.Error(err)
			RespondError(c, http.StatusUnauthorized, "invalid_token", "invalid refresh token", nil)
			return
		}
		c.Set("refresh_claims", claims)
		c.Next()
	}
}

func (a *Auth) pickToken(c *gin.Context, access bool) string {
	// 1) Authorization: Bearer xxx
	h := c.GetHeader(a.cfg.AuthHeader)
	if h != "" && strings.HasPrefix(h, a.cfg.BearerPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(h, a.cfg.BearerPrefix))
	}
	// 2) Cookie
	if access && a.cfg.AccessCookie != "" {
		if v, err := c.Cookie(a.cfg.AccessCookie); err == nil && v != "" {
			return v
		}
	}
	if !access && a.cfg.RefreshCookie != "" {
		if v, err := c.Cookie(a.cfg.RefreshCookie); err == nil && v != "" {
			return v
		}
	}
	return ""
}

// StubValidator — на время разработки.
type StubValidator struct{}

func (StubValidator) ValidateAccess(_ *gin.Context, tok string) (Claims, error) {
	if tok == "ok" {
		return Claims{"sub": "demo", "role": "user"}, nil
	}
	return nil, errors.New("bad token")
}
func (StubValidator) ValidateRefresh(_ *gin.Context, tok string) (Claims, error) {
	if tok == "ok" {
		return Claims{"sub": "demo"}, nil
	}
	return nil, errors.New("bad token")
}
