package routes

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/repository/pg"
	"github.com/slaveofcode/hansip/routes/auth"
	"github.com/slaveofcode/hansip/routes/download"
	"github.com/slaveofcode/hansip/routes/files"
	"github.com/slaveofcode/hansip/routes/middleware"
	"github.com/slaveofcode/hansip/routes/user"
	"github.com/slaveofcode/hansip/routes/visit"
)

func routeAuth(r *gin.RouterGroup, pgRepo *pg.RepositoryPostgres) {
	r.POST("/register", auth.Register(pgRepo))
	r.POST("/login", auth.Login(pgRepo))
	r.POST("/refresh-token", auth.RefreshToken(pgRepo))
	r.GET("/check", auth.CheckToken(pgRepo))
}

func routeInternal(r *gin.RouterGroup, pgRepo *pg.RepositoryPostgres, s3Client *s3.Client) {
	r.POST("/files/request-group", files.CreateFileGroup(pgRepo))
	r.POST("/files/upload", files.Upload(pgRepo))
	r.POST("/files/bundle-group", files.BundleFileGroup(pgRepo, s3Client))

	r.GET("/users/query", user.UserQueries(pgRepo))
}

func Routes(routes *gin.Engine, pgRepo *pg.RepositoryPostgres, s3Client *s3.Client) {
	auth := routes.Group("/auth")
	routeAuth(auth, pgRepo)

	internal := routes.Group("/internal")
	internal.Use(middleware.UserData(pgRepo))
	routeInternal(internal, pgRepo, s3Client)

	routes.GET("/view/:code", visit.View(pgRepo))
	routes.POST("/view/:code", visit.ViewProtected(pgRepo))

	routes.POST("/download/do/:code", download.Download(pgRepo, s3Client))
}
