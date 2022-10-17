package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/repository/models"
	"github.com/slaveofcode/hansip/utils/token"
)

type RefreshTokenParam struct {
	Token string `json:"token" binding:"required"`
}

func RefreshToken(repo repository.Repository) func(c *gin.Context) {
	return func(c *gin.Context) {
		var bodyParams RefreshTokenParam
		if err := c.ShouldBindJSON(&bodyParams); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid body request",
			})
			return
		}

		// create user
		db := repo.GetDB()

		var extAccToken models.AccessToken
		res := db.Where(&models.AccessToken{
			RefreshToken: bodyParams.Token,
		}).First(&extAccToken)

		if res.RowsAffected == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Unable to refresh token",
			})
			return
		}

		if extAccToken.RefreshTokenExpiredAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Expired refresh token",
			})
			return
		}

		tokenInfo, err := token.GetFreshTokens(db)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unable to process refresh token",
			})
			return
		}

		acct := models.AccessToken{
			UserId:                extAccToken.UserId,
			Token:                 tokenInfo.AccessToken,
			RefreshToken:          tokenInfo.RefreshToken,
			TokenExpiredAt:        time.Now().Add(time.Hour),          // 1 hour
			RefreshTokenExpiredAt: time.Now().Add(time.Hour * 24 * 7), // 7 days
		}

		res = db.Create(&acct)
		if res.Error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unable to process refresh token",
			})
			return
		}

		go func() {
			res := db.Unscoped().Delete(&extAccToken)
			if res.Error != nil {
				log.Println("Error delete expired access token:", err)
			}
		}()

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"accessToken":  acct.Token,
				"refreshToken": acct.RefreshToken,
				"exp":          acct.TokenExpiredAt.Format(time.RFC3339),
				"expRefresh":   acct.RefreshTokenExpiredAt.Format(time.RFC3339),
			},
		})
	}
}
