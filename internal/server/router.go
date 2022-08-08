package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"zeus/internal/handler"
)

func (s *Server) Routes(handlers *handler.Handlers) *gin.Engine {
	if s.config.Debug == false {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:        true,
		AllowMethods:     []string{"*"},
		AllowHeaders:  []string{"*"},
	}))

	pprof.Register(r, "dev/pprof")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	v1 := r.Group("api/v1")
	clusters := v1.Group("clusters")
	nodes := clusters.Group("/:cluster/nodes")
	pods := clusters.Group("/:cluster/pods")
	namespaces := clusters.Group("/:cluster/namespaces")
	{
		h := handlers.Cluster
		clusters.GET("", h.List)
		clusters.GET("/:cluster", h.Get)
	}
	{
		h := handlers.Node
		nodes.GET("", h.List)
		nodes.GET("/:node", h.Get)
	}
	{
		h := handlers.Pod
		pods.GET("", h.List)
		namespaces.GET("/:namespace/pods/:pod", h.Get)
		namespaces.GET("/:namespace/pods/:pod/log", h.Log)
		namespaces.GET("/:namespace/pods/:pod/exec", h.Exec)
	}
	{
		h := handlers.Namespace
		namespaces.GET("", h.List)
		namespaces.GET("/:namespace", h.Get)
	}

	return r
}
