package server

import "github.com/gin-gonic/gin"

// AuthOnly — вернуть мидлвар только для выбранного роута/группы.
// access=true -> access middleware, иначе refresh.
func AuthOnly(v TokenValidator, access bool) gin.HandlerFunc {
	a := newAuth(AuthConfig{AuthHeader: "Authorization", BearerPrefix: "Bearer "}, v)
	if access {
		return a.AccessMiddleware()
	}
	return a.RefreshMiddleware()
}
