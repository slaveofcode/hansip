package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/securi/repository/pg"
	"github.com/slaveofcode/securi/routes/auth"
	"github.com/slaveofcode/securi/routes/files"
	"github.com/slaveofcode/securi/routes/middleware"
)

func routeAuth(r *gin.RouterGroup, pgRepo *pg.RepositoryPostgres) {
	r.POST("/register", auth.Register(pgRepo))
	r.POST("/login", auth.Login(pgRepo))
	r.POST("/refresh-token", auth.RefreshToken(pgRepo))
	r.GET("/check", auth.CheckToken(pgRepo))
}

func routeInternal(r *gin.RouterGroup, pgRepo *pg.RepositoryPostgres) {
	r.POST("/files/request-group", files.CreateFileGroup(pgRepo))
	r.POST("/files/upload", files.Upload(pgRepo))
	r.POST("/files/compile-group", files.CreateFileGroup(pgRepo))
}

func Routes(routes *gin.Engine, pgRepo *pg.RepositoryPostgres) {
	auth := routes.Group("/auth")
	routeAuth(auth, pgRepo)

	internal := routes.Group("/internal")
	internal.Use(middleware.UserData(pgRepo))
	routeInternal(internal, pgRepo)

	routes.GET("/v/:code", Visit)
	routes.POST("/d/:code", Download)
}
