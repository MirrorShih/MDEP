package router

import (
	"MDEP/controller"
	"MDEP/middleware"
	"github.com/gin-gonic/gin"
)

type Route struct {
	Method  func(engine *gin.RouterGroup, path string, handler func(c *gin.Context))
	Path    string
	Handler func(c *gin.Context)
}

var routes []Route

func GET(engine *gin.RouterGroup, path string, handler func(c *gin.Context)) {
	engine.GET(path, handler)
}

func POST(engine *gin.RouterGroup, path string, handler func(c *gin.Context)) {
	engine.POST(path, handler)
}

func PUT(engine *gin.RouterGroup, path string, handler func(c *gin.Context)) {
	engine.PUT(path, handler)
}

func DELETE(engine *gin.RouterGroup, path string, handler func(c *gin.Context)) {
	engine.DELETE(path, handler)
}

func register(method func(engine *gin.RouterGroup, path string, handler func(c *gin.Context)), path string, handler func(c *gin.Context)) {
	routes = append(routes, Route{method, path, handler})
}

func init() {
	register(GET, "/detector", controller.GetDetectorList)
	register(GET, "/detector/:id", controller.GetDetector)
	register(POST, "/detector", controller.CreateDetector)
	register(POST, "/task", controller.CreateTask)
	register(GET, "/report", controller.GetReportList)
	register(GET, "/report/:id", controller.GetReport)
	register(PUT, "/detector/:id", controller.UpdateDetector)
	register(DELETE, "/detector/:id", controller.DeleteDetector)
	register(GET, "/dataset", controller.GetDatasetList)
	register(GET, "/leaderboard/:dataset", controller.GetLeaderboard)
}

func NewRouter() *gin.Engine {
	router := gin.Default()
	routerGroup := router.Group("/api")
	routerGroup.Use(middleware.GitHubAPIMiddleware)

	for _, route := range routes {
		route.Method(routerGroup, route.Path, route.Handler)
	}
	router.GET("/auth/callback", controller.HandleCallback)
	return router
}
