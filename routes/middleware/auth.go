package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/repository/models"
)

const (
	CTX_USER_ID = "USER_ID"
)

type requestHeader struct {
	Authorization string `header:"Authorization"`
}

func UserData(repo repository.Repository) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		h := requestHeader{}
		if err := ctx.ShouldBindHeader(&h); err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			return
		}

		bearers := strings.Split(h.Authorization, " ")
		if len(bearers) <= 1 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
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
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			return
		}

		if acct.TokenExpiredAt.Before(time.Now()) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			return
		}

		ctx.Set(CTX_USER_ID, strconv.FormatUint(acct.UserId, 10))
	}
}

func GetUserId(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.GetString(CTX_USER_ID), 10, 64)
}
