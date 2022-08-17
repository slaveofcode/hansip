package visit

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/securi/repository/pg"
)

func Download(pgRepo *pg.RepositoryPostgres) func(c *gin.Context) {
	return func(c *gin.Context) {
		body := map[string]interface{}{}
		c.BindJSON(&body)
		log.Println(body["passcode"])
		c.JSON(http.StatusCreated, gin.H{
			"status":  true,
			"message": "Logged in!",
		})
	}
}
