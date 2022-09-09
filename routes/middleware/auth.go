package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/slaveofcode/hansip/repository/pg"
	"github.com/slaveofcode/hansip/repository/pg/models"
)

const (
	CTX_USER_ID = "USER_ID"
)

type requestHeader struct {
	Authorization string `header:"Authorization"`
}

func UserData(pgRepo *pg.RepositoryPostgres) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		h := requestHeader{}
		if err := ctx.ShouldBindHeader(&h); err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
		}

		bearers := strings.Split(h.Authorization, " ")
		if len(bearers) <= 1 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
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
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
		}

		if acct.TokenExpiredAt.Before(time.Now()) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
		}

		ctx.Set(CTX_USER_ID, acct.UserId.String())
	}
}

func GetUserId(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetString(CTX_USER_ID))
}
