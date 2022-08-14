package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/securi/repository/pg"
	"github.com/slaveofcode/securi/repository/pg/models"
)

type requestHeader struct {
	Authorization string `header:"Authorization"`
}

func CheckToken(pgRepo *pg.RepositoryPostgres) func(c *gin.Context) {
	return func(c *gin.Context) {
		h := requestHeader{}
		if err := c.ShouldBindHeader(&h); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Unauthorized",
			})
		}

		bearers := strings.Split(h.Authorization, " ")
		if len(bearers) <= 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Unauthorized",
			})
		}

		bearer := bearers[1]
		db := pgRepo.GetDB()

		var acct models.AccessToken
		res := db.Where(&models.AccessToken{
			Token: bearer,
		}).First(&acct)

		if res.RowsAffected <= 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Unauthorized",
			})
		}

		if acct.TokenExpiredAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Unauthorized",
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"status": true,
		})
	}
}
