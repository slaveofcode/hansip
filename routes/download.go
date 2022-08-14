package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Download(c *gin.Context) {
	body := map[string]interface{}{}
	c.BindJSON(&body)
	log.Println(body["passcode"])
	c.JSON(http.StatusCreated, gin.H{
		"status":  true,
		"message": "Logged in!",
	})
}
