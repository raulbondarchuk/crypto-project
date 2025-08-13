package server

import "github.com/gin-gonic/gin"

type Option func(*Server)

func WithRegistrar(r RouteRegistrar) Option {
	return func(s *Server) { s.routeRegs = append(s.routeRegs, r) }
}
func WithEngineMutator(f func(*gin.Engine)) Option {
	return func(s *Server) { s.engineMutators = append(s.engineMutators, f) }
}
func WithBeforeStart(h func(*gin.Engine) error) Option {
	return func(s *Server) { s.beforeStart = append(s.beforeStart, h) }
}
func WithBeforeStop(h func(*gin.Engine)) Option {
	return func(s *Server) { s.beforeStop = append(s.beforeStop, h) }
}
func WithTokenValidator(v TokenValidator) Option {
	return func(s *Server) { s.tokenValidator = v }
}
