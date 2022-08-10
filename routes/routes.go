package routes

import (
	"github.com/gin-gonic/gin"
)

func Routes(routes *gin.Engine) {
	routes.POST("/auth/login", Login)
	routes.POST("/upload", Upload)
}
