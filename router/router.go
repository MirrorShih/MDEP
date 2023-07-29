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

func PUT(engine *gin.Engine, path string, handler func(c *gin.Context)) {
	engine.PUT(path, handler)
}

func DELETE(engine *gin.Engine, path string, handler func(c *gin.Context)) {
	engine.DELETE(path, handler)
}

func register(method func(engine *gin.Engine, path string, handler func(c *gin.Context)), path string, handler func(c *gin.Context)) {
	routes = append(routes, Route{method, path, handler})
}

func init() {
	register(GET, "/api/detector", controller.GetDetectorList)
	register(POST, "/api/detector", controller.CreateDetector)
	register(POST, "/api/task", controller.CreateTask)
	register(GET, "/api/report", controller.GetReportList)
	register(GET, "/api/report/:id", controller.GetReport)
	register(PUT, "/api/detector/:id", controller.UpdateDetector)
	register(DELETE, "/api/detector/:id", controller.DeleteDetector)
}

func NewRouter() *gin.Engine {
	router := gin.Default()
	for _, route := range routes {
		route.Method(router, route.Path, route.Handler)
	}
	return router
}
