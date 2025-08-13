package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(cfg CORSConfig) gin.HandlerFunc {
	origins := normalize(cfg.AllowedOrigins)
	methods := strings.Join(unique(upperAll(defaultIfEmpty(cfg.AllowedMethods,
		[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}))), ", ")
	headers := strings.Join(unique(headerCaseAll(defaultIfEmpty(cfg.AllowedHeaders,
		[]string{"Content-Type", "Authorization", "X-Requested-With"}))), ", ")
	exposed := strings.Join(unique(headerCaseAll(cfg.ExposedHeaders)), ", ")
	maxAge := ""
	if cfg.MaxAge > 0 {
		maxAge = strconvI(int(cfg.MaxAge / time.Second))
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if originAllowed(origin, origins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowValue(origin, origins))
			if cfg.AllowCredentials {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if exposed != "" {
				c.Writer.Header().Set("Access-Control-Expose-Headers", exposed)
			}
		}
		if c.Request.Method == http.MethodOptions {
			if methods != "" {
				c.Writer.Header().Set("Access-Control-Allow-Methods", methods)
			}
			if headers != "" {
				c.Writer.Header().Set("Access-Control-Allow-Headers", headers)
			}
			if maxAge != "" {
				c.Writer.Header().Set("Access-Control-Max-Age", maxAge)
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func originAllowed(origin string, allowed []string) bool {
	if origin == "" || len(allowed) == 0 {
		return false
	}
	for _, a := range allowed {
		if a == "*" || strings.EqualFold(a, origin) {
			return true
		}
		if strings.HasPrefix(a, "http://*.") || strings.HasPrefix(a, "https://*.") {
			prefix := "://*."
			if i := strings.Index(a, prefix); i > 0 {
				scheme := a[:i]
				domain := a[i+len(prefix):]
				if strings.HasPrefix(strings.ToLower(origin), scheme+"://") &&
					strings.HasSuffix(strings.ToLower(origin), "."+strings.ToLower(domain)) {
					return true
				}
			}
		}
	}
	return false
}
func allowValue(origin string, allowed []string) string {
	for _, a := range allowed {
		if a == "*" {
			return "*"
		}
	}
	return origin
}

/* helpers */

func normalize(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		if s := strings.TrimSpace(v); s != "" {
			out = append(out, s)
		}
	}
	return out
}
func defaultIfEmpty(in, def []string) []string {
	if len(in) == 0 {
		return def
	}
	return in
}
func upperAll(in []string) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = strings.ToUpper(strings.TrimSpace(v))
	}
	return out
}
func headerCaseAll(in []string) []string {
	out := make([]string, len(in))
	for i, v := range in {
		parts := strings.Split(strings.TrimSpace(v), "-")
		for j, p := range parts {
			if p != "" {
				parts[j] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
			}
		}
		out[i] = strings.Join(parts, "-")
	}
	return out
}
func unique(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}
func strconvI(i int) string { return fmtInt(int64(i)) }
func fmtInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	pos := len(b)
	for n > 0 {
		pos--
		b[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}
