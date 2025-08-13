package server

import "github.com/gin-gonic/gin"

// RouteRegistrar — любой модуль может реализовать этот интерфейс и
// зарегистрировать свои эндпоинты, не зная о сервере.
type RouteRegistrar interface {
	Register(r *gin.RouterGroup)
}

// HandlerFuncRegistrar — лаконичный способ: функция как регистратор.
type HandlerFuncRegistrar func(r *gin.RouterGroup)

func (f HandlerFuncRegistrar) Register(r *gin.RouterGroup) { f(r) }
