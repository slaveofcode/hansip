package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/repository/models"
)

type requestHeader struct {
	Authorization string `header:"Authorization"`
}

func CheckToken(repo repository.Repository) func(c *gin.Context) {
	return func(c *gin.Context) {
		h := requestHeader{}
		if err := c.ShouldBindHeader(&h); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			return
		}

		bearers := strings.Split(h.Authorization, " ")
		if len(bearers) <= 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			return
		}

		bearer := bearers[1]
		db := repo.GetDB()

		var acct models.AccessToken
		res := db.Where(&models.AccessToken{
			Token: bearer,
		}).First(&acct)

		if res.RowsAffected <= 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			return
		}

		if acct.TokenExpiredAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	}
}
