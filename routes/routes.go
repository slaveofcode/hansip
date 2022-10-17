package routes

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/routes/auth"
	"github.com/slaveofcode/hansip/routes/download"
	"github.com/slaveofcode/hansip/routes/files"
	"github.com/slaveofcode/hansip/routes/middleware"
	"github.com/slaveofcode/hansip/routes/user"
	"github.com/slaveofcode/hansip/routes/visit"
)

func routeAuth(r *gin.RouterGroup, repo repository.Repository) {
	r.POST("/register", auth.Register(repo))
	r.POST("/login", auth.Login(repo))
	r.POST("/refresh-token", auth.RefreshToken(repo))
	r.GET("/check", auth.CheckToken(repo))
}

func routeInternal(r *gin.RouterGroup, repo repository.Repository, s3Client *s3.Client) {
	r.POST("/files/request-group", files.CreateFileGroup(repo))
	r.POST("/files/upload", files.Upload(repo))
	r.POST("/files/bundle-group", files.BundleFileGroup(repo, s3Client))

	r.GET("/users/query", user.UserQueries(repo))
}

func Routes(routes *gin.Engine, repo repository.Repository, s3Client *s3.Client) {
	auth := routes.Group("/auth")
	routeAuth(auth, repo)

	internal := routes.Group("/internal")
	internal.Use(middleware.UserData(repo))
	routeInternal(internal, repo, s3Client)

	routes.GET("/view/:code", visit.View(repo))
	routes.POST("/view/:code", visit.ViewProtected(repo))

	routes.POST("/download/do/:code", download.Download(repo, s3Client))
}
