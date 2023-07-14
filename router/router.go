package router

import (
	"MDEP/controller"

	"github.com/gin-gonic/gin"
)

type Route struct {
	Method  func(engine *gin.Engine, path string, handler func(c *gin.Context))
	Path    string
	Handler func(c *gin.Context)
}

var routes []Route

func GET(engine *gin.Engine, path string, handler func(c *gin.Context)) {
	engine.GET(path, handler)
}

func POST(engine *gin.Engine, path string, handler func(c *gin.Context)) {
	engine.POST(path, handler)
}
func register(method func(engine *gin.Engine, path string, handler func(c *gin.Context)), path string, handler func(c *gin.Context)) {
	routes = append(routes, Route{method, path, handler})
}

func init() {
	register(GET, "/api/binary", controller.DetectorList)
}

func NewRouter() *gin.Engine {
	router := gin.Default()
	for _, route := range routes {
		route.Method(router, route.Path, route.Handler)
	}
	return router
}
