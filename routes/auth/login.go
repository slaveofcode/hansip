package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/repository/models"
	"github.com/slaveofcode/hansip/utils/token"
	"golang.org/x/crypto/bcrypt"
)

type LoginParam struct {
	Email    string `json:"email" binding:"required,lowercase,email"`
	Password string `json:"password" binding:"required"`
}

func Login(repo repository.Repository) func(c *gin.Context) {
	return func(c *gin.Context) {
		var bodyParams LoginParam
		if err := c.ShouldBindJSON(&bodyParams); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid body request",
			})
			return
		}

		db := repo.GetDB()

		var userCred models.UserCredential
		res := db.Where(&models.UserCredential{
			IdentityType:  models.IdentityTypeEmail,
			IdentityValue: bodyParams.Email,
		}).First(&userCred)

		if res.RowsAffected == 0 {
			// unimplemented credential, throw
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid credential",
			})
			return
		}

		if userCred.CredentialType == models.CredentialTypePassword {
			err := bcrypt.CompareHashAndPassword([]byte(userCred.CredentialValue), []byte(bodyParams.Password))
			if err != nil {
				// password doesn't match
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid credential",
				})
				return
			}
		}

		// another credential type here...

		tokenInfo, err := token.GetFreshTokens(db)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Unable to process login",
			})
			return
		}

		acct := models.AccessToken{
			UserId:                userCred.UserId,
			Token:                 tokenInfo.AccessToken,
			RefreshToken:          tokenInfo.RefreshToken,
			TokenExpiredAt:        time.Now().Add(time.Hour),          // 1 hour
			RefreshTokenExpiredAt: time.Now().Add(time.Hour * 24 * 7), // 7 days
		}

		res = db.Create(&acct)

		if res.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Unable to process login",
			})
			return
		}

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
