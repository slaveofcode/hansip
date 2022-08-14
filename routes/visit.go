package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Visit(c *gin.Context) {
	body := map[string]interface{}{}
	c.BindJSON(&body)
	log.Println(body["email"])
	log.Println(body["password"])
	c.JSON(http.StatusCreated, gin.H{
		"status":  true,
		"message": "Logged in!",
	})
}
